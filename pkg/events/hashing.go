package events

import (
	"fmt"
)

type eventLookupTable struct {
	// refs contains all event references from parent to child (given that B refers to previous event A; {A -> B})
	refs    map[Event]Event
	// entries contains all events indexed by their ref {(ref(A) -> A, ref(B) -> B}
	entries map[string]Event
}

func newEventLookupTable() *eventLookupTable {
	return &eventLookupTable{refs: make(map[Event]Event, 0), entries: make(map[string]Event, 0)}
}

func (r eventLookupTable) lookup(ref Ref) Event {
	return r.entries[ref.String()]
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
