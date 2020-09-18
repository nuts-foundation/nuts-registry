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
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/log"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var eventFileRegex *regexp.Regexp

// ErrInvalidTimestamp is returned when a timestamp does not match the required pattern
var ErrInvalidTimestamp = errors.New("event timestamp does not match required pattern (yyyyMMddHHmmssmmm)")

// ErrEventSystemNotConfigured is returned when the event system is used but wasn't configured by calling Configure()
var ErrEventSystemNotConfigured = errors.New("the event system hasn't been configured, please call Configure()")

// ErrEventNotSigned is returned when the event is not signed
var ErrEventNotSigned = errors.New("the event is not signed")

const eventTimestampLayout = "20060102150405.000"
const eventFileFormat = "(\\d{17})-([a-zA-Z]+)\\.json"

func init() {
	r, err := regexp.Compile(eventFileFormat)
	if err != nil {
		panic(err)
	}
	eventFileRegex = r
}

// EventSystem is meant for registering and handling events.
type EventSystem interface {
	// RegisterEventHandler registers an event handler for the given type, which will be called when the an event of this
	// type is received.
	RegisterEventHandler(eventType EventType, handler EventHandler)
	ProcessEvent(event Event) error
	PublishEvent(event Event) error
	LoadAndApplyEvents() error
	Configure(location string) error
	Diagnostics() []core.DiagnosticResult
	EventLookup
}

// EventRegistrar is a function to register an event
type EventRegistrar func(EventType, EventHandler)

// EventHandler handles an event of a specific type.
type EventHandler func(Event, EventLookup) error

// EventMatcher defines a matching function for events. The function should return true if the event matches, otherwise false.
type EventMatcher func(Event) bool

// JwsVerifier defines a verification delegate for JWS'.
type JwsVerifier func(signature []byte, signingTime time.Time, verifier cert.Verifier) ([]byte, error)

type diskEventSystem struct {
	eventHandlers map[EventType][]EventHandler
	eventTypes    []EventType
	location      string
	lut           *eventLookupTable
	// eventsToBeRetried holds events which should be retried since it failed previously.
	eventsToBeRetried map[string]Event
}

// NewEventSystem creates and initializes a new event system.
func NewEventSystem(eventTypes ...EventType) EventSystem {
	return &diskEventSystem{
		eventTypes:        eventTypes,
		eventHandlers:     make(map[EventType][]EventHandler, 0),
		lut:               newEventLookupTable(),
		eventsToBeRetried: make(map[string]Event),
	}
}

func (system *diskEventSystem) Configure(location string) error {
	system.location = location
	return validateLocation(system.location)
}

func (system *diskEventSystem) RegisterEventHandler(eventType EventType, handler EventHandler) {
	system.eventHandlers[eventType] = append(system.eventHandlers[eventType], handler)
}

func (system *diskEventSystem) Diagnostics() []core.DiagnosticResult {
	return []core.DiagnosticResult{
		&core.GenericDiagnosticResult{
			Title:   "Number of events to be retried",
			Outcome: fmt.Sprintf("%d", len(system.eventsToBeRetried)),
		},
	}
}

// isEventType checks whether the given type is supported.
func (system diskEventSystem) isEventType(eventType EventType) bool {
	for _, actual := range system.eventTypes {
		if actual == eventType {
			return true
		}
	}
	return false
}

func (system *diskEventSystem) ProcessEvent(event Event) error {
	if err := system.assertConfigured(); err != nil {
		return err
	}
	if !system.isEventType(event.Type()) {
		return fmt.Errorf("unknown event type: %s", event.Type())
	}
	// If this event has already been processed, we can skip it
	if system.lut.Get(event.Ref()) != nil {
		return nil
	}
	// If there is a previous event which hasn't been processed yet, we set it aside to be processed later.
	if !event.PreviousRef().IsZero() && system.lut.Get(event.PreviousRef()) == nil {
		log.Logger().Infof("Event %s refers to previous event %s which hasn't been processed yet, setting it aside.", event.Ref(), event.PreviousRef())
		system.eventsToBeRetried[event.Ref().String()] = event
		return nil
	}
	err := system.processEvent(event)
	// There might be unprocessed (received out of order) events that depend on this event, so we retry events which failed earlier
	system.retryEvents()
	return err
}

