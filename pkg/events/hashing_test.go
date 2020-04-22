package events

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const eventType = "test"
var eventPayload = struct{}{}

func Test_eventLookupTable(t *testing.T) {
	lut := newEventLookupTable()
	t.Run("ok", func(t *testing.T) {
		err := lut.register(CreateEvent(eventType, eventPayload, nil))
		assert.NoError(t, err)
	})
	t.Run("error - referred event not found", func(t *testing.T) {
		err := lut.register(CreateEvent(eventType, eventPayload, []byte{1, 2, 3}))
		assert.EqualError(t, err, "previous event not found: 010203")
	})
	t.Run("error - event already referred to", func(t *testing.T) {
		event1 := CreateEvent(eventType, eventPayload, nil)
		err := lut.register(event1)
		if !assert.NoError(t, err) {
			return
		}
		event2 := CreateEvent(eventType, eventPayload, event1.Ref())
		err = lut.register(event2)
		if !assert.NoError(t, err) {
			return
		}
		err = lut.register(CreateEvent(eventType, eventPayload, event1.Ref()))
		assert.EqualError(t, err, "previous event (" + event1.Ref().String() + ") already referred to")
	})
}
