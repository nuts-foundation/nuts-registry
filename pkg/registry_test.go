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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/gommon/random"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
				Mode:     core.ServerEngineMode,
				SyncMode: "unknown",
				Datadir:  ".",
			},
			Db: &db.MemoryDb{},
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
				Mode:     core.ServerEngineMode,
				SyncMode: "fs",
				Datadir:  ".",
			},
			Db: &db.MemoryDb{},
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
				Mode:     core.ServerEngineMode,
				SyncMode: "fs",
				Datadir:  ":",
			},
			Db: &db.MemoryDb{},
		}

		err := registry.Start()

		if err == nil {
			t.Error("Expected error, got nothing")
		}
	})

	t.Run("Shutdown stops the file watcher", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     core.ServerEngineMode,
				SyncMode: "fs",
				Datadir:  ".",
			},
			Db: &db.MemoryDb{},
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
				Mode:    core.ServerEngineMode,
				Datadir: "../test_data/valid_files",
			},
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
		}

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
	t.Run("error while configuring event system", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:    core.ServerEngineMode,
				Datadir: "///",
			},
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
		}
		err := registry.Configure()
		assert.Error(t, err)
	})

	t.Run("error while loading events", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		defer repo.Cleanup()
		registry := Registry{
			Config: RegistryConfig{
				Mode:    core.ServerEngineMode,
				Datadir: repo.Directory,
			},
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
		}
		os.MkdirAll(filepath.Join(repo.Directory, "events"), os.ModePerm)
		err = ioutil.WriteFile(filepath.Join(repo.Directory, "events/20200123091400001-RegisterOrganizationEvent.json"), []byte("this is a file"), os.ModePerm)
		if !assert.NoError(t, err) {
			return
		}
		err = registry.Configure()
		assert.Error(t, err)
	})
}

func TestRegistry_FileUpdate(t *testing.T) {
	configureIdleTimeout()

	t.Run("New files are loaded", func(t *testing.T) {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)

		wg := sync.WaitGroup{}
		wg.Add(1)

		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		defer repo.Cleanup()
		registry := Registry{
			Config: RegistryConfig{
				Mode:     core.ServerEngineMode,
				Datadir:  repo.Directory,
				SyncMode: "fs",
			},
			OnChange: func(registry *Registry) {
				wg.Done()
			},
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
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
		err = repo.ImportDir("../test_data/valid_files")
		if !assert.NoError(t, err) {
			return
		}

		wg.Wait()

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func TestRegistry_GithubUpdate(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	configureIdleTimeout()

	t.Run("New files are downloaded", func(t *testing.T) {
		handler := &ZipHandler{}
		server := httptest.NewServer(handler)
		defer server.Close()

		wg := sync.WaitGroup{}
		wg.Add(1)

		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		defer repo.Cleanup()

		registry := Registry{
			Config: RegistryConfig{
				Mode:         core.ServerEngineMode,
				Datadir:      repo.Directory,
				SyncMode:     "github",
				SyncAddress:  server.URL,
				SyncInterval: 60,
			},
			OnChange: func(registry *Registry) {
				println("EVENT")
				wg.Done()
			},
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
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

func configureIdleTimeout() {
	ReloadRegistryIdleTimeout = 100 * time.Millisecond
}

func TestRegistry_EndpointsByOrganizationAndType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		mockDb.EXPECT().FindEndpointsByOrganizationAndType("id", nil)
		(&Registry{Db: mockDb}).EndpointsByOrganizationAndType("id", nil)
	})
}

func TestRegistry_SearchOrganizations(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		mockDb.EXPECT().SearchOrganizations("query")
		(&Registry{Db: mockDb}).SearchOrganizations("query")
	})
}

func TestRegistry_OrganizationById(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		mockDb.EXPECT().OrganizationById("id")
		(&Registry{Db: mockDb}).OrganizationById("id")
	})
}

func TestRegistry_ReverseLookup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		mockDb.EXPECT().ReverseLookup("id")
		(&Registry{Db: mockDb}).ReverseLookup("id")
	})
}
