package client

import (
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/sirupsen/logrus"
	"time"
)

// NewRegistryClient creates a new Local- or RemoteClient for the nuts registry
func NewRegistryClient() pkg.RegistryClient {
	registry := pkg.RegistryInstance()

	if registry.Config.Mode == "server" {
		if err := registry.Configure(); err != nil {
			logrus.Panic(err)
		}

		return registry
	} else {
		return api.HttpClient{
			ServerAddress: registry.Config.Address,
			Timeout: time.Second,
		}
	}
}
