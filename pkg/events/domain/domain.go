package domain

import (
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/log"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// RegisterEndpoint event type
	RegisterEndpoint events.EventType = "RegisterEndpointEvent"
	// RegisterVendor event type
	RegisterVendor events.EventType = "RegisterVendorEvent"
	// VendorClaim event type
	VendorClaim events.EventType = "VendorClaimEvent"
)

const (
	// HealthcareDomain is a const for domain 'healthcare'
	HealthcareDomain string = "healthcare"
	// PersonalDomain is a const for domain 'personal' (which are "PGO's")
	PersonalDomain = "personal"
	// InsuranceDomain is a const for domain 'insurance'
	InsuranceDomain = "insurance"
	// FallbackDomain is a const for the fallback domain in case there's no domain set, which can be the case for legacy data.
	FallbackDomain = HealthcareDomain
)

// Identifier defines component schema for Identifier.
type Identifier string

func (i Identifier) String() string {
	return string(i)
}

// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Organization Identifier        `json:"organization"`
	URL          string            `json:"URL"`
	EndpointType string            `json:"endpointType"`
	Identifier   Identifier        `json:"identifier"`
	Status       string            `json:"status"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// RegisterVendorEvent event
type RegisterVendorEvent struct {
	Identifier Identifier    `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

func (r *RegisterVendorEvent) PostProcessUnmarshal(event events.Event) error {
	// Default fallback to 'healthcare' domain when none is set, for handling legacy data when 'domain' didn't exist.
	if r.Domain == "" {
		r.Domain = FallbackDomain
	}
	if err := cert.ValidateJWK(r.Keys...); err != nil {
		return err
	}
	for _, key := range r.Keys {
		var vendorID = ""
		if certificate := cert.GetCertificate(key); certificate != nil {
			var err error
			if vendorID, err = cert2.GetVendorSubjectAltName(certificate); err != nil {
				return errors2.Wrap(err, "unable to unmarshal vendor ID from certificate")
			}
		}
		if string(r.Identifier) != vendorID {
			return fmt.Errorf("vendor ID in certificate (%s) doesn't match event (%s)", vendorID, r.Identifier)
		}
	}
	return nil
}

// VendorEventMatcher returns an EventMatcher which matches the RegisterVendorEvent for the vendor with the specified ID.
func VendorEventMatcher(vendorID string) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != RegisterVendor {
			return false
		}
		var p = RegisterVendorEvent{}
		_ = event.Unmarshal(&p)
		return Identifier(vendorID) == p.Identifier
	}
}

// OrganizationEventMatcher returns an EventMatcher which matches the VendorClaimEvent which registered the organization
// with the specified ID for the specified vendor (also by ID).
func OrganizationEventMatcher(vendorID string, organizationID string) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != VendorClaim {
			return false
		}
		var payload = VendorClaimEvent{}
		_ = event.Unmarshal(&payload)
		return Identifier(vendorID) == payload.VendorIdentifier && Identifier(organizationID) == payload.OrgIdentifier
	}
}

// VendorClaimEvent event
type VendorClaimEvent struct {
	VendorIdentifier Identifier `json:"vendorIdentifier"`
	OrgIdentifier    Identifier `json:"orgIdentifier"`
	OrgName          string     `json:"orgName"`
	// OrgKeys is a list of JWKs which are used to
	// 1. encrypt data to be decrypted by the organization,
	// 2. sign consent JWTs,
	// 3. sign organization related events (e.g. endpoint registration).
	OrgKeys []interface{} `json:"orgKeys,omitempty"`
	Start   time.Time     `json:"start"`
	End     *time.Time    `json:"end,omitempty"`
}

