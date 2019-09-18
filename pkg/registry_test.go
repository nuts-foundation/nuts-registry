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
	"testing"
	"time"
)

func TestRegistry_Start(t *testing.T) {
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

		if registry.watcher == nil {
			t.Error("Expected watcher to be set")
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

		// stopping the file watcher removes all files from the watch list
		if len(registry.watcher.WatchedFiles()) > 0 {
			t.Error("Expected no watched files")
		}
	})
}

func TestRegistry_Configure(t *testing.T) {
	t.Run("Configure loads the BD", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				Datadir:  "../test_data/valid_files",
			},
		}

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}
