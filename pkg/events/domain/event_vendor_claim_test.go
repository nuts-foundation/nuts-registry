package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVendorClaimEvent(t *testing.T) {
	t.Run("invalid JWK", func(t *testing.T) {
		event := events.CreateEvent(VendorClaim, VendorClaimEvent{
			VendorID: test.VendorID("v1"),
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
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", test.OrganizationID("abc"), "Org", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: test.OrganizationID("abc"),
			OrgKeys:        []interface{}{certToMap(cert)},
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
			OrganizationID: test.OrganizationID("abc"),
			OrgKeys:        []interface{}{jwkAsMap},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.NoError(t, err)
	})
	t.Run("error - certificate organization doesn't match", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", test.OrganizationID("def"), "Org", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: test.OrganizationID("abc"),
			OrgKeys:        []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.EqualError(t, err, "organization ID in certificate (urn:oid:2.16.840.1.113883.2.4.6.1:def) doesn't match event (urn:oid:2.16.840.1.113883.2.4.6.1:abc)")
	})
}

func TestOrganizationEventMatcher(t *testing.T) {
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent("foobar", struct{}{}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorID: test.VendorID("123")}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{OrganizationID: test.OrganizationID("456")}, nil)))
	assert.True(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorID: test.VendorID("123"), OrganizationID: test.OrganizationID("456")}, nil)))
}