func (system *diskEventSystem) retryEvents() {
	events := make([]Event, len(system.eventsToBeRetried))
	i := 0
	for _, event := range system.eventsToBeRetried {
		events[i] = event
		i++
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].IssuedAt().Before(events[j].IssuedAt())
	})
	for _, event := range events {
		if err := system.processEvent(event); err != nil {
			log.Logger().Debugf("Error while processing set-aside event %s, will retry later: %v", event.Ref(), err)
		}
	}
}

func (system *diskEventSystem) processEvent(event Event) error {
	handlers := system.eventHandlers[event.Type()]
	if handlers == nil {
		return fmt.Errorf("no handlers registered for event (type = %s), handlers are: %v", event.Type(), system.eventHandlers)
	}
	for _, handler := range handlers {
		if err := handler(event, system.lut); err != nil {
			log.Logger().Warnf("Error while processing event %s, event will set aside to be processed later: %v", event.Ref(), err)
			system.eventsToBeRetried[event.Ref().String()] = event
			return err
		}
	}
	if err := system.lut.register(event); err != nil {
		return err
	}
	logrus.WithFields(map[string]interface{}{
		"ref":      event.Ref(),
		"prev":     event.PreviousRef(),
		"type":     event.Type(),
		"issuedAt": event.IssuedAt(),
	}).Info("Event processed")
	delete(system.eventsToBeRetried, event.Ref().String())
	return nil
}

func (system *diskEventSystem) PublishEvent(event Event) error {
	if err := system.assertConfigured(); err != nil {
		return err
	}
	if err := system.ProcessEvent(event); err != nil {
		return err
	}

	eventFileName := SuggestEventFileName(event)
	err := ioutil.WriteFile(normalizeLocation(system.location, eventFileName), event.Marshal(), os.ModePerm)
	if err != nil {
		return errors2.Wrap(err, "event processed, but enable to save it to disk")
	}
	return nil
}

func (system diskEventSystem) Get(ref Ref) Event {
	return system.lut.Get(ref)
}

func (system diskEventSystem) FindLastEvent(matcher EventMatcher) (Event, error) {
	return system.lut.FindLastEvent(matcher)
}

// Load the db files from the datadir
func (system *diskEventSystem) LoadAndApplyEvents() error {
	if err := system.assertConfigured(); err != nil {
		return err
	}
	entries, err := ioutil.ReadDir(system.location)
	if err != nil {
		return err
	}
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if !isJSONFile(entry) {
			continue
		}
		logrus.Debugf("Parsing event: %s", entry.Name())

		matches := eventFileRegex.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			return fmt.Errorf("file does not match event file name format (file = %s, expected format = %s)", entry.Name(), eventFileFormat)
		}
		event, err := readEvent(normalizeLocation(system.location, entry.Name()), matches[1])
		if err != nil {
			return errors2.Wrapf(err, "error reading event: %s", entry.Name())
		}
		if err := system.ProcessEvent(event); err != nil {
			return errors2.Wrap(err, fmt.Sprintf("error while applying event (event = %s)", entry.Name()))
		}
	}
	return nil
}

// SuggestEventFileName suggests a file name for a event, when writing that event to disk.
func SuggestEventFileName(event Event) string {
	return strings.Replace(event.IssuedAt().UTC().Format(eventTimestampLayout), ".", "", 1) + "-" + string(event.Type()) + ".json"
}

func (system diskEventSystem) assertConfigured() error {
	if system.location == "" {
		return ErrEventSystemNotConfigured
	}
	return nil
}

type TrustStore interface {
	cert.Verifier
	RegisterEventHandlers(func(EventType, EventHandler))
}

func isJSONFile(file os.FileInfo) bool {
	return !file.IsDir() && strings.HasSuffix(file.Name(), ".json")
}

func parseTimestamp(timestamp string) (time.Time, error) {
	if len(timestamp) != 17 {
		return time.Time{}, ErrInvalidTimestamp
	}
	t, err := time.Parse(eventTimestampLayout, timestamp[0:14]+"."+timestamp[14:])
	if err != nil {
		return time.Time{}, ErrInvalidTimestamp
	}
	return t, nil
}

func readEvent(file string, timestamp string) (Event, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to parse event file")
	}
	event, err := EventFromJSON(data)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to parse event JSON")
	}
	je := event.(*jsonEvent)
	if je.EventIssuedAt.IsZero() {
		t, err := parseTimestamp(timestamp)
		if err != nil {
			return nil, err
		}
		je.EventIssuedAt = t
	}
	return je, nil
}
