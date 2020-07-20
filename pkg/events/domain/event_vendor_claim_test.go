package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVendorClaimEvent(t *testing.T) {
	t.Run("invalid JWK", func(t *testing.T) {
		event := events.CreateEvent(VendorClaim, VendorClaimEvent{
			VendorIdentifier: "v1",
			OrgKeys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		}, nil)
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var payload = VendorClaimEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestVendorClaimEvent_PostProcessUnmarshal(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", "abc", "Org", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.NoError(t, err)
	})
	t.Run("ok - key does not contain certificate", func(t *testing.T) {
		key, _ := rsa.GenerateKey(rand.Reader, 1024)
		keyAsJwk, _ := jwk.New(key)
		jwkAsMap, _ := cert.JwkToMap(keyAsJwk)
		jwkAsMap["kty"] = string(jwkAsMap["kty"].(jwa.KeyType))
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{jwkAsMap},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.NoError(t, err)
	})
	t.Run("error - certificate organization doesn't match", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", "def", "Org", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.EqualError(t, err, "organization ID in certificate (def) doesn't match event (abc)")
	})
}

func TestOrganizationEventMatcher(t *testing.T) {
	assert.False(t, OrganizationEventMatcher("123", "456")(events.CreateEvent("foobar", struct{}{}, nil)))
	assert.False(t, OrganizationEventMatcher("123", "456")(events.CreateEvent(VendorClaim, VendorClaimEvent{}, nil)))
	assert.False(t, OrganizationEventMatcher("123", "456")(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorIdentifier: "123"}, nil)))
	assert.False(t, OrganizationEventMatcher("123", "456")(events.CreateEvent(VendorClaim, VendorClaimEvent{OrgIdentifier: "456"}, nil)))
	assert.True(t, OrganizationEventMatcher("123", "456")(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorIdentifier: "123", OrgIdentifier: "456"}, nil)))
}
