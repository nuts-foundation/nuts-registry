package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetActiveCertificates(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	t.Run("no keys", func(t *testing.T) {
		certificates := GetActiveCertificates(make([]interface{}, 0), time.Now())
		assert.Empty(t, certificates)
	})
	t.Run("no certificate for key", func(t *testing.T) {
		key, _ := jwk.New(rsaKey)
		certificates := GetActiveCertificates([]interface{}{jwkToMap(key)}, time.Now())
		assert.Empty(t, certificates)
	})
	t.Run("single entry", func(t *testing.T) {
		certBytes := test.GenerateCertificateEx(time.Now().AddDate(0, 0, -1), 2, rsaKey)
		cert, err := x509.ParseCertificate(certBytes)
		if !assert.NoError(t, err) {
			return
		}
		key, _ := pkg.CertificateToJWK(cert)
		certs := GetActiveCertificates([]interface{}{jwkToMap(key)}, time.Now())
		assert.Len(t, certs, 1)
	})
}

func jwkToMap(key jwk.Key) map[string]interface{} {
	m, _ := pkg.JwkToMap(key)
	keyAsJSON, _ := json.MarshalIndent(m, "", "  ")
	j := map[string]interface{}{}
	json.Unmarshal(keyAsJSON, &j)
	return j
}