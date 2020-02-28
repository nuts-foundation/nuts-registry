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

