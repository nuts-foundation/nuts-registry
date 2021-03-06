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