func (v VendorClaimEvent) PostProcessUnmarshal(event events.Event) error {
	if err := cert.ValidateJWK(v.OrgKeys...); err != nil {
		return err
	}
	for _, key := range v.OrgKeys {
		if certificate := cert.GetCertificate(key); certificate != nil {
			var err error
			var organizationId string
			if organizationId, err = cert2.GetOrganizationSubjectAltName(certificate); err != nil {
				return errors2.Wrap(err, "unable to unmarshal organization ID from certificate")
			}
			if string(v.OrgIdentifier) != organizationId {
				return fmt.Errorf("organization ID in certificate (%s) doesn't match event (%s)", organizationId, v.OrgIdentifier)
			}
		}
	}
	return nil
}

func GetEventTypes() []events.EventType {
	return []events.EventType{
		RegisterEndpoint,
		RegisterVendor,
		VendorClaim,
	}
}

type trustStore struct {
	certPool *x509.CertPool
}

func NewTrustStore() events.TrustStore {
	return &trustStore{certPool: x509.NewCertPool()}
}

func (t *trustStore) RegisterEventHandlers(fn func(events.EventType, events.EventHandler)) {
	for _, eventType := range GetEventTypes() {
		fn(eventType, t.handleEvent)
	}
}

func (t trustStore) Verify(certificate *x509.Certificate, moment time.Time) error {
	chains, err := certificate.Verify(x509.VerifyOptions{Roots: t.certPool, CurrentTime: moment})
	if err != nil {
		return err
	}
	// Make sure that all certificates in the chain have the same domain
	if err = verifyCertChainNutsDomain(chains[0]); err != nil {
		// TODO: Nuts Domain is in PoC state, should be made mandatory later
		// https://github.com/nuts-foundation/nuts-registry/issues/120
		log.Logger().Warnf("Couldn't validate Nuts domain: %v", err)
	}
	return nil
}

func (t *trustStore) handleEvent(event events.Event, _ events.EventLookup) error {
	if event.Type() == RegisterVendor {
		// Vendors (for now) self-sign their vendor CA certificates. These have to be in our trust store.
		payload := RegisterVendorEvent{}
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		for _, key := range payload.Keys {
			chain, err := cert.MapToX509CertChain(key.(map[string]interface{}))
			if err != nil {
				return err
			}
			// Certificates are self-signed, so chain length should be 1
			if len(chain) == 1 {
				t.certPool.AddCert(chain[0])
				logrus.Infof("Registered self-signed vendor CA certificate for vendor: %s", payload.Name)
			} else {
				return errors.New("unexpected X.509 certificate chain length for vendor (it should be self-signed)")
			}
		}
	} else if event.Type() == VendorClaim {
		payload := VendorClaimEvent{}
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		for _, key := range payload.OrgKeys {
			chain, err := cert.MapToX509CertChain(key.(map[string]interface{}))
			if err != nil {
				return err
			}
			if len(chain) > 0 {
				// Make sure the certificate is issued by a trusted vendor
				certificate := chain[0]
				if err := t.Verify(certificate, event.IssuedAt()); err != nil {
					return errors2.Wrap(err, "organization certificate is not trusted (issued by untrusted vendor certificate?)")
				}
				// We only add the actual certificate, since the issuing certificate is vendor CA certificate, which is
				// already in our chain.
				t.certPool.AddCert(certificate)
			}
		}
	}
	return nil
}

// verifyCertChainNutsDomain verifies that all certificates contain the right domain. The expected domain is taken from
// the topmost certificate and should not be empty or missing.
// If one of the certificates violate this condition or it couldn't be checked, an error is returned. If all is OK, nil is returned.
func verifyCertChainNutsDomain(chain []*x509.Certificate) error {
	var expectedDomain string
	for _, certificate := range chain {
		domainInCert, err := cert2.GetDomain(certificate)
		if err != nil {
			return err
		}
		if domainInCert == "" {
			return fmt.Errorf("certificate is missing domain (subject: %s)", certificate.Subject.String())
		}
		if expectedDomain == "" {
			expectedDomain = domainInCert
		}
		if expectedDomain != domainInCert {
			return fmt.Errorf("domain (%s) in certificate (subject: %s) differs from expected domain (%s)", domainInCert, certificate.Subject.String(), expectedDomain)
		}
	}
	return nil
}
