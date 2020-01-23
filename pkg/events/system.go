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
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
)

var eventFileRegex *regexp.Regexp = nil

const eventFileFormat = "\\d{17}-([a-zA-Z]+)\\.json"

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
	RegisterEventHandler(eventType reflect.Type, handler EventHandler)
	RegisteredEventTypes() []reflect.Type
	ProcessEvent(event interface{}) error
	PublishEvent(event interface{}) error
	LoadAndApplyEvents(location string) error
}

type EventHandler func(interface{}) error

type eventSystem struct {
	eventHandlers map[reflect.Type]EventHandler
	// lastLoadedEvent contains the file name of the last event that was loaded. It is used
	lastLoadedEvent string
}

func NewEventSystem() *eventSystem {
	return &eventSystem{eventHandlers: make(map[reflect.Type]EventHandler, 0)}
}

func (system *eventSystem) RegisterEventHandler(eventType reflect.Type, handler EventHandler) {
	system.eventHandlers[eventType] = handler
}

func (system eventSystem) RegisteredEventTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(system.eventHandlers))
	for t, _ := range system.eventHandlers {
		types = append(types, t)
	}
	return types
}

func (system eventSystem) ProcessEvent(event interface{}) error {
	eventType := reflect.TypeOf(event)
	handler := system.eventHandlers[eventType]
	if handler == nil {
		return errors.New(fmt.Sprintf("no handler registered for event (type = %T), handlers are: %v", event, system.eventHandlers))
	}
	return handler(event)
}

func (system eventSystem) PublishEvent(event interface{}) error {
	if reflect.TypeOf(event).Kind() == reflect.Struct {
		// TODO: This is actually a pretty ugly API, so this should be fixed in future
		return errors.New("a struct was passed to PublishEvent, which should've been a pointer to the struct")
	}
	// TODO: In future we'll publish the event to the mesh network here
	return system.ProcessEvent(event)
}

// Load the db files from the datadir
func (system *eventSystem) LoadAndApplyEvents(location string) error {
	err := validateLocation(location)
	if err != nil {
		return err
	}

	type fileEvent struct {
		file  string
		event interface{}
	}
	events := make([]fileEvent, 0)
	entries, err := ioutil.ReadDir(location)
	for i := system.findStartIndex(entries); i < len(entries); i++ {
		entry := entries[i]
		if entry.IsDir() {
			continue
		}

		matches := eventFileRegex.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			return errors.New(fmt.Sprintf("file does not match event file format (file = %s, expected format = %s)", entry.Name(), eventFileFormat))
		}
		eventName := matches[1]
		resolvedEventType := system.resolveEventType(eventName)

		if resolvedEventType == nil {
			return errors.New(fmt.Sprintf("unsupported event type (type = %s, supported = %v)", eventName, system.RegisteredEventTypes()))
		}
		logrus.Debugf("Reading %s file: %s", resolvedEventType.Name(), entry.Name())
		event, err := readEvent(normalizeLocation(location, entry.Name()), resolvedEventType)
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
		system.lastLoadedEvent = e.file
		logrus.Infof("Applied event: %s", e.file)
	}
	return nil
}

func (system eventSystem) resolveEventType(eventName string) reflect.Type {
	for _, eventType := range system.RegisteredEventTypes() {
		if eventType.Elem().Name() == eventName {
			return eventType
		}
	}
	return nil
}

func (system eventSystem) findStartIndex(entries []os.FileInfo) int {
	for index, entry := range entries {
		if entry.Name() == system.lastLoadedEvent {
			return index
		}
	}
	return 0
}

func readEvent(file string, eventType reflect.Type) (interface{}, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return reflect.Value{}, err
	}
	event := reflect.New(eventType.Elem()).Interface()
	err = json.Unmarshal(data, event)
	if err != nil {
		return reflect.Value{}, err
	}
	return event, nil
}
