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
	Identifier Identifier    `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

// VendorClaimEvent event
type VendorClaimEvent struct {
	VendorIdentifier Identifier `json:"vendorIdentifier"`
	OrgIdentifier    Identifier `json:"orgIdentifier"`
	OrgName          string     `json:"orgName"`
	// OrgKeys is a list of JWKs which are used to
	// 1. encrypt data to be decrypted by the organization,
	// 2. sign consent JWTs,
	// 3. sign organization related events (e.g. endpoint registration).
	OrgKeys []interface{} `json:"orgKeys,omitempty"`
	Start   time.Time     `json:"start"`
	End     *time.Time    `json:"end,omitempty"`
}

type jsonEvent struct {
	EventType     string      `json:"type"`
	EventIssuedAt time.Time   `json:"issuedAt"`
	EventPayload  interface{} `json:"payload,omitempty"`
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
	return e, nil
}

// CreateEvent creates an event of the given type and the provided payload. If the event can't be created, an error is
// returned.
func CreateEvent(eventType EventType, payload interface{}) Event {
	return jsonEvent{
		EventType:     string(eventType),
		EventIssuedAt: time.Now(),
		EventPayload:  payload,
	}
}

func (j jsonEvent) IssuedAt() time.Time {
	return j.EventIssuedAt
}

func (j jsonEvent) Type() EventType {
	return EventType(j.EventType)
}

func (j jsonEvent) Unmarshal(out interface{}) error {
	data := j.Marshal()
	decoder := json.NewDecoder(bytes.NewReader(data))
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
	data, _ := json.Marshal(j)
	return data
}
