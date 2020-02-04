package pkg

import (
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/google/uuid"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	"github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
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
func (r *Registry) RegisterVendor(id string, name string, domain string) (events.Event, error) {
	r.logger().Infof("Registering vendor, id=%s, name=%s, domain=%s", id, name, domain)
	entity := types.LegalEntity{URI: id}
	err := r.crypto.GenerateKeyPairFor(entity)
	if err != nil {
		return nil, err
	}

	certificate, err := r.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
		// TODO: Make env configurable
		return cert.VendorCACertificateRequest(id, name, domain, "")
	}, entity, entity, crypto.CertificateProfile{
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		NumDaysValid: vendorCACertificateDaysValid,
	})
	if err != nil {
		return nil, err
	}

	jwkAsMap, err := certToJwkMap(certificate, cert.VendorCACertificate)
	if err != nil {
		return nil, errors2.Wrap(err, ErrJWKConstruction.Error())
	}

	event := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
		Identifier: events.Identifier(id),
		Name:       name,
		Domain:     domain,
		Keys:       []interface{}{jwkAsMap},
	})
	err = r.EventSystem.PublishEvent(event)
	if err != nil {
		return nil, err
	}
	return event, err
}

// VendorClaim registers an organization under a vendor. The specified vendor has to exist and have a valid CA certificate
// as to issue the organisation certificate. If specified orgKeys are interpreted as the organization's keys in JWK format.
// If not specified, a new key pair is generated.
func (r *Registry) VendorClaim(vendorID string, orgID string, orgName string, orgKeys []interface{}) (events.Event, error) {
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))

	vendor := r.Db.VendorByID(vendorID)
	if vendor == nil {
		return nil, fmt.Errorf("vendor not found (id=%s)", vendorID)
	}
	certificates := vendor.GetActiveCertificates()
	if len(certificates) == 0 {
		return nil, fmt.Errorf("vendor has no active certificates (id = %s)", vendorID)
	}

	// If no keys are supplied, make there's a key in the crypto module for the organisation
	if orgKeys == nil || len(orgKeys) == 0 {
		logrus.Infof("No keys specified for organisation (id = %s). Keys will be loaded from crypto module.", orgID)
		_, err := r.loadOrGenerateKey(orgID)
		if err != nil {
			return nil, err
		}
	}

	certificate, err := r.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
		// TODO: Make env configurable
		return cert.OrganisationCertificateRequest(vendor.Name, orgID, orgName, vendor.Domain, "")
	}, types.LegalEntity{URI: orgID}, types.LegalEntity{URI: vendorID}, crypto.CertificateProfile{
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

	return r.publishEvent(events.VendorClaim, events.VendorClaimEvent{
		VendorIdentifier: events.Identifier(vendorID),
		OrgIdentifier:    events.Identifier(orgID),
		OrgName:          orgName,
		OrgKeys:          orgKeys,
		Start:            time.Now(),
	})
}

// RegisterEndpoint registers an endpoint for an organization
func (r *Registry) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error) {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s",
		organizationID, id, endpointType, url, status)
	if id == "" {
		id = uuid.New().String()
	}
	return r.publishEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
		Organization: events.Identifier(organizationID),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   events.Identifier(id),
		Status:       status,
		Properties:   properties,
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

func (r *Registry) publishEvent(eventType events.EventType, payload interface{}) (events.Event, error) {
	event := events.CreateEvent(eventType, payload)
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
