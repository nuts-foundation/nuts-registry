package cert

import (
	"crypto/x509"
	"errors"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"sort"
	"time"
)

// GetActiveCertificates converts the given JWKs to X509 certificates and returns them sorted,
// longest valid certificate first. Expired certificates aren't returned.
func GetActiveCertificates(jwks []interface{}, instant time.Time) []*x509.Certificate {
	var activeCertificates []*x509.Certificate
	for _, keyAsMap := range jwks {
		chain, err := jwkMapToCertChain(keyAsMap)
		if err != nil {
			continue
		}
		if len(chain) == 0 {
			continue
		}
		activeCertificates = append(activeCertificates, chain[0])
	}
	sort.Slice(activeCertificates, func(i, j int) bool {
		first := activeCertificates[i]
		second := activeCertificates[j]
		return first.NotAfter.UnixNano()-instant.UnixNano() > second.NotAfter.UnixNano()-instant.UnixNano()
	})
	return activeCertificates
}

func jwkMapToCertChain(keyAsMap interface{}) ([]*x509.Certificate, error) {
	key, err := crypto.MapToJwk(keyAsMap.(map[string]interface{}))
	if err != nil {
		return nil, err
	}
	chainInterf, exists := key.Get("x5c")
	if !exists {
		// JWK does not contain x5c component (X.509 certificate chain)
		return nil, errors.New("JWK has no x5c field")
	}
	return chainInterf.([]*x509.Certificate), nil
}
