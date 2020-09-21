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
)

type EventLookup interface {
	// Get retrieves the event specified the reference. If the event doesn't exist, nil is returned.
	Get(ref Ref) Event
	// FindLastEvent finds the last event in the event path which matches the specified matcher. If there are multiple
	// event paths that match, an error is returned. If no events match, nil is returned.
	FindLastEvent(matcher EventMatcher) (Event, error)
}

type eventLookupTable struct {
	// refs contains all event references from parent to child (given that B refers to previous event A; {A -> B})
	refs map[Event]Event
	// entries contains all events indexed by their ref {(ref(A) -> A, ref(B) -> B}
	entries map[string]Event
}

func newEventLookupTable() *eventLookupTable {
	return &eventLookupTable{refs: make(map[Event]Event, 0), entries: make(map[string]Event, 0)}
}

func (r eventLookupTable) Get(ref Ref) Event {
	return r.entries[ref.String()]
}

func (r eventLookupTable) FindLastEvent(matcher EventMatcher) (Event, error) {
	var matches = make(map[Event]bool)
	for _, event := range r.entries {
		if matcher(event) {
			matches[event] = false
		}
	}
	// Find paths for matching events
	var paths = make([][]Event, 0)
	for event, matched := range matches {
		if matched {
			// Event already in a previously found path
			continue
		}
		path := r.findPath(event)
		// Mark all events in path as matched
		for _, eventInPath := range path {
			matches[eventInPath] = true
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil, nil
	}
	if len(paths) > 1 {
		return nil, errors.New("multiple event paths match")
	}
	return paths[0][len(paths[0])-1], nil
}

func (r eventLookupTable) findPath(event Event) []Event {
	var path []Event
	current := r.findHeadOfPath(event)
	path = append(path, current)
	for ; r.refs[current] != nil; {
		current = r.refs[current]
		path = append(path, current)
	}
	return path
}

// findHeadOfPath finds the head (first event) of the event path (events referring to previous events). Returns an error
// if there the path is broken (missing events).
func (r eventLookupTable) findHeadOfPath(event Event) Event {
	if !event.PreviousRef().IsZero() {
		// This event refers to another event
		// Recursion should be safe for now, since we don't have paths with tens of thousands of events yet.
		return r.findHeadOfPath(r.entries[event.PreviousRef().String()])
	}
	// This is the first event in its path
	return event
}

func (r *eventLookupTable) register(event Event) error {
	prevRef := event.PreviousRef()
	if !prevRef.IsZero() {
		// Event refers to a previous event, validate that:
		// - referred event exists,
		// - referred event isn't already referred to,
		// - referred event is of same type
		prevEvent := r.entries[prevRef.String()]
		if prevEvent == nil {
			return fmt.Errorf("previous event not found: %s", prevRef)
		}
		if r.refs[prevEvent] != nil {
			return fmt.Errorf("previous event (%s) already referred to", prevRef)
		}
		if prevEvent.Type() != event.Type() {
			return fmt.Errorf("previous event type differs (pref: %s=%s, this: %s=%s)", prevRef, prevEvent.Type(), event.Ref(), event.Type())
		}
		r.refs[prevEvent] = event
	}
	r.entries[event.Ref().String()] = event
	return nil
}
