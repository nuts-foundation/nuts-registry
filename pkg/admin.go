package pkg

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"github.com/lestrrat-go/jwx/jwk"
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

// RegisterVendor registers a vendor
func (r *Registry) RegisterVendor(id string, name string, domain string) (events.Event, error) {
	r.logger().Infof("Registering vendor, id=%s, name=%s, domain=%s", id, name, domain)
	entity := types.LegalEntity{URI: id}
	err := r.crypto.GenerateKeyPairFor(entity)
	if err != nil {
		return nil, err
	}

	certASN1, err := cert.IssueCertificate(r.crypto, func() (x509.CertificateRequest, error) {
		// TODO: Make env configurable
		return cert.VendorCACertificateRequest(id, name, domain, "")
	}, entity, entity, crypto.CertificateProfile{
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		NumDaysValid: vendorCACertificateDaysValid,
	})
	if err != nil {
		return nil, err
	}

	jwkAsMap, err := marshalJwk(certASN1, cert.VendorCACertificate)
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

func marshalJwk(certASN1 []byte, certType cert.CertificateType) (map[string]interface{}, error) {
	certificate, err := x509.ParseCertificate(certASN1)
	if err != nil {
		return nil, err
	}
	jsonWebKey, err := jwk.New(certificate.PublicKey)
	if err != nil {
		return nil, err
	}
	jwkAsMap, err := crypto.JwkToMap(jsonWebKey)
	if err != nil {
		return nil, err
	}
	jwkAsMap[jwk.X509CertChainKey] = base64.StdEncoding.EncodeToString(certASN1)
	jwkAsMap[cert.JwkCertificateType] = certType
	return jwkAsMap, nil
}

// VendorClaim registers an organization under a vendor. orgKeys are the organization's keys in JWK format
func (r *Registry) VendorClaim(vendorID string, orgID string, orgName string, orgKeys []interface{}) (events.Event, error) {
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))

	if orgKeys == nil || len(orgKeys) == 0 {
		logrus.Infof("No keys specified for organisation (id = %s). Keys will be loaded from crypto module.", orgID)
		orgKey, err := r.loadOrGenerateKey(orgID)
		if err != nil {
			return nil, err
		}
		orgKeys = append(orgKeys, orgKey)
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
func (r *Registry) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, version string) (events.Event, error) {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s, version=%s",
		organizationID, id, endpointType, url, status, version)
	return r.publishEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
		Organization: events.Identifier(organizationID),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   events.Identifier(id),
		Status:       status,
		Version:      version,
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
