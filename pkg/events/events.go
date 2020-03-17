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

// unmarshalPostProcessor allows to define custom logic that should be executed after unmarshalling
type unmarshalPostProcessor interface {
	unmarshalPostProcess() error
}

// EventType defines a supported type of event, which is used for executing the right handler.
type EventType string

// ErrMissingEventType is given when the event being unmarshalled has no type attribute.
var ErrMissingEventType = errors.New("unmarshalling error: missing event type")

var eventTypes []EventType

// IsEventType checks whether the given type is supported.
func IsEventType(eventType EventType) bool {
	for _, actual := range eventTypes {
		if actual == eventType {
			return true
		}
	}
	return false
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
	err := decoder.Decode(&out)
	if err != nil {
		return err
	}
	postProc, ok := out.(unmarshalPostProcessor)
	if ok {
		if err := postProc.unmarshalPostProcess(); err != nil {
			return err
		}
	}
	return nil
}

func (j jsonEvent) Marshal() []byte {
	data, _ := json.Marshal(j)
	return data
}
