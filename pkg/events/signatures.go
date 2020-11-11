/*
 * Nuts registry
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */
package events

import (
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/logging"
	errors2 "github.com/pkg/errors"
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
		logging.Log().Warnf("Event not signed, this is accepted for now but it will be rejected in future (event = %v).", event.IssuedAt())
	} else {
		// TODO: event.IssuedAt is not signed, what extra safety does it add checking it against the certificate validity?
		// TODO: is the event signed by the expected entity (correct vendor/organization)?
		if event.Version() <= currentEventVersion {
			_, err := v.verifier(event.Signature(), event.IssuedAt(), v.certVerifier)
			if err := err; err != nil {
				return errors2.Wrapf(err, "event signature verification failed, it will not be processed (event = %v)", event.IssuedAt())
			}
		} else {
			logging.Log().Warnf("Unsupported signature version (%d), unable to validate signature. This should be fixed in the future using canonicalization (event = %v).", event.Version(), event.IssuedAt())
		}
	}
	return nil
}