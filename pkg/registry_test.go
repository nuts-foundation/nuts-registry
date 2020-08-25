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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-go-test/io"
	pkg2 "github.com/nuts-foundation/nuts-network/pkg"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/spf13/cobra"

	"github.com/golang/mock/gomock"
	"github.com/labstack/gommon/random"
	cryptoMock "github.com/nuts-foundation/nuts-crypto/test/mock"
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
	configureIdentity()
	NewTestRegistryInstance(io.TestDirectory(t))
	registry1 := RegistryInstance()
	registry2 := RegistryInstance()
	assert.Same(t, registry1, registry2)
}

func TestRegistry_Start(t *testing.T) {
	configureIdleTimeout()
	configureIdentity()
	t.Run("Start with an incorrect configuration returns error", func(t *testing.T) {
		registry := NewTestRegistryInstance(io.TestDirectory(t))
		registry.Config.SyncMode = "unknown"
		err := registry.Start()

		assert.EqualError(t, err, "invalid syncMode: unknown")
	})

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
		// Make sure engine services aren't initialized when running in client mode
		assert.Nil(t, registry.EventSystem)
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
	t.Run("error - vendor CA certificate validity invalid", func(t *testing.T) {
		registry := create(t)
		registry.Config.VendorCACertificateValidity = 0
		err := registry.Configure()
		assert.EqualError(t, err, "vendor CA certificate validity must be at least 1 day")
	})
	t.Run("error - organisation certificate validity invalid", func(t *testing.T) {
		registry := create(t)
		registry.Config.OrganisationCertificateValidity = 0
		err := registry.Configure()
		assert.EqualError(t, err, "organisation certificate validity must be at least 1 day")
	})
}

func TestRegistry_FileUpdate(t *testing.T) {
	configureIdleTimeout()
	configureIdentity()

	t.Run("New files are loaded", func(t *testing.T) {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)

		eventHandled := false
		cxt := createTestContext(t)
		cxt.registry.EventSystem.RegisterEventHandler(domain.VendorClaim, func(event events.Event, lookup events.EventLookup) error {
			eventHandled = true
			return nil
		})
		defer cxt.close()
		if err := cxt.registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if err := cxt.registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if len(cxt.registry.Db.SearchOrganizations("")) != 0 {
			t.Error("Expected empty db")
		}

		// copy valid files
		err := test.TestRepo{Directory: cxt.registry.Config.Datadir}.ImportDir("../test_data/valid_files")
		if !assert.NoError(t, err) {
			return
		}

		for i := 0; i < 5; i++ {
			if eventHandled {
				// Events were loaded
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatal("No events were loaded")
	})
}

func TestRegistry_GithubUpdate(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	configureIdleTimeout()
	configureIdentity()

	t.Run("New files are downloaded", func(t *testing.T) {
		handler := &ZipHandler{}
		server := httptest.NewServer(handler)
		defer server.Close()

		testDirectory := io.TestDirectory(t)
		registry := Registry{
			Config:      TestRegistryConfig(testDirectory),
			crypto:      pkg.NewTestCryptoInstance(testDirectory),
			network:     pkg2.NewTestNetworkInstance(testDirectory),
			EventSystem: events.NewEventSystem(domain.GetEventTypes()...),
		}
		registry.Config.SyncMode = "github"
		registry.Config.SyncAddress = server.URL
		defer registry.Shutdown()

		if err := registry.Configure(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		eventHandled := false
		registry.EventSystem.RegisterEventHandler(domain.VendorClaim, func(event events.Event, lookup events.EventLookup) error {
			eventHandled = true
			return nil
		})

		if err := registry.Start(); err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		for i := 0; i < 5; i++ {
			if eventHandled {
				// Events were loaded
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatal("No events were loaded")
	})
}

func TestRegistry_EndpointsByOrganizationAndType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		orgID := test.OrganizationID("id")
		mockDb.EXPECT().FindEndpointsByOrganizationAndType(orgID, nil)
		(&Registry{Db: mockDb}).EndpointsByOrganizationAndType(orgID, nil)
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
		orgID := test.OrganizationID("id")
		mockDb.EXPECT().OrganizationById(orgID)
		(&Registry{Db: mockDb}).OrganizationById(orgID)
	})
}

func TestRegistry_ReverseLookup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		orgID := test.OrganizationID("id")
		mockDb.EXPECT().ReverseLookup(orgID.String())
		(&Registry{Db: mockDb}).ReverseLookup(orgID.String())
	})
}

