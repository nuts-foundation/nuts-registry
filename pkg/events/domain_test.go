package events

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterVendorEvent(t *testing.T) {
	t.Run("check default domain fallback", func(t *testing.T) {
		event := CreateEvent(RegisterVendor, RegisterVendorEvent{})
		data := event.Marshal()
		unmarshalledEvent, _ := EventFromJSON(data)
		var registerVendorEvent = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&registerVendorEvent)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", registerVendorEvent.Domain)
	})
	t.Run("invalid JWK", func(t *testing.T) {
		event := CreateEvent(RegisterVendor, RegisterVendorEvent{
			Keys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		data := event.Marshal()
		unmarshalledEvent, _ := EventFromJSON(data)
		var payload = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestVendorClaimEvent(t *testing.T) {
	t.Run("invalid JWK", func(t *testing.T) {
		event := CreateEvent(VendorClaim, VendorClaimEvent{
			VendorIdentifier: "v1",
			OrgKeys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		data := event.Marshal()
		unmarshalledEvent, _ := EventFromJSON(data)
		var payload = VendorClaimEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestRegisterEndpointEvent(t *testing.T) {
	t.Run("unmarshal event with no post processors", func(t *testing.T) {
		event := CreateEvent(RegisterEndpoint, RegisterEndpointEvent{})
		data := event.Marshal()
		unmarshalledEvent, _ := EventFromJSON(data)
		var registerEndpointEvent = RegisterEndpointEvent{}
		err := unmarshalledEvent.Unmarshal(&registerEndpointEvent)
		if !assert.NoError(t, err) {
			return
		}
	})
}
