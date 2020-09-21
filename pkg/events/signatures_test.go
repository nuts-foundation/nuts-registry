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
	"errors"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestSignatureValidator_RegisterEventHandlers(t *testing.T) {
	fn := func(eventType EventType, _ EventHandler) {
		assert.Equal(t, string(eventType), "foo")
	}
	NewSignatureValidator(test.NoopJwsVerifier, test.NoopCertificateVerifier).RegisterEventHandlers(fn, []EventType{"foo"})
}

func TestSignatureValidator_verify(t *testing.T) {
	t.Run("ok - signed", func(t *testing.T) {
		event := CreateEvent("foo", struct{}{}, nil)
		event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return bytes2, nil
		})
		verifier := func(signature []byte, signingTime time.Time, verifier cert.Verifier) (bytes []byte, err error) {
			return signature, nil
		}
		err := NewSignatureValidator(verifier, test.NoopCertificateVerifier).validate(event, nil)
		assert.NoError(t, err)
	})
	t.Run("ok - not signed", func(t *testing.T) {
		err := NewSignatureValidator(test.NoopJwsVerifier, test.NoopCertificateVerifier).validate(CreateEvent("foo", struct{}{}, nil), nil)
		assert.NoError(t, err)
	})
	t.Run("error - verification failed", func(t *testing.T) {
		event := CreateEvent("foo", struct{}{}, nil)
		event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return bytes2, nil
		})
		verifier := func(signature []byte, signingTime time.Time, verifier cert.Verifier) (bytes []byte, err error) {
			return nil, errors.New("failed")
		}
		err := NewSignatureValidator(verifier, test.NoopCertificateVerifier).validate(event, nil)
		assert.Error(t, err)
	})
}