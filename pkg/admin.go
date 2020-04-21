package pkg

import (
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/google/uuid"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	dom "github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const vendorCACertificateDaysValid = 365

// ErrJWKConstruction indicates that a JWK couldn't be constructed
var ErrJWKConstruction = errors.New("unable to construct JWK")

// ErrCertificateIssue indicates a certificate couldn't be issued
var ErrCertificateIssue = errors.New("unable to issue certificate")

// RegisterVendor registers a vendor
func (r *Registry) RegisterVendor(name string, domain string) (events.Event, error) {
	id := core.NutsConfig().Identity()
	r.logger().Infof("Registering vendor, id=%s, name=%s, domain=%s", id, name, domain)
	entity := types.LegalEntity{URI: id}
	err := r.crypto.GenerateKeyPairFor(entity)
	if err != nil {
		return nil, err
	}

	certificate, err := r.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
		return cert.VendorCertificateRequest(id, name, "CA Intermediate", domain)
	}, entity, entity, crypto.CertificateProfile{
		IsCA:         true,
		NumDaysValid: vendorCACertificateDaysValid,
	})
	if err != nil {
		return nil, err
	}

	jwkAsMap, err := certToJwkMap(certificate, cert.VendorCACertificate)
	if err != nil {
		return nil, errors2.Wrap(err, ErrJWKConstruction.Error())
	}

	// The event is signed with the vendor certificate, which is issued by the just issued vendor CA.
	event, err := r.signAndPublishEvent(dom.RegisterVendor, dom.RegisterVendorEvent{
		Identifier: dom.Identifier(id),
		Name:       name,
		Domain:     domain,
		Keys:       []interface{}{jwkAsMap},
	}, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsVendor(id, name, domain, dataToBeSigned, instant)
	})
	if err == nil && r.vendor.Identifier == "" {
		// This node isn't configured with a vendor yet but we just registered it, so make it our current vendor.
		r.vendor = db.Vendor{
			Identifier: db.Identifier(id),
			Name:       name,
			Domain:     domain,
			Keys:       []interface{}{jwkAsMap},
		}
	}
	return event, err
}

// VendorClaim registers an organization under a vendor. The specified vendor has to exist and have a valid CA certificate
// as to issue the organisation certificate. If specified orgKeys are interpreted as the organization's keys in JWK format.
// If not specified, a new key pair is generated.
func (r *Registry) VendorClaim(orgID string, orgName string, orgKeys []interface{}) (events.Event, error) {
	vendorID := core.NutsConfig().Identity()
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))

	vendor := r.Db.VendorByID(vendorID)
	if vendor == nil {
		return nil, fmt.Errorf("vendor not found (id=%s)", vendorID)
	}

	// If no keys are supplied, make sure there's a key in the crypto module for the organisation
	if len(orgKeys) == 0 {
		logrus.Infof("No keys specified for organisation (id=%s). Keys will be generated or loaded from crypto module.", orgID)
		_, err := r.loadOrGenerateKey(orgID)
		if err != nil {
			return nil, err
		}
	}

	var orgHasCerts bool
	if len(vendor.GetActiveCertificates()) > 0 {
		// If the vendor has certificates, it means it has (should have) a CA certificate which can issue a certificate to the new org
		certificate, err := r.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
			return cert.OrganisationCertificateRequest(vendor.Name, orgID, orgName, vendor.Domain)
		}, types.LegalEntity{URI: orgID}, types.LegalEntity{URI: vendorID}, crypto.CertificateProfile{
			IsCA:         true,
			KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
			NumDaysValid: r.Config.OrganisationCertificateValidity,
		})
		if err != nil {
			return nil, errors2.Wrap(err, ErrCertificateIssue.Error())
		}

		jwkAsMap, err := certToJwkMap(certificate, cert.OrganisationCertificate)
		orgKeys = append(orgKeys, jwkAsMap)
		if err != nil {
			return nil, errors2.Wrap(err, ErrJWKConstruction.Error())
		}
		orgHasCerts = true
	} else {
		// https://github.com/nuts-foundation/nuts-registry/issues/84
		// Vendor has no certificates, we just make sure the org has a plain JWK without X.509 certificate, either
		// provided or freshly generated. This else-branch should be removed when signing events is mandatory!
		if len(orgKeys) == 0 {
			orgKey, err := r.loadOrGenerateKey(orgID)
			if err != nil {
				return nil, err
			}
			orgKeys = append(orgKeys, orgKey)
		}
		orgHasCerts = false
	}

	return r.signAndPublishEvent(dom.VendorClaim, dom.VendorClaimEvent{
		VendorIdentifier: dom.Identifier(vendorID),
		OrgIdentifier:    dom.Identifier(orgID),
		OrgName:          orgName,
		OrgKeys:          orgKeys,
		Start:            time.Now(),
	}, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsOrganization(orgID, orgName, dataToBeSigned, instant, orgHasCerts)
	})
}

