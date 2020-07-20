package domain

import (
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterEndpointEvent(t *testing.T) {
	t.Run("unmarshal event with no post processors", func(t *testing.T) {
		event := events.CreateEvent(RegisterEndpoint, RegisterEndpointEvent{}, nil)
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var registerEndpointEvent = RegisterEndpointEvent{}
		err := unmarshalledEvent.Unmarshal(&registerEndpointEvent)
		if !assert.NoError(t, err) {
			return
		}
	})
}
