package events

import (
	"errors"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestSignatureValidator_RegisterEventHandlers(t *testing.T) {
	fn := func(eventType EventType, _ EventHandler) {
		assert.Equal(t, string(eventType), "foo")
	}
	NewSignatureValidator(NoopJwsVerifier, NoopTrustStore).RegisterEventHandlers(fn, []EventType{"foo"})
}

func TestSignatureValidator_verify(t *testing.T) {
	t.Run("ok - signed", func(t *testing.T) {
		event := CreateEvent("foo", struct{}{})
		event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return bytes2, nil
		})
		verifier := func(signature []byte, signingTime time.Time, verifier crypto.CertificateVerifier) (bytes []byte, err error) {
			return signature, nil
		}
		err := NewSignatureValidator(verifier, NoopTrustStore).validate(event)
		assert.NoError(t, err)
	})
	t.Run("ok - not signed", func(t *testing.T) {
		err := NewSignatureValidator(NoopJwsVerifier, NoopTrustStore).validate(CreateEvent("foo", struct{}{}))
		assert.NoError(t, err)
	})
	t.Run("error - verification failed", func(t *testing.T) {
		event := CreateEvent("foo", struct{}{})
		event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return bytes2, nil
		})
		verifier := func(signature []byte, signingTime time.Time, verifier crypto.CertificateVerifier) (bytes []byte, err error) {
			return nil, errors.New("failed")
		}
		err := NewSignatureValidator(verifier, NoopTrustStore).validate(event)
		assert.Error(t, err)
	})
}