// RegisterEndpoint registers an endpoint for an organization
func (r *Registry) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error) {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s",
		organizationID, id, endpointType, url, status)
	if id == "" {
		id = uuid.New().String()
	}
	org, err := r.Db.OrganizationById(organizationID)
	if err != nil {
		return nil, err
	}
	return r.signAndPublishEvent(dom.RegisterEndpoint, dom.RegisterEndpointEvent{
		Organization: dom.Identifier(organizationID),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   dom.Identifier(id),
		Status:       status,
		Properties:   properties,
	}, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsOrganization(org.Identifier.String(), org.Name, dataToBeSigned, instant, len(org.GetActiveCertificates()) > 0)
	})
}

func (r *Registry) loadOrGenerateKey(identifier string) (map[string]interface{}, error) {
	entity := types.LegalEntity{URI: identifier}
	if !r.crypto.KeyExistsFor(entity) {
		logrus.Infof("No keys found for entity (id = %s), will generate a new key pair.", identifier)
		if err := r.crypto.GenerateKeyPairFor(entity); err != nil {
			return nil, err
		}
	}
	keyAsJwk, err := r.crypto.PublicKeyInJWK(entity)
	if err != nil {
		return nil, err
	}
	return crypto.JwkToMap(keyAsJwk)
}

func (r *Registry) signAndPublishEvent(eventType events.EventType, payload interface{}, signer func([]byte, time.Time) ([]byte, error)) (events.Event, error) {
	event := events.CreateEvent(eventType, payload)
	if signer != nil {
		err := event.Sign(func(data []byte) ([]byte, error) {
			return signer(data, event.IssuedAt())
		})
		if err != nil {
			return nil, errors2.Wrap(err, "unable to sign event")
		}
	}
	if err := r.EventSystem.PublishEvent(event); err != nil {
		return nil, err
	}
	return event, nil
}

// createAndSubmitCSR issues an X.509 certificate to entity 'subject' through Certificate Authority 'ca'. It assumes
// the CA is under control of the application since it expects the crypto module to directly issue the certificate.
// Both the subject's and CA's key pair should be available in the crypto module. If subject and CA are equal,
// it issues a self-signed certificate. Otherwise, the CA's certificate should also be present in the crypto module.
func (r *Registry) createAndSubmitCSR(csrTemplateFn func() (x509.CertificateRequest, error),
	subject types.LegalEntity, ca types.LegalEntity, profile crypto.CertificateProfile) (*x509.Certificate, error) {
	csrTemplate, err := csrTemplateFn()
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR template")
	}

	subjectPrivKey, err := r.crypto.GetOpaquePrivateKey(subject)
	if err != nil {
		return nil, errors2.Wrapf(err, "unable to retrieve subject private key: %s", subject)
	}

	csrTemplate.PublicKey = subjectPrivKey.Public()
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, subjectPrivKey)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR")
	}

	certASN1, err := r.crypto.SignCertificate(subject, ca, csr, profile)
	if err != nil {
		return nil, errors2.Wrap(err, "error while signing certificate")
	}

	return x509.ParseCertificate(certASN1)
}

func certToJwkMap(certificate *x509.Certificate, certType cert.CertificateType) (map[string]interface{}, error) {
	key, _ := crypto.CertificateToJWK(certificate)
	keyAsMap, _ := crypto.JwkToMap(key)
	keyAsMap[cert.JwkCertificateType] = certType
	return keyAsMap, nil
}

func (r *Registry) signAsVendor(vendorId string, vendorName string, domain string, payload []byte, instant time.Time) ([]byte, error) {
	csr, err := cert.VendorCertificateRequest(vendorId, vendorName, "", domain)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR for JWS signing")
	}
	signature, err := r.crypto.JWSSignEphemeral(payload, types.LegalEntity{URI: vendorId}, csr, instant)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to sign JWS")
	}
	return signature, nil
}

func (r *Registry) signAsOrganization(orgID string, orgName string, payload []byte, instant time.Time, hasCerts bool) ([]byte, error) {
	var signature []byte
	// https://github.com/nuts-foundation/nuts-registry/issues/84
	// The check below is for backwards compatibility when the organization or vendor creating the organization has no
	// certificates, so we can't sign
	//the event. This should be removed when event signing is mandatory, when
	// all vendors and organizations have certificates.
	// Or maybe this check should be changed (by then) to let it return an error since the vendor
	// should first make sure to have an active certificate.
	if hasCerts {
		csr, err := cert.OrganisationCertificateRequest(r.vendor.Name, orgID, orgName, r.vendor.Domain)
		if err != nil {
			return nil, errors2.Wrap(err, "unable to create CSR for JWS signing")
		}
		signature, err = r.crypto.JWSSignEphemeral(payload, types.LegalEntity{URI: orgID}, csr, instant)
		if err != nil {
			return nil, errors2.Wrap(err, "unable to sign JWS")
		}
	}
	return signature, nil
}
