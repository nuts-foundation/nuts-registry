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
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRegisterVendorEvent(t *testing.T) {
	t.Run("check default domain fallback", func(t *testing.T) {
		event := events.CreateEvent(RegisterVendor, RegisterVendorEvent{}, nil)
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var registerVendorEvent = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&registerVendorEvent)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", registerVendorEvent.Domain)
	})
	t.Run("invalid JWK", func(t *testing.T) {
		event := events.CreateEvent(RegisterVendor, RegisterVendorEvent{
			Keys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		}, nil)
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var payload = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestRegisterVendorEvent_PostProcessUnmarshal(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("abc"), "Vendor", "CA", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := RegisterVendorEvent{
			Identifier: test.VendorID("abc"),
			Keys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event, nil))
		assert.NoError(t, err)
	})
	t.Run("ok - fallback to healthcare domain", func(t *testing.T) {
		event := RegisterVendorEvent{}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event, nil))
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", event.Domain)
	})
	t.Run("certificate vendor doesn't match", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("def"), "Vendor", "CA", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := RegisterVendorEvent{
			Identifier: test.VendorID("abc"),
			Keys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event, nil))
		assert.EqualError(t, err, "vendor ID in certificate (urn:oid:1.3.6.1.4.1.54851.4:def) doesn't match event (urn:oid:1.3.6.1.4.1.54851.4:abc)")
	})
	t.Run("backwards compatibility for string x5c instead of []string", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("abc"), "Vendor", "CA", types.HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		// Craft event with a string x5c instead of []string
		event := RegisterVendorEvent{
			Identifier: test.VendorID("abc"),
			Keys:       []interface{}{certToMap(cert)},
		}
		keyAsMap := event.Keys[0].(map[string]interface{})
		keyAsMap["x5c"] = keyAsMap["x5c"].([]string)[0]
		err := event.PostProcessUnmarshal(nil)
		if !assert.NoError(t, err) {
			return
		}
	})
}

func TestVendorEventMatcher(t *testing.T) {
	assert.False(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent(RegisterVendor, RegisterVendorEvent{}, nil)))
	assert.False(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent("foobar", struct{}{}, nil)))
	assert.True(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent(RegisterVendor, RegisterVendorEvent{Identifier: test.VendorID("123")}, nil)))
}
