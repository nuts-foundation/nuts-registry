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
package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	test2 "github.com/nuts-foundation/nuts-crypto/test"
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
	vendorID := test.VendorID("VendorID")
	orgID := test.OrganizationID("abc")
	t.Run("ok - signed by vendor signing certificate (>= 0.15)", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(vendorID, "Vendor Name", "", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: orgID,
			VendorID:       vendorID,
			OrgKeys:        []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.NoError(t, err)
	})
	t.Run("ok - signed by organization certificate (<= 0.14)", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", orgID, "Org", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: orgID,
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
			OrganizationID: orgID,
			OrgKeys:        []interface{}{jwkAsMap},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.NoError(t, err)
	})
	t.Run("error - vendor ID in certificate doesn't match", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("SomeOtherID"), "Some other vendor", "", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: orgID,
			VendorID:       vendorID,
			OrgKeys:        []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.EqualError(t, err, "vendor ID validation failed: vendor ID in certificate (urn:oid:1.3.6.1.4.1.54851.4:SomeOtherID) doesn't match event (VendorID)")
	})
	t.Run("error - organization ID in certificate doesn't match", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", test.OrganizationID("def"), "Org", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrganizationID: orgID,
			OrgKeys:        []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.EqualError(t, err, "organization ID validation failed: organization ID in certificate (urn:oid:2.16.840.1.113883.2.4.6.1:def) doesn't match event (urn:oid:2.16.840.1.113883.2.4.6.1:abc)")
	})
	t.Run("error - no vendor ID or organization ID in certificate", func(t *testing.T) {
		privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		cert, _ := x509.ParseCertificate(test2.GenerateCertificate(time.Now(), 2, privateKey))
		event := VendorClaimEvent{
			OrganizationID: orgID,
			OrgKeys:        []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event, nil))
		assert.EqualError(t, err, "event should either be signed by organization or vendor signing certificate")
	})
}

func TestOrganizationEventMatcher(t *testing.T) {
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent("foobar", struct{}{}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorID: test.VendorID("123")}, nil)))
	assert.False(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{OrganizationID: test.OrganizationID("456")}, nil)))
	assert.True(t, OrganizationEventMatcher(test.VendorID("123"), test.OrganizationID("456"))(events.CreateEvent(VendorClaim, VendorClaimEvent{VendorID: test.VendorID("123"), OrganizationID: test.OrganizationID("456")}, nil)))
}
