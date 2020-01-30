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
	"testing"
)

func TestUnknownEventType(t *testing.T) {
	system := NewEventSystem()
	input := `{
		"type": "non-existing"
	}`
	event, err := EventFromJSON([]byte(input))
	assert.NoError(t, err)
	err = system.ProcessEvent(event)
	assert.Error(t, err, "unknown event type: non-existing")
}

func TestNoEventHandler(t *testing.T) {
	system := NewEventSystem()
	input := `{
		"type": "RegisterOrganizationEvent"
	}`
	event, err := EventFromJSON([]byte(input))
	assert.NoError(t, err)
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

	err := system.LoadAndApplyEvents("../../test_data/valid_files")

	assert.NoError(t, err)
	assert.Equal(t, 2, organizationsCreated, "unexpected number of events for: RegisterOrganization")
	assert.Equal(t, 2, endpointsCreated, "unexpected number of events for: RegisterEndpoint")
	assert.Equal(t, 2, endpointOrganizationsCreated, "unexpected number of events for: RegisterEndpointOrganization")
}

func TestLoadEventsInvalidJson(t *testing.T) {
	system := NewEventSystem()
	err := system.LoadAndApplyEvents("../../test_data/invalid_files")
	assert.EqualError(t, err, "invalid character '{' looking for beginning of object key string")
}

func TestLoadEventsEmptyFile(t *testing.T) {
	system := NewEventSystem()
	err := system.LoadAndApplyEvents("../../test_data/empty_files")
	assert.EqualError(t, err, "unexpected end of JSON input")
}
