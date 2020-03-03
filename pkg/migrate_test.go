package pkg

import (
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

type mockConfig struct {
	identity string
}

func (m mockConfig) ServerAddress() string {
	panic("implement me")
}

func (m mockConfig) InStrictMode() bool {
	panic("implement me")
}

func (m mockConfig) Mode() string {
	panic("implement me")
}

func (m mockConfig) Identity() string {
	return m.identity
}

func (m mockConfig) GetEngineMode(engineMode string) string {
	panic("implement me")
}

func TestRegistry_verifyAndMigrateRegistry(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	registry := createRegistry(*repo)
	defer registry.Shutdown()

	t.Run("vendor not registered", func(t *testing.T) {
		registry.verifyAndMigrateRegistry(mockConfig{"abc"})
	})
	t.Run("vendor has no certificates", func(t *testing.T) {
		err := registry.EventSystem.PublishEvent(events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{Name: "Some Vendor", Identifier: "noCerts"}))
		if !assert.NoError(t, err) {
			return
		}
		registry.verifyAndMigrateRegistry(mockConfig{"noCerts"})
	})
	t.Run("vendor has certificates but no key material", func(t *testing.T) {
		_, err := registry.RegisterVendor("certsWithoutKeyMaterial", "Vendor", events.HealthcareDomain)
		if !assert.NoError(t, err) {
			return
		}
		// Empty key material directory
		eventsDirectory := filepath.Join(repo.Directory, "keys")
		if !assert.NoError(t, os.RemoveAll(eventsDirectory)) {
			return
		}
		if !assert.NoError(t, os.MkdirAll(eventsDirectory, os.ModePerm)) {
			return
		}
		registry.verifyAndMigrateRegistry(mockConfig{"certsWithoutKeyMaterial"})
	})
}
