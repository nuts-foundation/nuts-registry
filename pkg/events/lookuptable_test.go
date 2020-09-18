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
package events

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const eventType = "test"

var eventPayload = struct{}{}

func Test_EventLookup_register(t *testing.T) {
	lut := newEventLookupTable()
	t.Run("ok", func(t *testing.T) {
		err := lut.register(CreateEvent(eventType, eventPayload, nil))
		assert.NoError(t, err)
	})
	t.Run("error - referred event not found", func(t *testing.T) {
		err := lut.register(CreateEvent(eventType, eventPayload, []byte{1, 2, 3}))
		assert.EqualError(t, err, "previous event not found: 010203")
	})
	t.Run("error - referred event is of different type", func(t *testing.T) {
		event1 := CreateEvent(eventType, eventPayload, nil)
		lut.register(event1)
		err := lut.register(CreateEvent(eventType+"2", eventPayload, event1.Ref()))
		assert.Contains(t, err.Error(), "previous event type differs")
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
		assert.EqualError(t, err, "previous event ("+event1.Ref().String()+") already referred to")
	})
}

func Test_EventLookup_Lookup(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		lut := newEventLookupTable()
		event1 := CreateEvent(eventType, eventPayload, nil)
		lut.register(event1)
		event := lut.Get(event1.Ref())
		assert.Equal(t, event1, event)
	})
	t.Run("ok - not found", func(t *testing.T) {
		lut := newEventLookupTable()
		event := lut.Get([]byte{1, 2, 3})
		assert.Nil(t, event)
	})
}

func Test_EventLookup_FindLastEvent(t *testing.T) {
	t.Run("ok - all match, 1-member path", func(t *testing.T) {
		lut := newEventLookupTable()
		lut.register(CreateEvent(eventType, eventPayload, nil))
		event, err := lut.FindLastEvent(func(event Event) bool {
			return true
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("ok - all events match, 2-member path", func(t *testing.T) {
		lut := newEventLookupTable()
		event1 := CreateEvent(eventType, eventPayload, nil)
		lut.register(event1)
		lut.register(CreateEvent(eventType, eventPayload, event1.Ref()))
		event, err := lut.FindLastEvent(func(event Event) bool {
			return true
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("ok - second event matches, 2-member path", func(t *testing.T) {
		lut := newEventLookupTable()
		event1 := CreateEvent(eventType, eventPayload, nil)
		lut.register(event1)
		event2 := CreateEvent(eventType, eventPayload, event1.Ref())
		lut.register(event2)
		event, err := lut.FindLastEvent(func(event Event) bool {
			return event == event2
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("ok - no matches", func(t *testing.T) {
		lut := newEventLookupTable()
		event, err := lut.FindLastEvent(func(event Event) bool {
			return true
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, event)
	})
	t.Run("error - multiple paths match", func(t *testing.T) {
		lut := newEventLookupTable()
		lut.register(CreateEvent(eventType, eventPayload, nil))
		lut.register(CreateEvent(eventType, eventPayload, nil))
		event, err := lut.FindLastEvent(func(event Event) bool {
			return true
		})
		assert.Nil(t, event)
		assert.EqualError(t, err, "multiple event paths match")
	})
}
