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
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"time"
)

// EventRef is a reference to an event
type Ref []byte

func (r *Ref) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if bytes, err := hex.DecodeString(str); err != nil {
		return err
	} else {
		*r = bytes
	}
	return nil
}

func (r Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r Ref) IsZero() bool {
	return len(r) == 0
}

func (r Ref) String() string {
	return hex.EncodeToString(r)
}

// Version
type Version int

const currentEventVersion Version = 1

// Event defines an event which can be (un)marshalled.
type Event interface {
	Type() EventType
	IssuedAt() time.Time
	// Version holds the version of the event, which can be used for differentiate processing/ignoring legacy events
	Version() Version
	// Ref holds the reference to the current event
	Ref() Ref
	// PreviousRef holds the reference to the previous event
	PreviousRef() Ref
	Unmarshal(out interface{}) error
	Marshal() []byte
	Signature() []byte
	Sign(signFn func([]byte) ([]byte, error)) error
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
	EventVersion     Version          `json:"version"`
	EventType        string           `json:"type"`
	EventIssuedAt    time.Time        `json:"issuedAt"`
	ThisEventRef     Ref              `json:"ref,omitempty"`
	PreviousEvent    Ref              `json:"prev,omitempty"`
	JWS              string           `json:"jws,omitempty"`
	EventPayload     interface{}      `json:"payload,omitempty"`
	signatureDetails SignatureDetails `json:"-"`
}

func (j jsonEvent) Version() Version {
	return j.EventVersion
}

func (j jsonEvent) Ref() Ref {
	// Make sure ThisEventRef is not set since it should included in the hash. This can't mutate the struct itself,
	// since the struct is passed by value to this function, not by reference.
	j.ThisEventRef = nil
	eventJson, err := json.Marshal(j)
	if err != nil {
		// This should never happen
		panic(err)
	}
	sum := sha1.Sum(canonicalizeJSON(eventJson))
	return sum[:]
}

func (j *jsonEvent) PreviousRef() Ref {
	return j.PreviousEvent
}

func (j *jsonEvent) Sign(signFn func([]byte) ([]byte, error)) error {
	payload, err := json.Marshal(j.EventPayload)
	if err != nil {
		return err
	}
	signature, err := signFn(canonicalizeJSON(payload))
	if err != nil {
		return err
	}
	j.JWS = string(signature)
	return nil
}

func canonicalizeJSON(input []byte) []byte {
	// TODO: When actually canonicalized, signature verification should be made backwards compatible since there are
	// signatures of non-canonicalized JSON. This is the case when event.version = 0 (and maybe 1 if canonicalization is introduced in version 1)
	return input
}

// EventFromJSON unmarshals an event. If the event can't be unmarshalled, an error is returned.
func EventFromJSON(data []byte) (Event, error) {
	e := jsonEvent{}
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	if e.EventType == "" {
		return nil, ErrMissingEventType
	}
	return &e, nil
}

// CreateEvent creates an event of the given type and the provided payload.
func CreateEvent(eventType EventType, payload interface{}, previousEvent Ref) Event {
	return &jsonEvent{
		EventVersion:  currentEventVersion,
		EventType:     string(eventType),
		PreviousEvent: previousEvent,
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
		// https://github.com/nuts-foundation/nuts-registry/issues/84
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
	// Marshal a copy since Ref should be calculated
	var e = j
	e.ThisEventRef = e.Ref()
	data, _ := json.MarshalIndent(e, "", "  ")
	return data
}

func (j jsonEvent) Signature() []byte {
	if j.JWS == "" {
		return nil
	}
	return []byte(j.JWS)
}
