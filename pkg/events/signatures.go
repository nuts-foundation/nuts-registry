package events

import (
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SignatureValidator validates event signatures.
type SignatureValidator struct {
	verifier     JwsVerifier
	certVerifier cert.Verifier
}

// NewSignatureValidator creates a new SignatureValidator for the given event types.
func NewSignatureValidator(verifier JwsVerifier, certVerifier cert.Verifier) SignatureValidator {
	return SignatureValidator{verifier: verifier, certVerifier: certVerifier}
}

// RegisterEventHandlers registers event handlers which will validate the event signatures.
func (v SignatureValidator) RegisterEventHandlers(fn EventRegistrar, eventType []EventType) {
	for _, eventType := range eventType {
		fn(eventType, v.validate)
	}
}

func (v SignatureValidator) validate(event Event, _ EventLookup) error {
	if len(event.Signature()) == 0 {
		// https://github.com/nuts-foundation/nuts-registry/issues/84
		logrus.Warnf("Event not signed, this is accepted for now but it will be rejected in future (event = %v).", event.IssuedAt())
	} else {
		// TODO: event.IssuedAt is not signed, what extra safety does it add checking it against the certificate validity?
		// TODO: is the event signed by the expected entity (correct vendor/organization)?
		if event.Version() <= currentEventVersion {
			_, err := v.verifier(event.Signature(), event.IssuedAt(), v.certVerifier)
			if err := err; err != nil {
				return errors2.Wrapf(err, "event signature verification failed, it will not be processed (event = %v)", event.IssuedAt())
			}
		} else {
			logrus.Warnf("Unsupported signature version (%d), unable to validate signature. This should be fixed in the future using canonicalization (event = %v).", event.Version(), event.IssuedAt())
		}
	}
	return nil
}