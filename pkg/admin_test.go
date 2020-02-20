package pkg

import (
	"fmt"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRegistry_RegisterEndpoint(t *testing.T) {
	cleanup()
	defer cleanup()

	registry := createRegistry()
	defer registry.Shutdown()

	var event = events.RegisterEndpointEvent{}
	registry.EventSystem.RegisterEventHandler(events.RegisterEndpoint, func(e events.Event) error {
		return e.Unmarshal(&event)
	})

	t.Run("ok", func(t *testing.T) {
		err := registry.RegisterEndpoint("orgId", "endpointId", "url", "type", "status", "version")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "orgId", string(event.Organization))
		assert.Equal(t, "endpointId", string(event.Identifier))
		assert.Equal(t, "url", event.URL)
		assert.Equal(t, "type", event.EndpointType)
		assert.Equal(t, "version", event.Version)
		assert.Equal(t, "status", event.Status)
	})
}

func TestRegistry_RegisterVendor(t *testing.T) {
	cleanup()
	defer cleanup()

	registry := createRegistry()
	defer registry.Shutdown()

	var event = events.RegisterVendorEvent{}
	registry.EventSystem.RegisterEventHandler(events.RegisterVendor, func(e events.Event) error {
		return e.Unmarshal(&event)
	})

	t.Run("ok", func(t *testing.T) {
		err := registry.RegisterVendor("id", "name")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "id", string(event.Identifier))
		assert.Equal(t, "name", event.Name)
	})
}

func TestRegistry_VendorClaim(t *testing.T) {
	cleanup()
	defer cleanup()

	registry := createRegistry()
	defer registry.Shutdown()

	var event = events.VendorClaimEvent{}
	registry.EventSystem.RegisterEventHandler(events.VendorClaim, func(e events.Event) error {
		return e.Unmarshal(&event)
	})

	t.Run("ok", func(t *testing.T) {
		var keys = []interface{}{
			map[string]interface{}{
				"e": 1234,
			},
		}
		err := registry.VendorClaim("vendorId", "orgId", "orgName", keys)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "vendorId", string(event.VendorIdentifier))
		assert.Equal(t, "orgId", string(event.OrgIdentifier))
		assert.Equal(t, "orgName", event.OrgName)
		assert.Equal(t, fmt.Sprintf("%v", keys), fmt.Sprintf("%v", event.OrgKeys))
		assert.False(t, event.Start.IsZero())
		assert.Nil(t, event.End)
	})
}

func createRegistry() Registry {
	registry := Registry{
		Config: RegistryConfig{
			Mode:     core.ServerEngineMode,
			Datadir:  "../tmp",
			SyncMode: "fs",
		},
		EventSystem: events.NewEventSystem(),
	}
	return registry
}

func cleanup() {
	err := os.RemoveAll("../tmp")
	if err != nil {
		logrus.Warnf("unable to clean tmp dir: %v", err)
	}
}