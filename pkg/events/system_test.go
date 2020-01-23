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
	"reflect"
	"testing"
)

func TestProcessEvent(t *testing.T) {
	type event1 struct {
	}
	system := NewEventSystem()
	var called = 0
	system.RegisterEventHandler(reflect.TypeOf(event1{}), func(event interface{}) error {
		_, ok := event.(event1)
		assert.True(t, ok)
		called++
		return nil
	})
	assert.Nil(t, system.ProcessEvent(event1{}))
	assert.Equal(t, 1, called)
}

func TestNoEventProcessor(t *testing.T) {
	type event1 struct {
	}
	system := NewEventSystem()
	assert.Equal(t, "no handler registered for event (type = events.event1), handlers are: map[]", system.ProcessEvent(event1{}).Error())
}

func TestLoadEventsFromFile(t *testing.T) {
	system := NewEventSystem()
	organizationsCreated := 0
	system.RegisterEventHandler(reflect.TypeOf(&RegisterOrganizationEvent{}), func(i interface{}) error {
		organizationsCreated++
		return nil
	})
	endpointsCreated := 0
	system.RegisterEventHandler(reflect.TypeOf(&RegisterEndpointEvent{}), func(i interface{}) error {
		endpointsCreated++
		return nil
	})
	endpointOrganizationsCreated := 0
	system.RegisterEventHandler(reflect.TypeOf(&RegisterEndpointOrganizationEvent{}), func(i interface{}) error {
		endpointOrganizationsCreated++
		return nil
	})

	assert.Equal(t, reflect.TypeOf(RegisterEndpointOrganizationEvent{}), reflect.TypeOf(RegisterEndpointOrganizationEvent{}))
	err := system.LoadAndApplyEvents("../../test_data/valid_files")

	assert.NoError(t, err)
	assert.Equal(t, 2, organizationsCreated)
	assert.Equal(t, 2, endpointsCreated)
	assert.Equal(t, 2, endpointOrganizationsCreated)
}