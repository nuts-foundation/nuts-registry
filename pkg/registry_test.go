// +build !race

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

package pkg

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-go-test/io"
	pkg2 "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/network"
	"github.com/nuts-foundation/nuts-registry/test"

	"github.com/spf13/cobra"

	"github.com/labstack/gommon/random"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/stretchr/testify/assert"
)

type ZipHandler struct {
}

var vendorId core.PartyID

const vendorName = "Test Vendor"

func init() {
	vendorId = test.VendorID("4")
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
	configureIdentity()
	NewTestRegistryInstance(io.TestDirectory(t))
	registry1 := RegistryInstance()
	registry2 := RegistryInstance()
	assert.Same(t, registry1, registry2)
}

func TestRegistry_Start(t *testing.T) {
	configureIdleTimeout()
	configureIdentity()
	t.Run("Starting sets the file watcher", func(t *testing.T) {
		registry := NewTestRegistryInstance(io.TestDirectory(t))
		err := registry.Start()
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, registry.closers, 1)
	})

	t.Run("Invalid datadir gives error on Start", func(t *testing.T) {
		registry := NewTestRegistryInstance(io.TestDirectory(t))
		registry.Config.Datadir = ":"
		err := registry.Start()
		assert.Error(t, err)
	})

	t.Run("Shutdown stops the file watcher", func(t *testing.T) {
		registry := NewTestRegistryInstance(io.TestDirectory(t))
		err := registry.Start()
		if !assert.NoError(t, err) {
			return
		}
		// watcher delay
		time.Sleep(time.Millisecond * 100)
		err = registry.Shutdown()
		if !assert.NoError(t, err) {
			return
		}
	})
}

func TestRegistry_Configure(t *testing.T) {
	configureIdleTimeout()
	configureIdentity()
	create := func(t *testing.T) *Registry {
		testDirectory := io.TestDirectory(t)
		return &Registry{
			Config:  TestRegistryConfig(testDirectory),
			crypto:  pkg.NewTestCryptoInstance(testDirectory),
			network: pkg2.NewTestNetworkInstance(testDirectory),
		}
	}
	t.Run("ok", func(t *testing.T) {
		registry := create(t)
		registry.Config.Datadir = "../test_data/valid_files"
		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}
		if len(registry.Db.SearchOrganizations("")) == 0 {
			t.Error("Expected loaded organizations, got 0")
		}
	})
	t.Run("ok - client mode", func(t *testing.T) {
		os.Setenv("NUTS_MODE", "cli")
		defer os.Unsetenv("NUTS_MODE")
		core.NutsConfig().Load(&cobra.Command{})
		registry := create(t)
		err := registry.Configure()
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("error - configuring event system", func(t *testing.T) {
		registry := create(t)
		registry.Config.Datadir = "///"
		err := registry.Configure()
		assert.Error(t, err)
	})

	t.Run("error - loading events", func(t *testing.T) {
		registry := create(t)
		os.MkdirAll(filepath.Join(registry.Config.Datadir, "events"), os.ModePerm)
		err := ioutil.WriteFile(filepath.Join(registry.Config.Datadir, "events/20200123091400001-RegisterOrganizationEvent.json"), []byte("this is a file"), os.ModePerm)
		if !assert.NoError(t, err) {
			return
		}
		err = registry.Configure()
		assert.Error(t, err)
	})
}

func TestRegistry_Diagnostics(t *testing.T) {
	registry := createTestContext(t).registry
	diagnostics := registry.Diagnostics()
	assert.NotEmpty(t, diagnostics)
}

func configureIdentity() {
	os.Setenv("NUTS_IDENTITY", vendorId.String())
	core.NutsConfig().Load(&cobra.Command{})
}

func configureIdleTimeout() {
	ReloadRegistryIdleTimeout = 100 * time.Millisecond
}

func createRegistry(repo *test.TestRepo) *Registry {
	core.NutsConfig().Load(&cobra.Command{})
	return NewTestRegistryInstance(repo.Directory)
}

type testContext struct {
	identity          core.PartyID
	vendorName        string
	registry          *Registry
	networkAmbassador *network.MockAmbassador
	mockCtrl          *gomock.Controller
	repo              *test.TestRepo
}

func (cxt *testContext) empty() {
	cxt.repo.Cleanup()
	os.MkdirAll(filepath.Join(cxt.repo.Directory, "crypto"), os.ModePerm)
	os.MkdirAll(filepath.Join(cxt.repo.Directory, "events"), os.ModePerm)
}

func (cxt *testContext) close() {
	defer cxt.mockCtrl.Finish()
	defer os.Unsetenv("NUTS_IDENTITY")
	defer cxt.registry.Shutdown()
	defer cxt.repo.Cleanup()
}

func createTestContext(t *testing.T) testContext {
	os.Setenv("NUTS_IDENTITY", vendorId.String())
	repo, err := test.NewTestRepo(t)
	if err != nil {
		panic(err)
	}
	mockCtrl := gomock.NewController(t)
	context := testContext{
		identity:          vendorId,
		vendorName:        vendorName,
		registry:          createRegistry(repo),
		mockCtrl:          mockCtrl,
		networkAmbassador: network.NewMockAmbassador(mockCtrl),
		repo:              repo,
	}
	return context
}