func TestRegistry_Verify(t *testing.T) {
	configureIdentity()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	t.Run("ok", func(t *testing.T) {
		mockDb := mock.NewMockDb(mockCtrl)
		mockDb.EXPECT().VendorByID(vendorId).Return(&db.Vendor{Identifier: vendorId})
		mockDb.EXPECT().OrganizationsByVendorID(vendorId).Return(nil)
		defer os.Unsetenv("NUTS_IDENTITY")
		evts, fix, err := (&Registry{Db: mockDb}).Verify(false)
		assert.NoError(t, err)
		assert.False(t, fix)
		assert.Empty(t, evts)
	})
}

func TestRegistry_VendorCAs(t *testing.T) {
	configureIdentity()
	pk1, _ := rsa.GenerateKey(rand.Reader, 1024)
	pk2, _ := rsa.GenerateKey(rand.Reader, 1024)
	pk3, _ := rsa.GenerateKey(rand.Reader, 1024)
	certBytes := test.GenerateCertificateEx(time.Now().AddDate(0, 0, -1), 2, pk1)
	root, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Intermediate CA", root, pk2, pk1)
	ca, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Vendor CA 1", ca, pk3, pk2)
	vca1, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Vendor CA 2", ca, pk3, pk2)
	vca2, _ := x509.ParseCertificate(certBytes)

	t.Run("returns empty slice when truststore is empty", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		repo, _ := test.NewTestRepo(t)
		trustStore, err := cert.NewTrustStore(fmt.Sprintf("%s/truststore.pem", repo.Directory))
		assert.NoError(t, err)

		cMock := cryptoMock.NewMockClient(mockCtrl)
		cMock.EXPECT().TrustStore().AnyTimes().Return(trustStore)
		cas := (&Registry{crypto: cMock}).VendorCAs()
		assert.Len(t, cas, 0)
	})

	t.Run("returns empty slice when truststore only contains a root", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		repo, _ := test.NewTestRepo(t)
		trustStore, err := cert.NewTrustStore(fmt.Sprintf("%s/truststore.pem", repo.Directory))
		assert.NoError(t, err)
		trustStore.AddCertificate(root)

		cMock := cryptoMock.NewMockClient(mockCtrl)
		cMock.EXPECT().TrustStore().AnyTimes().Return(trustStore)
		cas := (&Registry{crypto: cMock}).VendorCAs()
		assert.Len(t, cas, 0)
	})

	t.Run("returns empty slice when truststore only contains a root and 1 intermediate", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		repo, _ := test.NewTestRepo(t)
		trustStore, err := cert.NewTrustStore(fmt.Sprintf("%s/truststore.pem", repo.Directory))
		assert.NoError(t, err)
		trustStore.AddCertificate(root)
		trustStore.AddCertificate(ca)

		cMock := cryptoMock.NewMockClient(mockCtrl)
		cMock.EXPECT().TrustStore().AnyTimes().Return(trustStore)
		cas := (&Registry{crypto: cMock}).VendorCAs()
		assert.Len(t, cas, 0)
	})

	t.Run("returns 2 chains when truststore contains a root, an intermediate and 2 vendor CAs", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		repo, _ := test.NewTestRepo(t)
		trustStore, err := cert.NewTrustStore(fmt.Sprintf("%s/truststore.pem", repo.Directory))
		assert.NoError(t, err)
		trustStore.AddCertificate(root)
		trustStore.AddCertificate(ca)
		trustStore.AddCertificate(vca1)
		trustStore.AddCertificate(vca2)

		cMock := cryptoMock.NewMockClient(mockCtrl)
		cMock.EXPECT().TrustStore().AnyTimes().Return(trustStore)
		cas := (&Registry{crypto: cMock}).VendorCAs()
		assert.Len(t, cas, 2)
		assert.Len(t, cas[0], 3)
		assert.Len(t, cas[1], 3)
		assert.NotEqual(t, cas[0][0], cas[1][0])
		assert.Equal(t, cas[0][2], cas[1][2])
	})
}

func configureIdentity() {
	os.Setenv("NUTS_IDENTITY", vendorId.String())
	core.NutsConfig().Load(&cobra.Command{})
}

func configureIdleTimeout() {
	ReloadRegistryIdleTimeout = 100 * time.Millisecond
}