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
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"time"
)

var eventFileRegex *regexp.Regexp = nil

const eventFileFormat = "(\\d{17})-([a-zA-Z]+)\\.json"

func init() {
	r, err := regexp.Compile(eventFileFormat)
	if err != nil {
		panic(err)
	}
	eventFileRegex = r
}

type EventSystem interface {
	// RegisterEventHandler registers an event handler for the given type, which will be called when the an event of this
	// type is received.
	RegisterEventHandler(eventType EventType, handler EventHandler)
	ProcessEvent(event Event) error
	PublishEvent(event Event) error
	LoadAndApplyEvents(location string) error
}

type EventHandler func(Event) error

type eventSystem struct {
	eventHandlers map[EventType]EventHandler
	// lastLoadedEvent contains the identifier of the last event that was loaded. It is used to keep track from
	// what event to resume when the events are reloaded (from disk)
	lastLoadedEvent time.Time
	mux             sync.Mutex
}

func NewEventSystem() *eventSystem {
	return &eventSystem{eventHandlers: make(map[EventType]EventHandler, 0)}
}

func (system *eventSystem) RegisterEventHandler(eventType EventType, handler EventHandler) {
	system.eventHandlers[eventType] = handler
}

func (system *eventSystem) ProcessEvent(event Event) error {
	if !IsEventType(event.Type()) {
		return fmt.Errorf("unknown event type: %s", event.Type())
	}
	handler := system.eventHandlers[event.Type()]
	if handler == nil {
		return fmt.Errorf("no handler registered for event (type = %s), handlers are: %v", event.Type(), system.eventHandlers)
	}
	err := handler(event)
	if err == nil {
		system.lastLoadedEvent = event.IssuedAt()
		logrus.Infof("Processed event: %v - %s", event.IssuedAt(), event.Type())
	}
	return err
}

func (system eventSystem) PublishEvent(event Event) error {
	// TODO: In future we'll publish the event to the mesh network here
	return system.ProcessEvent(event)
}

// Load the db files from the datadir
func (system *eventSystem) LoadAndApplyEvents(location string) error {
	system.mux.Lock()
	defer system.mux.Unlock()
	err := validateLocation(location)
	if err != nil {
		return err
	}

	type fileEvent struct {
		file  string
		event Event
	}
	events := make([]fileEvent, 0)
	entries, err := ioutil.ReadDir(location)
	for i := system.findStartIndex(entries); i < len(entries); i++ {
		entry := entries[i]
		logrus.Infof("Resuming from %s", entry.Name())
		if entry.IsDir() {
			continue
		}

		matches := eventFileRegex.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			return errors.New(fmt.Sprintf("file does not match event file name format (file = %s, expected format = %s)", entry.Name(), eventFileFormat))
		}
		event, err := readEvent(normalizeLocation(location, entry.Name()), matches[1])
		if err != nil {
			return err
		}
		events = append(events, fileEvent{
			file:  entry.Name(),
			event: event,
		})
	}
	logrus.Info("Applying events...")
	for _, e := range events {
		err := system.ProcessEvent(e.event)
		if err != nil {
			return errors2.Wrap(err, fmt.Sprintf("error while applying event (event = %s)", e.file))
		}
	}
	return nil
}

func (system eventSystem) findStartIndex(entries []os.FileInfo) int {
	for index, entry := range entries {
		timestamp, err := parseTimestamp(entry.Name()[:17])
		if err == nil {
			if timestamp.After(system.lastLoadedEvent) {
				return index
			}
		}
	}
	return 0
}

func parseTimestamp(timestamp string) (time.Time, error) {
	t, err := time.Parse("20060102150405.000", timestamp[0:14]+"."+timestamp[14:])
	if err != nil {
		return time.Time{}, errors2.Wrap(err, "event timestamp does not match required pattern (yyyyMMddHHmmssmmm)")
	}
	return t, nil
}

func readEvent(file string, timestamp string) (Event, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	event, err := EventFromJson(data)
	if err != nil {
		return nil, err
	}
	if !event.IssuedAt().IsZero() {
		return nil, fmt.Errorf("event from file should not contain issuedAt, since it's derived from the file name")
	}
	t, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, err
	}
	je := event.(jsonEvent)
	je.EventIssuedAt = t
	return je, nil
}
