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
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"time"
)

type Event interface {
	Type() EventType
	IssuedAt() time.Time
	Unmarshal(out interface{}) error
	Marshal() []byte
}

type EventType string

const (
	RegisterOrganization         EventType = "RegisterOrganizationEvent"
	RemoveOrganization           EventType = "RemoveOrganizationEvent"
	RegisterEndpoint             EventType = "RegisterEndpointEvent"
	RegisterEndpointOrganization EventType = "RegisterEndpointOrganizationEvent"
)

var eventTypes []EventType

func init() {
	eventTypes = []EventType{
		RegisterOrganization,
		RemoveOrganization,
		RegisterEndpoint,
		RegisterEndpointOrganization,
	}
}

func IsEventType(eventType EventType) bool {
	for _, actual := range eventTypes {
		if actual == eventType {
			return true
		}
	}
	return false
}

type RegisterOrganizationEvent struct {
	Type         string
	Organization db.Organization `json:"payload"`
}

type RemoveOrganizationEvent struct {
	OrganizationId string `json:"payload"`
}

type RegisterEndpointEvent struct {
	Endpoint db.Endpoint `json:"payload"`
}

type RegisterEndpointOrganizationEvent struct {
	EndpointOrganization db.EndpointOrganization `json:"payload"`
}

type jsonEvent struct {
	EventType     string                 `json:"type"`
	EventIssuedAt time.Time              `json:"issuedAt"`
	EventPayload  map[string]interface{} `json:"payload"`
	data          []byte
}

func EventFromJson(data []byte) (Event, error) {
	e := jsonEvent{}
	err := json.Unmarshal(data, &e)
	e.data = data
	return e, err
}

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
