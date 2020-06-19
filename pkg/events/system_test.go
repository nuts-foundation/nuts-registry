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
	"crypto/rand"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnknownEventType(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem()
	system.Configure(repo.Directory + "/events")
	input := `{
		"type": "non-existing"
	}`
	event, err := EventFromJSON([]byte(input))
	if !assert.NoError(t, err) {
		return
	}
	err = system.ProcessEvent(event)
	assert.EqualError(t, err, "unknown event type: non-existing")
}

func TestNoEventHandler(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem("some-type")
	system.Configure(repo.Directory + "/events")
	input := "{\"type\":\"some-type\"}"
	event, err := EventFromJSON([]byte(input))
	if !assert.NoError(t, err) {
		return
	}
	err = system.ProcessEvent(event)
	assert.EqualError(t, err, "no handlers registered for event (type = some-type), handlers are: map[]")
}

func TestLoadEvents(t *testing.T) {
	system := NewEventSystem("RegisterVendorEvent", "VendorClaimEvent", "RegisterEndpointEvent")
	vendorsCreated := 0
	system.RegisterEventHandler("RegisterVendorEvent", func(e Event, _ EventLookup) error {
		vendorsCreated++
		return nil
	})
	organizationsCreated := 0
	system.RegisterEventHandler("VendorClaimEvent", func(e Event, _ EventLookup) error {
		organizationsCreated++
		return nil
	})
	endpointsCreated := 0
	system.RegisterEventHandler("RegisterEndpointEvent", func(e Event, _ EventLookup) error {
		endpointsCreated++
		return nil
	})

	assertEventsHandled := func(vc int, oc int, ec int) {
		assert.Equal(t, vc, vendorsCreated, "unexpected number of events for: RegisterVendor")
		assert.Equal(t, oc, organizationsCreated, "unexpected number of events for: VendorClaim")
		assert.Equal(t, ec, endpointsCreated, "unexpected number of events for: RegisterEndpoint")
	}

	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system.Configure(repo.Directory + "/events")

	const sourceDir = "../../test_data/valid_files"
	t.Run("All fresh system state, all events should be loaded", func(t *testing.T) {
		if !assert.NoError(t, repo.ImportDir(sourceDir)) {
			return
		}
		err := system.LoadAndApplyEvents()
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(1, 2, 2)
	})

	t.Run("New event file, should trigger an incremental change", func(t *testing.T) {
		err := repo.ImportFileAs(filepath.Join(sourceDir, "events/20200123091400005-RegisterEndpointEvent.json"), "events/20210123091400005-RegisterEndpointEvent.json")
		if !assert.NoError(t, err) {
			return
		}
		err = system.LoadAndApplyEvents()
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(1, 2, 3)
	})

	t.Run("No incremental change", func(t *testing.T) {
		err := system.LoadAndApplyEvents()
		if !assert.NoError(t, err) {
			return
		}
		assertEventsHandled(1, 2, 3)
	})

	t.Run("Added non-JSON file", func(t *testing.T) {
		eventSystem := system.(*diskEventSystem)
		eventSystem.lastLoadedEvent = time.Time{}
		err := repo.ImportFileAs("system_test.go", "events/system_test.go")
		if !assert.NoError(t, err) {
			return
		}
		err = system.LoadAndApplyEvents()
		if !assert.NoError(t, err) {
			return
		}
	})
}

func TestSystemNotConfigured(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		system := NewEventSystem()
		err := system.PublishEvent(CreateEvent("RegisterVendorEvent", struct{}{}, nil))
		assert.EqualError(t, err, ErrEventSystemNotConfigured.Error())
	})
	t.Run("process", func(t *testing.T) {
		system := NewEventSystem()
		err := system.ProcessEvent(CreateEvent("RegisterVendorEvent", struct{}{}, nil))
		assert.EqualError(t, err, ErrEventSystemNotConfigured.Error())
	})
	t.Run("LoadAndApply", func(t *testing.T) {
		system := NewEventSystem()
		err := system.LoadAndApplyEvents()
		assert.EqualError(t, err, ErrEventSystemNotConfigured.Error())
	})
}

func TestLoadEventsInvalidJson(t *testing.T) {
	repo, err := test.NewTestRepoFrom(t.Name(), "../../test_data/invalid_files")
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem()
	system.Configure(repo.Directory + "/events")
	err = system.LoadAndApplyEvents()
	assert.EqualError(t, err, "error reading event: 20200123091400001-InvalidJson.json: unable to parse event JSON: invalid character '{' looking for beginning of object key string")
}

