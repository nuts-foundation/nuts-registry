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

package events

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/nuts-foundation/nuts-registry/logging"
	"github.com/sirupsen/logrus"
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

func (r Ref) Equal(other Ref) bool {
	return bytes.Equal(r, other)
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
	SignatureDetails() SignatureDetails
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
	cachedData       []byte
}

func (j jsonEvent) SignatureDetails() SignatureDetails {
	return j.signatureDetails
}

func (j jsonEvent) Version() Version {
	return j.EventVersion
}

func (j jsonEvent) Ref() Ref {
	// Make sure ThisEventRef is not set since it should included in the hash. This can't mutate the struct itself,
	// since the struct is passed by value to this function, not by reference.
	eventAsMap := make(map[string]interface{})
	eventAsJSON, _ := json.Marshal(j)
	_ = json.Unmarshal(eventAsJSON, &eventAsMap)
	// Make a list of keys to be included in the ref
	var includeKeys = []string{"issuedAt", "type", "jws", "payload"}
	if j.Version() >= 1 {
		includeKeys = append(includeKeys, "prev", "version")
	}
	// Remove all fields from the map that shouldn't be in there for this version
	for key, _ := range eventAsMap {
		included := false
		for _, k := range includeKeys {
			if k == key {
				included = true
				break
			}
		}
		if !included {
			delete(eventAsMap, key)
		}
	}
	strippedJSON, _ := json.Marshal(eventAsMap)
	canonicalizedJSON, err := canonicalizeJSON(strippedJSON)
	if logging.Log().Logger.IsLevelEnabled(logrus.DebugLevel) {
		logging.Log().WithFields(map[string]interface{}{
			"event":         string(eventAsJSON),
			"canonicalized": string(canonicalizedJSON),
		}).Trace("Calculating event ref")
	}
	if err != nil {
		// This should never happen
		panic(err)
	}
	sum := sha1.Sum(canonicalizedJSON)
	return sum[:]
}

func (j jsonEvent) PreviousRef() Ref {
	return j.PreviousEvent
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
	j.cachedData = nil // Reset cached data since we modified the event contents
	return nil
}

func canonicalizeJSON(input []byte) ([]byte, error) {
	return jsoncanonicalizer.Transform(input)
}

// EventsFromJSON is the same as EventsFromJSON except that it unmarshals a list of events rather than a single event.
func EventsFromJSON(data []byte) ([]Event, error) {
	var events []jsonEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	var result = make([]Event, len(events))
	for i, _ := range events {
		result[i] = &events[i]
	}
	return result, nil
}

// EventFromJSON unmarshals an event. If the event can't be unmarshalled, an error is returned.
func EventFromJSON(data []byte) (Event, error) {
	e := jsonEvent{}
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	e.cachedData = data
	if err := validateEvent(e); err != nil {
		return nil, err
	}
	return &e, nil
}

// EventFromJSONWithIssuedAt functions exactly as EventFromJSON but allows the caller to set issuedAt if it's not
// present in the event which is the case for v0 events. If it's present, the supplied issuedAt is ignored.
func EventFromJSONWithIssuedAt(data []byte, issuedAt time.Time) (Event, error) {
	if event, err := EventFromJSON(data); err != nil {
		return nil, err
	} else {
		jEvent := event.(*jsonEvent)
		if jEvent.IssuedAt().IsZero() {
			jEvent.EventIssuedAt = issuedAt
		}
		return jEvent, nil
	}
}

func validateEvent(evt jsonEvent) error {
	if evt.EventType == "" {
		return ErrMissingEventType
	}
	if !evt.ThisEventRef.IsZero() {
		actualRef := evt.Ref()
		if !evt.ThisEventRef.Equal(actualRef) {
			return fmt.Errorf("event ref is invalid (specified: %s, actual: %s)", evt.ThisEventRef, actualRef)
		}
	}
	return nil
}

// CreateEvent creates an event of the given type and the provided payload.
func CreateEvent(eventType EventType, payload interface{}, previousEvent Ref) Event {
	return &jsonEvent{
		EventVersion:  currentEventVersion,
		EventType:     string(eventType),
		PreviousEvent: previousEvent,
		EventIssuedAt: time.Now().UTC(),
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
		for {
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
	if j.cachedData == nil {
		var e = j
		e.ThisEventRef = e.Ref()
		j.cachedData, _ = json.MarshalIndent(e, "", "  ")
	}
	return j.cachedData
}

func (j jsonEvent) Signature() []byte {
	if j.JWS == "" {
		return nil
	}
	return []byte(j.JWS)
}
