package pkg

import (
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	types2 "github.com/nuts-foundation/nuts-registry/pkg/types"
	"time"

	"github.com/google/uuid"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	certutil "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	dom "github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrJWKConstruction indicates that a JWK couldn't be constructed
var ErrJWKConstruction = errors.New("unable to construct JWK")

// ErrCertificateIssue indicates a certificate couldn't be issued
var ErrCertificateIssue = errors.New("unable to issue certificate")

// ErrOrganizationNotFound is returned when the specified organization was not found
var ErrOrganizationNotFound = errors.New("organization not found")

// RegisterVendor registers a vendor
func (r *Registry) RegisterVendor(certificate *x509.Certificate) (events.Event, error) {
	id := core.NutsConfig().VendorID()
	// Find out whether this is a registration or update operation
	previousEvent, err := r.EventSystem.FindLastEvent(dom.VendorEventMatcher(id))
	if err != nil {
		return nil, err
	}
	if previousEvent == nil {
		r.logger().Infof("Registering vendor: %s", certificate.Subject)
	} else {
		r.logger().Infof("Updating vendor: %s", certificate.Subject)
	}
	name := certificate.Subject.Organization[0]
	domain, err := certutil.NewNutsCertificate(certificate).GetDomain()
	if err != nil {
		return nil, errors2.Wrap(err, "unable to get domain from vendor certificate")
	}
	certificateAsJWK, _ := cert.CertificateToJWK(certificate)
	certificateAsMap, _ := cert.JwkToMap(certificateAsJWK)
	if err := r.crypto.StoreVendorCACertificate(certificate); err != nil {
		return nil, err
	}
	// The event is signed with the vendor certificate, which need to be issued by the just issued vendor CA.
	entity := types.LegalEntity{URI: id.String()}
	if _, _, err := r.crypto.RenewSigningCertificate(entity); err != nil {
		return nil, err
	}
	return r.signAndPublishEvent(dom.RegisterVendor, dom.RegisterVendorEvent{
		Identifier: id,
		Name:       name,
		Domain:     domain,
		Keys:       []interface{}{certificateAsMap},
	}, previousEvent, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.crypto.SignJWS(dataToBeSigned, types.KeyForEntity(entity).WithQualifier(crypto.SigningCertificateQualifier))
	})
}

// VendorClaim registers an organization under a vendor. The specified vendor has to exist and have a valid CA certificate
// as to issue the organisation certificate. If specified orgKeys are interpreted as the organization's keys in JWK format.
// If not specified, a new key pair is generated.
func (r *Registry) VendorClaim(orgID core.PartyID, orgName string, orgKeys []interface{}) (events.Event, error) {
	vendorID := core.NutsConfig().VendorID()
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))

	vendor, err := r.getVendor()
	if err != nil {
		return nil, err
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
		jwkAsMap, err := r.issueOrganizationCertificate(vendor, orgID, orgName)
		if err != nil {
			return nil, err
		}
		orgKeys = append(orgKeys, jwkAsMap)
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
		VendorID:       vendorID,
		OrganizationID: orgID,
		OrgName:        orgName,
		OrgKeys:        orgKeys,
		Start:          time.Now(),
	}, nil, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsOrganization(orgID, orgName, dataToBeSigned, instant, orgHasCerts)
	})
}

func (r *Registry) issueOrganizationCertificate(vendor *db.Vendor, orgID core.PartyID, orgName string) (map[string]interface{}, error) {
	_, err := r.loadOrGenerateKey(orgID)
	if err != nil {
		return nil, err
	}
	certificate, err := r.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
		return certutil.OrganisationCertificateRequest(vendor.Name, orgID, orgName, vendor.Domain)
	}, types.LegalEntity{URI: orgID.String()}, types.LegalEntity{URI: vendor.Identifier.String()}, crypto.CertificateProfile{
		IsCA:         true,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		NumDaysValid: r.Config.OrganisationCertificateValidity,
	})
	if err != nil {
		return nil, errors2.Wrap(err, ErrCertificateIssue.Error())
	}

	key, _ := cert.CertificateToJWK(certificate)
	return cert.JwkToMap(key)
}

func (r *Registry) RefreshOrganizationCertificate(organizationID core.PartyID) (events.Event, error) {
	logrus.Infof("Issuing new certificate for organization using existing private key (if present) (id=%s)", organizationID)
	vendor, err := r.getVendor()
	if err != nil {
		return nil, err
	}
	// This operation can only be used to issue a new certificate for an existing organization. The resulting event refers
	// to the last VendorClaimEvent.
	prevEvent, err := r.EventSystem.FindLastEvent(dom.OrganizationEventMatcher(vendor.Identifier, organizationID))
	if err != nil {
		return nil, err
	}
	if prevEvent == nil {
		return nil, ErrOrganizationNotFound
	}
	var prevEventPayload = dom.VendorClaimEvent{}
	_ = prevEvent.Unmarshal(&prevEventPayload)
	// Issue certificate, apply as update to the last event and emit
	jwkAsMap, err := r.issueOrganizationCertificate(vendor, prevEventPayload.OrganizationID, prevEventPayload.OrgName)
	if err != nil {
		return nil, err
	}
	prevEventPayload.OrgKeys = append(prevEventPayload.OrgKeys, jwkAsMap)
	return r.signAndPublishEvent(dom.VendorClaim, prevEventPayload, prevEvent, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsOrganization(organizationID, prevEventPayload.OrgName, dataToBeSigned, instant, true)
	})
}

