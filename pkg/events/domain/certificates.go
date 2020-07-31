package domain

import (
	"crypto/x509"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type certificateEventHandler struct {
	trustStore cert.TrustStore
}

// NewCertificateEventHandler constructs a new handler that verifies certificates in events
// and adds them to the truststore if necessary.
func NewCertificateEventHandler(trustStore cert.TrustStore) events.TrustStore {
	return &certificateEventHandler{trustStore: trustStore}
}

// RegisterEventHandlers this event handler which is required for it to actually work.
func (t *certificateEventHandler) RegisterEventHandlers(fn func(events.EventType, events.EventHandler)) {
	for _, eventType := range GetEventTypes() {
		fn(eventType, t.handleEvent)
	}
}

// Verify verifies the given certificate against the truststore. In addition it also verifies the correctness of the
// "Nuts Domain" in the certificate tree.
func (t certificateEventHandler) Verify(certificate *x509.Certificate, moment time.Time) error {
	_, err := t.verify(certificate, moment)
	return err
}

func (t certificateEventHandler) verify(certificate *x509.Certificate, moment time.Time) ([]*x509.Certificate, error) {
	chains, err := certificate.Verify(x509.VerifyOptions{Roots: t.trustStore.Pool(), CurrentTime: moment})
	if err != nil {
		return nil, err
	}
	// Make sure that all certificates in the chain have the same domain
	if err = verifyCertChainNutsDomain(chains[0]); err != nil {
		// TODO: Nuts Domain is in PoC state, should be made mandatory later
		// https://github.com/nuts-foundation/nuts-registry/issues/120
		logrus.Debugf("Nuts domain verification failed: %v", err)
	}
	// We're not supporting multiple chains
	return chains[0], nil
}

func (t *certificateEventHandler) handleEvent(event events.Event, _ events.EventLookup) error {
	certificates := make([]*x509.Certificate, 0)
	var err error
	if event.Type() == RegisterVendor {
		// Vendors (for now) self-sign their vendor CA certificates. These have to be in our trust store.
		payload := RegisterVendorEvent{}
		if err = event.Unmarshal(&payload); err != nil {
			return err
		}
		if certificates, err = t.getCertificatesToBeTrusted(payload.Keys, event.IssuedAt(), false); err != nil {
			return errors2.Wrap(err, "certificate problem in RegisterVendor event")
		}
	} else if event.Type() == VendorClaim {
		payload := VendorClaimEvent{}
		if err := event.Unmarshal(&payload); err != nil {
			return err
		}
		if certificates, err = t.getCertificatesToBeTrusted(payload.OrgKeys, event.IssuedAt(), true); err != nil {
			return errors2.Wrap(err, "certificate problem in VendorClaim event")
		}
	}
	for _, certificate := range certificates {
		if err = t.trustStore.AddCertificate(certificate); err != nil {
			return errors2.Wrap(err, "unable to add certificate to truststore")
		}
	}

	return nil
}

func (t *certificateEventHandler) getCertificatesToBeTrusted(jwks []interface{}, moment time.Time, mustVerify bool) ([]*x509.Certificate, error) {
	certificates := make([]*x509.Certificate, 0)
	for _, key := range jwks {
		chain, err := cert.MapToX509CertChain(key.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		if len(chain) == 0 {
			continue
		}
		if mustVerify {
			if verifiedChain, err := t.verify(chain[0], moment); err != nil {
				return nil, errors2.Wrapf(err, "certificate not trusted: %s (issuer: %s, serial: %d)", chain[0].Subject, chain[0].Issuer, chain[0].SerialNumber)
			} else {
				chain = verifiedChain
			}
		}
		for _, certificate := range chain {
			if certificate.IsCA {
				certificates = append(certificates, certificate)
			}
		}
	}
	return certificates, nil
}

// verifyCertChainNutsDomain verifies that all certificates contain the right domain. The expected domain is taken from
// the topmost certificate and should not be empty or missing.
// If one of the certificates violate this condition or it couldn't be checked, an error is returned. If all is OK, nil is returned.
func verifyCertChainNutsDomain(chain []*x509.Certificate) error {
	var expectedDomain string
	for _, certificate := range chain {
		domainInCert, err := cert2.NewNutsCertificate(certificate).GetDomain()
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
		// TODO: Check that domain is one of the known types
	}
	return nil
}