func TestLoadEventsEmptyFile(t *testing.T) {
	repo, err := test.NewTestRepoFrom(t.Name(), "../../test_data/empty_files")
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem()
	system.Configure(repo.Directory + "/events")
	err = system.LoadAndApplyEvents()
	assert.EqualError(t, err, "error reading event: 20200123091400001-EmptyFile.json: unable to parse event JSON: unexpected end of JSON input")
}

func TestParseTimestamp(t *testing.T) {
	t.Run("Timestamp OK", func(t *testing.T) {
		timestamp, err := parseTimestamp("20200123091400001")
		assert.Equal(t, "2020-01-23 09:14:00.001 +0000 UTC", timestamp.String())
		assert.NoError(t, err)
	})
	t.Run("Timestamp has invalid length", func(t *testing.T) {
		timestamp, err := parseTimestamp("asdasd")
		assert.True(t, timestamp.IsZero())
		assert.Error(t, err)
	})
	t.Run("Timestamp has invalid characters", func(t *testing.T) {
		timestamp, err := parseTimestamp("a2345678901234567")
		assert.True(t, timestamp.IsZero())
		assert.Error(t, err)
	})

}

func TestEventLookup(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem("evt")
	t.Run("ok - Get is delegated", func(t *testing.T) {
		assert.Nil(t, system.Get(Ref([]byte{1, 2, 3})))
	})
	t.Run("ok - FindLastEvent is delegated", func(t *testing.T) {
		event, err := system.FindLastEvent(func(event Event) bool {
			return true
		})
		assert.NoError(t, err)
		assert.Nil(t, event)
	})
}

func TestPublishEvents(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	system := NewEventSystem("evt")
	system.Configure(repo.Directory)
	called := 0
	system.RegisterEventHandler("evt", func(event Event, _ EventLookup) error {
		called++
		return nil
	})
	t.Run("assert event handler is called", func(t *testing.T) {
		err = system.PublishEvent(&jsonEvent{EventType: "evt"})
		assert.NoError(t, err)
		assert.Equal(t, called, 1)
	})
	t.Run("assert event file is written and valid JSON", func(t *testing.T) {
		repo.Cleanup()
		os.MkdirAll(repo.Directory, os.ModePerm)
		dirEntriesBeforePublish, err := ioutil.ReadDir(repo.Directory)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, dirEntriesBeforePublish, 0, "directory empty") {
			return
		}
		event := CreateEvent("evt", struct{}{}, nil)
		err = system.PublishEvent(event)
		if !assert.NoError(t, err) {
			return
		}
		dirEntriesAfterPublish, err := ioutil.ReadDir(repo.Directory)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, dirEntriesAfterPublish, 1, "directory not empty") {
			return
		}
		data, err := ioutil.ReadFile(filepath.Join(repo.Directory, dirEntriesAfterPublish[0].Name()))
		if !assert.NoError(t, err) {
			return
		}
		_, err = EventFromJSON(data)
		assert.NoError(t, err)
	})
}

func Test_readEvent(t *testing.T) {
	t.Run("v0", func(t *testing.T) {
		event, err := readEvent("../../test_data/valid_files/events/20200123091400001-RegisterVendorEvent.json", "20200123091400001")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "5e3afa352f26033bd62ac8d6843ccae0f5334296", event.Ref().String())
	})
	t.Run("v1", func(t *testing.T) {
		// From v1 on event contains issuedAt which should be used instead of the file name
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		defer repo.Cleanup()

		event := CreateEvent(eventType, eventPayload, nil)
		(event.(*jsonEvent)).EventIssuedAt = time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)
		eventFilePath := normalizeLocation(repo.Directory, SuggestEventFileName(event))
		ioutil.WriteFile(eventFilePath, event.Marshal(), os.ModePerm)
		// Random timestamp = random string, makes sure it isn't parsable, which is only performed when overriding event.IssuedAt (because it's not set in source JSON).
		randomTimestamp := make([]byte, 20)
		rand.Read(randomTimestamp)
		event, err = readEvent(eventFilePath, string(randomTimestamp))
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "00fc6cf20e8105ef53a90914aab5414e8ed0059d", event.Ref().String())
	})
	t.Run("error - can't parse timestamp", func(t *testing.T) {
		event, err := readEvent("../../test_data/valid_files/events/20200123091400001-RegisterVendorEvent.json", "foobar")
		assert.EqualError(t, err, "event timestamp does not match required pattern (yyyyMMddHHmmssmmm)")
		assert.Nil(t, event)
	})
	t.Run("error - can't read file", func(t *testing.T) {
		event, err := readEvent("asadasd", "asdsd")
		assert.EqualError(t, err, "unable to parse event file: open asadasd: no such file or directory")
		assert.Nil(t, event)
	})
}
