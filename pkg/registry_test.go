// +build !race

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

package pkg

import (
	"fmt"
	"github.com/labstack/gommon/random"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type ZipHandler struct {
}

func (h *ZipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// open zip file
	bytes, _ := ioutil.ReadFile("../test_data/valid_files.tar.gz")

	// random Etag
	w.Header().Add("ETag", random.String(8))

	// write bytes to w
	w.Write(bytes)
}

func TestRegistry_Instance(t *testing.T) {
	registry1 := RegistryInstance()
	registry2 := RegistryInstance()
	assert.Same(t, registry1, registry2)
	assert.NotNil(t, registry1.EventSystem)
}

func TestRegistry_Start(t *testing.T) {
	configureIdleTimeout()
	t.Run("Start with an incorrect configuration returns error", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				SyncMode: "unknown",
				Datadir:  ".",
			},
		}

		err := registry.Start()

		if err == nil {
			t.Error("Expected error, got nothing")
		}

		expected := "invalid syncMode: unknown"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got [%v]", expected, err)
		}
	})

	t.Run("Starting sets the file watcher", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				SyncMode: "fs",
				Datadir:  ".",
			},
		}

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(registry.closers) != 1 {
			t.Error("Expected watcher to be started")
		}
	})

	t.Run("Invalid datadir gives error on Start", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				SyncMode: "fs",
				Datadir:  ":",
			},
		}

		err := registry.Start()

		if err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("Shutdown stops the file watcher", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				SyncMode: "fs",
				Datadir:  ".",
			},
		}

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		// watcher delay
		time.Sleep(time.Millisecond * 100)

		if err := registry.Shutdown(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}
	})
}

func TestRegistry_Configure(t *testing.T) {
	configureIdleTimeout()
	t.Run("Configure loads the BD", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:    "server",
				Datadir: "../test_data/valid_files",
			},
			EventSystem: events.NewEventSystem(),
		}

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func TestRegistry_FileUpdate(t *testing.T) {
	cleanup()
	defer cleanup()
	configureIdleTimeout()

	t.Run("New files are loaded", func(t *testing.T) {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)

		wg := sync.WaitGroup{}
		wg.Add(1)

		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				Datadir:  "../tmp",
				SyncMode: "fs",
			},
			OnChange: func(registry *Registry) {
				wg.Done()
			},
			EventSystem: events.NewEventSystem(),
		}
		defer registry.Shutdown()

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(registry.Db.SearchOrganizations("")) != 0 {
			t.Error("Expected empty db")
		}

		// copy valid files
		copyDir("../test_data/valid_files/events", "../tmp/events")

		wg.Wait()

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func TestRegistry_GithubUpdate(t *testing.T) {
	cleanup()
	defer cleanup()
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	configureIdleTimeout()

	t.Run("New files are downloaded", func(t *testing.T) {
		handler := &ZipHandler{}
		server := httptest.NewServer(handler)
		defer server.Close()

		wg := sync.WaitGroup{}
		wg.Add(1)

		os.Mkdir("../tmp", os.ModePerm)
		copyDir("../test_data/all_empty_files", "../tmp")

		registry := Registry{
			Config: RegistryConfig{
				Mode:         "server",
				Datadir:      "../tmp",
				SyncMode:     "github",
				SyncAddress:  server.URL,
				SyncInterval: 60,
			},
			OnChange: func(registry *Registry) {
				println("EVENT")
				wg.Done()
			},
			EventSystem: events.NewEventSystem(),
		}
		defer registry.Shutdown()

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		// wait for download
		wg.Wait()

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func TestRegistry_EventsOnUpdate(t *testing.T) {
	t.Run("Check event emitted: register organization", func(t *testing.T) {
		eventSystem := &MockEventSystem{Events: []events.Event{}}
		registry := Registry{
			Config: RegistryConfig{
			},
			EventSystem: eventSystem,
		}
		err := registry.RegisterOrganization(db.Organization{Name: "bla"})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 1, len(eventSystem.Events))
		assert.Equal(t, "RegisterOrganizationEvent", string(eventSystem.Events[0].Type()))
		assert.False(t, eventSystem.Events[0].IssuedAt().IsZero())
		event := events.RegisterOrganizationEvent{}
		if assert.NoError(t, eventSystem.Events[0].Unmarshal(&event)) {
			assert.Equal(t, "bla", event.Organization.Name)
		}
	})
	t.Run("Check event emitted: update organization", func(t *testing.T) {
		eventSystem := &MockEventSystem{Events: []events.Event{}}
		registry := Registry{
			Config: RegistryConfig{
			},
			EventSystem: eventSystem,
		}
		err := registry.RemoveOrganization("abc")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 1, len(eventSystem.Events))
		assert.Equal(t, "RemoveOrganizationEvent", string(eventSystem.Events[0].Type()))
		assert.False(t, eventSystem.Events[0].IssuedAt().IsZero())
		event := events.RemoveOrganizationEvent{}
		if assert.NoError(t, eventSystem.Events[0].Unmarshal(&event)) {
			assert.Equal(t, "abc", event.OrganizationID)
		}
	})
	t.Run("Check event emitted: update organization", func(t *testing.T) {
		eventSystem := &MockEventSystem{Events: []events.Event{}}
		registry := Registry{
			Config: RegistryConfig{
			},
			EventSystem: eventSystem,
		}
		err := registry.RemoveOrganization("abc")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 1, len(eventSystem.Events))
		assert.Equal(t, "RemoveOrganizationEvent", string(eventSystem.Events[0].Type()))
		assert.False(t, eventSystem.Events[0].IssuedAt().IsZero())
		event := events.RemoveOrganizationEvent{}
		if assert.NoError(t, eventSystem.Events[0].Unmarshal(&event)) {
			assert.Equal(t, "abc", event.OrganizationID)
		}
	})
}

func configureIdleTimeout() {
	ReloadRegistryIdleTimeout = 100 * time.Millisecond
}

func copyDir(src string, dst string) {
	for _, file := range findJSONFiles(src) {
		if strings.HasSuffix(file, ".json") {
			err := copyFile(fmt.Sprintf("%s/%s", src, file), fmt.Sprintf("%s/%s", dst, file))
			if err != nil {
				panic(err)
			}
		}
	}
}

func findJSONFiles(src string) []string {
	dir, err := ioutil.ReadDir(src)
	if err != nil {
		panic(err)
	}
	files := make([]string, 0)
	for _, entry := range dir {
		if strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}
	return files
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	return err
}

func cleanup() {
	err := os.RemoveAll("../tmp")
	if err != nil {
		logrus.Warnf("unable to clean tmp dir: %v", err)
	}
}

type MockEventSystem struct {
	Events []events.Event
}

func (m MockEventSystem) RegisterEventHandler(events.EventType, events.EventHandler) {
	// NOP
}

func (m MockEventSystem) ProcessEvent(events.Event) error {
	// NOP
	return nil
}

func (m *MockEventSystem) PublishEvent(event events.Event) error {
	m.Events = append(m.Events, event)
	return nil
}

func (m MockEventSystem) LoadAndApplyEvents(string) error {
	// NOP
	return nil
}
