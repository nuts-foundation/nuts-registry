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
	"encoding/json"
	"errors"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"time"
)

// Event defines an event which can be (un)marshalled.
type Event interface {
	Type() EventType
	IssuedAt() time.Time
	Unmarshal(out interface{}) error
	Marshal() []byte
}

// EventType defines a supported type of event, which is used for executing the right handler.
type EventType string

const (
	// RegisterOrganization event type
	RegisterOrganization         EventType = "RegisterOrganizationEvent"
	// RemoveOrganization event type
	RemoveOrganization           EventType = "RemoveOrganizationEvent"
	// RegisterEndpoint event type
	RegisterEndpoint             EventType = "RegisterEndpointEvent"
	// RegisterEndpointOrganization event type
	RegisterEndpointOrganization EventType = "RegisterEndpointOrganizationEvent"
)

// ErrMissingEventType is given when the event being unmarshalled has no type attribute.
var ErrMissingEventType = errors.New("unmarshalling error: missing event type")

var eventTypes []EventType

func init() {
	eventTypes = []EventType{
		RegisterOrganization,
		RemoveOrganization,
		RegisterEndpoint,
		RegisterEndpointOrganization,
	}
}

// IsEventType checks whether the given type is supported.
func IsEventType(eventType EventType) bool {
	for _, actual := range eventTypes {
		if actual == eventType {
			return true
		}
	}
	return false
}

// RegisterOrganizationEvent event
type RegisterOrganizationEvent struct {
	Organization db.Organization `json:"payload"`
}

// RemoveOrganizationEvent event
type RemoveOrganizationEvent struct {
	OrganizationID string `json:"payload"`
}

// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Endpoint db.Endpoint `json:"payload"`
}

// RegisterEndpointOrganizationEvent event
type RegisterEndpointOrganizationEvent struct {
	EndpointOrganization db.EndpointOrganization `json:"payload"`
}

type jsonEvent struct {
	EventType     string                 `json:"type"`
	EventIssuedAt time.Time              `json:"issuedAt"`
	EventPayload  map[string]interface{} `json:"payload"`
	data          []byte
}

// EventFromJSON unmarshals an event. If the event can't be unmarshalled, an error is returned.
func EventFromJSON(data []byte) (Event, error) {
	e := jsonEvent{}
	err := json.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}
	if e.EventType == "" {
		return nil, ErrMissingEventType
	}
	e.data = data
	return e, nil
}

// CreateEvent creates an event of the given type and the provided payload. If the event can't be created, an error is
// returned.
func CreateEvent(eventType EventType, payload interface{}) (Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	payloadMap := map[string]interface{}{}
	if err := json.Unmarshal(data, &payloadMap); err != nil {
		return nil, err
	}
	return jsonEvent{
		EventType:     string(eventType),
		EventIssuedAt: time.Now(),
		EventPayload:  payloadMap,
		data:          data,
	}, nil
}

func (j jsonEvent) IssuedAt() time.Time {
	return j.EventIssuedAt
}

func (j jsonEvent) Type() EventType {
	return EventType(j.EventType)
}

func (j jsonEvent) Unmarshal(out interface{}) error {
	return json.Unmarshal(j.data, out)
}

func (j jsonEvent) Marshal() []byte {
	return j.data
}
