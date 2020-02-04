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
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	// RegisterEndpoint event type
	RegisterEndpoint EventType = "RegisterEndpointEvent"
	// RegisterVendor event type
	RegisterVendor EventType = "RegisterVendorEvent"
	// VendorClaim event type
	VendorClaim EventType = "VendorClaimEvent"
)

// ErrMissingEventType is given when the event being unmarshalled has no type attribute.
var ErrMissingEventType = errors.New("unmarshalling error: missing event type")

var eventTypes []EventType

func init() {
	eventTypes = []EventType{
		RegisterEndpoint,
		RegisterVendor,
		VendorClaim,
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

// Identifier defines component schema for Identifier.
type Identifier string

// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Organization Identifier `json:"organization"`
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// RegisterVendorEvent event
type RegisterVendorEvent struct {
	Identifier Identifier `json:"identifier"`
	Name       string     `json:"name"`
}

// VendorClaimEvent event
type VendorClaimEvent struct {
	VendorIdentifier Identifier    `json:"vendorIdentifier"`
	OrgIdentifier    Identifier    `json:"orgIdentifier"`
	OrgName          string        `json:"orgName"`
	OrgKeys          []interface{} `json:"orgKeys,omitempty"`
	Start            time.Time     `json:"start"`
	End              time.Time     `json:"end,omitempty"`
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
	type e struct {
		jsonEvent
		P interface{} `json:"payload"`
	}
	event := e{
		jsonEvent: jsonEvent{
			EventType:     string(eventType),
			EventIssuedAt: time.Now(),
		},
		P: payload,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	event.data = data
	return event, nil
}

func (j jsonEvent) IssuedAt() time.Time {
	return j.EventIssuedAt
}

func (j jsonEvent) Type() EventType {
	return EventType(j.EventType)
}

func (j jsonEvent) Unmarshal(out interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(j.data))
	// Look for "payload" field
	for ; ; {
		token, err := decoder.Token()
		if err == io.EOF {
			return errors.New("event has no payload")
		}
		if err != nil {
			return err
		}
		str, ok := token.(string)
		if ok && str == "payload" {
			break
		}
	}
	return decoder.Decode(&out)
}

func (j jsonEvent) Marshal() []byte {
	return j.data
}
