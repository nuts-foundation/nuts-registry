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
	"crypto/x509"
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
	Signature() []byte
	Sign(signFn func([]byte) ([]byte, error)) error
	SignatureDetails() SignatureDetails
	SetSignatureDetails(details SignatureDetails)
}

// SignatureDetails describes the properties of the signature that secures the event
type SignatureDetails struct {
	// Certificate contains the X.509 certificate that signed the event
	Certificate *x509.Certificate
	// Payload contains the event data that is protected by the signature
	Payload []byte
}

// UnmarshalPostProcessor allows to define custom logic that should be executed after unmarshalling
type UnmarshalPostProcessor interface {
	PostProcessUnmarshal(event Event) error
}

// EventType defines a supported type of event, which is used for executing the right handler.
type EventType string

// ErrMissingEventType is given when the event being unmarshalled has no type attribute.
var ErrMissingEventType = errors.New("unmarshalling error: missing event type")

type jsonEvent struct {
	EventType        string           `json:"type"`
	EventIssuedAt    time.Time        `json:"issuedAt"`
	JWS              string           `json:"jws,omitempty"`
	EventPayload     interface{}      `json:"payload,omitempty"`
	signatureDetails SignatureDetails `json:"-"`
}

func (j jsonEvent) SignatureDetails() SignatureDetails {
	return j.signatureDetails
}

func (j *jsonEvent) SetSignatureDetails(details SignatureDetails) {
	j.signatureDetails = details
}

func (j *jsonEvent) Sign(signFn func([]byte) ([]byte, error)) error {
	payload, err := json.Marshal(j.EventPayload)
	if err != nil {
		return err
	}
	signature, err := signFn(payload)
	if err != nil {
		return err
	}
	j.JWS = string(signature)
	return nil
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
	return &e, nil
}

// CreateEvent creates an event of the given type and the provided payload. If the event can't be created, an error is
// returned.
func CreateEvent(eventType EventType, payload interface{}) Event {
	return &jsonEvent{
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
	if j.signatureDetails.Payload == nil {
		// Backwards compatibility for events that aren't signed
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
	} else {
		if err := json.Unmarshal(j.signatureDetails.Payload, &out); err != nil {
			return err
		}
	}
	postProc, ok := out.(UnmarshalPostProcessor)
	if ok {
		if err := postProc.PostProcessUnmarshal(&j); err != nil {
			return err
		}
	}
	return nil
}

func (j jsonEvent) Marshal() []byte {
	data, _ := json.MarshalIndent(j, "", "  ")
	return data
}

func (j jsonEvent) Signature() []byte {
	return []byte(j.JWS)
}
