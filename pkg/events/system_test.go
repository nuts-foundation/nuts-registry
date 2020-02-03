/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
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
	"io/ioutil"
	"os"
	"testing"
)

func TestUnknownEventType(t *testing.T) {
	system := NewEventSystem()
	input := `{
		"type": "non-existing"
	}`
	event, err := EventFromJSON([]byte(input))
	if !assert.NoError(t, err) {
		return
	}
	err = system.ProcessEvent(event)
	assert.Error(t, err, "unknown event type: non-existing")
}

func TestNoEventHandler(t *testing.T) {
	system := NewEventSystem()
	input := `{
		"type": "RegisterOrganizationEvent"
	}`
	event, err := EventFromJSON([]byte(input))
	if !assert.NoError(t, err) {
		return
	}
	err = system.ProcessEvent(event)
	assert.Error(t, err, "no handler registered for event (type = RegisterOrganizationEvent), handlers are: map[]")
}

func TestLoadEvents(t *testing.T) {
	system := NewEventSystem()
	organizationsCreated := 0
	system.RegisterEventHandler(RegisterOrganization, func(e Event) error {
		organizationsCreated++
		return nil
	})
	endpointsCreated := 0
	system.RegisterEventHandler(RegisterEndpoint, func(e Event) error {
		endpointsCreated++
		return nil
	})
	endpointOrganizationsCreated := 0
	system.RegisterEventHandler(RegisterEndpointOrganization, func(e Event) error {
		endpointOrganizationsCreated++
		return nil
	})

	assertEventsHandled := func(oc int, ec int, eoc int) {
		assert.Equal(t, oc, organizationsCreated, "unexpected number of events for: RegisterOrganization")
		assert.Equal(t, ec, endpointsCreated, "unexpected number of events for: RegisterEndpoint")
		assert.Equal(t, eoc, endpointOrganizationsCreated, "unexpected number of events for: RegisterEndpointOrganization")
	}

	const dir = "../../test_data/valid_files/events"
	t.Run("All fresh system state, all events should be loaded", func(t *testing.T) {
		err := system.LoadAndApplyEvents(dir)
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(2, 2, 2)
	})

	const newFile = dir + "/20210123091400006-RegisterEndpointOrganizationEvent.json"
	t.Run("New event file, should trigger an incremental change", func(t *testing.T) {
		err := cp(dir+"/20200123091400006-RegisterEndpointOrganizationEvent.json", newFile)
		if !assert.NoError(t, err) {
			return
		}
		err = system.LoadAndApplyEvents(dir)
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(2, 2, 3)
	})
	defer os.Remove(newFile)

	t.Run("No incremental change", func(t *testing.T) {
		err := system.LoadAndApplyEvents(dir)
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(2, 2, 3)
	})
}

func TestLoadEventsInvalidJson(t *testing.T) {
	system := NewEventSystem()
	err := system.LoadAndApplyEvents("../../test_data/invalid_files/events")
	assert.EqualError(t, err, "invalid character '{' looking for beginning of object key string")
}

func TestLoadEventsEmptyFile(t *testing.T) {
	system := NewEventSystem()
	err := system.LoadAndApplyEvents("../../test_data/empty_files/events")
	assert.EqualError(t, err, "unexpected end of JSON input")
}

func cp(src string, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

func TestParseTimestamp(t *testing.T) {
	t.Run("Timestamp OK", func(t *testing.T) {
		timestamp, err := parseTimestamp("20200123091400001")
		assert.Equal(t, "2020-01-23 09:14:00.001 +0000 UTC", timestamp.String())
		assert.NoError(t, err)
	})
	t.Run("Timestamp has invalid length", func(t *testing.T) {
		timestamp, err := parseTimestamp("asdasd")
		assert.True(t, timestamp.IsZero())
		assert.Error(t, err)
	})
	t.Run("Timestamp has invalid characters", func(t *testing.T) {
		timestamp, err := parseTimestamp("a2345678901234567")
		assert.True(t, timestamp.IsZero())
		assert.Error(t, err)
	})

}

func TestPublishEventCallsEventHandlers(t *testing.T) {
	system := NewEventSystem()
	called := 0
	system.RegisterEventHandler(RegisterOrganization, func(event Event) error {
		called++
		return nil
	})
	err := system.PublishEvent(jsonEvent{EventType: string(RegisterOrganization)})
	assert.NoError(t, err)
	assert.Equal(t, called, 1)
}
