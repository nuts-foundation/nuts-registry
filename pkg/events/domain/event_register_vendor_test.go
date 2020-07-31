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
}

func TestVendorEventMatcher(t *testing.T) {
	assert.False(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent(RegisterVendor, RegisterVendorEvent{}, nil)))
	assert.False(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent("foobar", struct{}{}, nil)))
	assert.True(t, VendorEventMatcher(test.VendorID("123"))(events.CreateEvent(RegisterVendor, RegisterVendorEvent{Identifier: test.VendorID("123")}, nil)))
}
