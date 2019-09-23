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
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
	t.Run("Configure loads the BD", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:    "server",
				Datadir: "../test_data/valid_files",
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

func TestRegistry_FileUpdate(t *testing.T) {
	t.Run("New files are loaded", func(t *testing.T) {
		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				Datadir:  "../tmp",
				SyncMode: "fs",
			},
		}

		os.Mkdir("../tmp", os.ModePerm)
		copyDir("../test_data/all_empty_files", "../tmp")

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
		copyDir("../test_data/valid_files", "../tmp")

		time.Sleep(time.Millisecond * 500)

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func TestRegistry_GithubUpdate(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)

	t.Run("New files are downloaded", func(t *testing.T) {
		handler := &ZipHandler{}
		server := httptest.NewServer(handler)
		defer server.Close()

		registry := Registry{
			Config: RegistryConfig{
				Mode:     "server",
				Datadir:  "../tmp",
				SyncMode: "github",
				SyncAddress: server.URL,
				SyncInterval: 60,
			},
		}

		os.Mkdir("../tmp", os.ModePerm)
		copyDir("../test_data/all_empty_files", "../tmp")

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		// wait for download
		time.Sleep(time.Millisecond * 500)

		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
}

func copyDir(src string, dst string) {
	for _, f := range []string{"organizations.json", "endpoints.json", "endpoints_organizations.json"} {
		copyFile(fmt.Sprintf("%s/%s", src, f), fmt.Sprintf("%s/%s", dst, f))
	}
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