// RegisterEndpoint registers an endpoint for an organization
func (r *Registry) RegisterEndpoint(organizationID core.PartyID, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error) {
	logrus.Infof("Registering/updating endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s",
		organizationID, id, endpointType, url, status)
	if id == "" {
		id = uuid.New().String()
	}
	org, err := r.Db.OrganizationById(organizationID)
	if err != nil {
		return nil, err
	}
	// Find out if this should be an update. That's the case if there's a RegisterEndpointEvent for the same organization
	// and endpoint (ID).
	parentEvent, err := r.EventSystem.FindLastEvent(func(event events.Event) bool {
		if event.Type() != dom.RegisterEndpoint {
			return false
		}
		var payload = dom.RegisterEndpointEvent{}
		_ = event.Unmarshal(&payload)
		return types2.EndpointID(id) == payload.Identifier && organizationID == payload.Organization
	})
	if err != nil {
		return nil, err
	}
	return r.signAndPublishEvent(dom.RegisterEndpoint, dom.RegisterEndpointEvent{
		Organization: organizationID,
		URL:          url,
		EndpointType: endpointType,
		Identifier:   types2.EndpointID(id),
		Status:       status,
		Properties:   properties,
	}, parentEvent, func(dataToBeSigned []byte, instant time.Time) ([]byte, error) {
		return r.signAsOrganization(org.Identifier, org.Name, dataToBeSigned, instant, len(org.GetActiveCertificates()) > 0)
	})
}

func (r *Registry) loadOrGenerateKey(party core.PartyID) (map[string]interface{}, error) {
	key := types.KeyForEntity(types.LegalEntity{URI: party.String()})
	if !r.crypto.PrivateKeyExists(key) {
		logrus.Infof("No keys found for entity (%s), will generate a new key pair.", party)
		if _, err := r.crypto.GenerateKeyPair(key, false); err != nil {
			return nil, err
		}
	}
	keyAsJwk, err := r.crypto.GetPublicKeyAsJWK(key)
	if err != nil {
		return nil, err
	}
	return cert.JwkToMap(keyAsJwk)
}

func (r *Registry) signAndPublishEvent(eventType events.EventType, payload interface{}, previousEvent events.Event, signer func([]byte, time.Time) ([]byte, error)) (events.Event, error) {
	var previousEventRef events.Ref
	if previousEvent != nil {
		previousEventRef = previousEvent.Ref()
	}
	event := events.CreateEvent(eventType, payload, previousEventRef)
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
	subjectKey := types.KeyForEntity(subject)
	caKey := types.KeyForEntity(ca)
	subjectPrivKey, err := r.crypto.GetPrivateKey(subjectKey)
	if err != nil {
		return nil, errors2.Wrapf(err, "unable to retrieve subject private key: %s", subject)
	}

	csrTemplate.PublicKey = subjectPrivKey.Public()
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, subjectPrivKey)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR")
	}

	certASN1, err := r.crypto.SignCertificate(subjectKey, caKey, csr, profile)
	if err != nil {
		return nil, errors2.Wrap(err, "error while signing certificate")
	}

	return x509.ParseCertificate(certASN1)
}

func (r *Registry) signAsOrganization(orgID core.PartyID, orgName string, payload []byte, instant time.Time, hasCerts bool) ([]byte, error) {
	vendor, err := r.getVendor()
	if err != nil {
		return nil, err
	}
	var signature []byte
	// https://github.com/nuts-foundation/nuts-registry/issues/84
	// The check below is for backwards compatibility when the organization or vendor creating the organization has no
	// certificates, so we can't sign
	//the event. This should be removed when event signing is mandatory, when
	// all vendors and organizations have certificates.
	// Or maybe this check should be changed (by then) to let it return an error since the vendor
	// should first make sure to have an active certificate.
	if hasCerts {
		logrus.Debug("Signing event as organization")
		csr, err := certutil.OrganisationCertificateRequest(vendor.Name, orgID, orgName, vendor.Domain)
		if err != nil {
			return nil, errors2.Wrap(err, "unable to create CSR for JWS signing")
		}
		signature, err = r.crypto.SignJWSEphemeral(payload, types.KeyForEntity(types.LegalEntity{URI: orgID.String()}), csr, instant)
		if err != nil {
			return nil, errors2.Wrap(err, "unable to sign JWS")
		}
	}
	return signature, nil
}

func (r *Registry) getVendor() (*db.Vendor, error) {
	id := core.NutsConfig().VendorID()
	vendor := r.Db.VendorByID(id)
	if vendor == nil {
		return nil, fmt.Errorf("vendor not found (id=%s)", id)
	}
	return vendor, nil
}
