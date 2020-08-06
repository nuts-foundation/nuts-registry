package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg"
	pkg2 "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
	"path"
)

func NewTestRegistryInstance(testDirectory string) *Registry {
	config := TestRegistryConfig(testDirectory)
	newInstance := NewRegistryInstance(config, pkg.NewTestCryptoInstance(testDirectory), pkg2.NewTestNetworkInstance(testDirectory))
	if err := newInstance.Configure(); err != nil {
		logrus.Fatal(err)
	}
	instance = newInstance
	return newInstance
}

func TestRegistryConfig(testDirectory string) RegistryConfig {
	config := DefaultRegistryConfig()
	config.Datadir = path.Join(testDirectory, "registry")
	return config
}