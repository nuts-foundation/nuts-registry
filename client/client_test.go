package client

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitialize(t *testing.T) {
	t.Run("server mode", func(t *testing.T) {
		instance := pkg.RegistryInstance()
		instance.Config.Mode = core.ServerEngineMode
		assert.IsType(t, &pkg.Registry{}, initialize(instance))
	})
	t.Run("client mode", func(t *testing.T) {
		instance := pkg.RegistryInstance()
		instance.Config.Mode = core.ClientEngineMode
		assert.IsType(t, api.HttpClient{}, initialize(instance))
	})
}

func TestNewRegistryClient(t *testing.T) {
	client := NewRegistryClient()
	assert.NotNil(t, client)
}