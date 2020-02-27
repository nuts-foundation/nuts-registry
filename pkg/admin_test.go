package pkg

import (
	"errors"
	"fmt"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/storage"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryAdministration_RegisterEndpoint(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	registry := createRegistry(*repo)
	defer registry.Shutdown()

	var event = events.RegisterEndpointEvent{}
	registry.EventSystem.RegisterEventHandler(events.RegisterEndpoint, func(e events.Event) error {
		return e.Unmarshal(&event)
	})

	t.Run("ok", func(t *testing.T) {
		_, err := registry.RegisterEndpoint("orgId", "endpointId", "url", "type", "status", "version")
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

func TestRegistryAdministration_VendorClaim(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	registry := createRegistry(*repo)
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
		_, err := registry.VendorClaim("vendorId", "orgId", "orgName", keys)
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
	t.Run("org keys loaded from crypto", func(t *testing.T) {
		err := registry.crypto.GenerateKeyPairFor(types.LegalEntity{URI: "orgId"})
		if !assert.NoError(t, err) {
			return
		}
		_, err = registry.VendorClaim("vendorId", "orgId", "orgName", nil)
		if !assert.NoError(t, err) {
			return
		}
	})
	t.Run("org keys loaded from crypto - keys generated", func(t *testing.T) {
		// Assert no keys in crypto
		entity := types.LegalEntity{URI: "noKeysOrgId"}
		key, _ := registry.crypto.PublicKeyInJWK(entity)
		if !assert.Nil(t, key) {
			return
		}
		_, err := registry.VendorClaim("vendorId", entity.URI, "orgName", nil)
		if !assert.NoError(t, err) {
			return
		}
		// Assert key now exists in crypto
		key, _ = registry.crypto.PublicKeyInJWK(entity)
		assert.NotNil(t, key)
	})

	t.Run("error while generating key", func(t *testing.T) {
		// Assert no keys in crypto
		entity := types.LegalEntity{URI: "keyGenerationError"}
		c := registry.crypto.(*crypto.Crypto)
		var defaultKeySize = c.Config.Keysize
		c.Config.Keysize = -1
		defer func() {
			c.Config.Keysize = defaultKeySize
		}()
		_, err := registry.VendorClaim("vendorId", entity.URI, "orgName", nil)
		assert.Error(t, err)
	})

	t.Run("unable to load existing key", func(t *testing.T) {
		repo.Cleanup()
		os.MkdirAll(repo.Directory, os.ModePerm)
		entity := types.LegalEntity{URI: "org"}
		err := registry.crypto.GenerateKeyPairFor(entity)
		if !assert.NoError(t, err) {
			return
		}
		dirEntries, _ := ioutil.ReadDir(repo.Directory)
		ioutil.WriteFile(filepath.Join(repo.Directory, dirEntries[0].Name()), []byte("this is not a private key"), os.ModePerm)
		_, err = registry.VendorClaim("vendorID", entity.URI, "orgName", nil)
		assert.Error(t, err)
	})
}


func TestRegistryAdministration_RegisterVendor(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	registry := createRegistry(*repo)
	defer registry.Shutdown()

	t.Run("ok", func(t *testing.T) {
		var registerVendorEvent *events.RegisterVendorEvent
		registry.EventSystem.RegisterEventHandler(events.RegisterVendor, func(event events.Event) error {
			e := events.RegisterVendorEvent{}
			if err := event.Unmarshal(&e); err != nil {
				return err
			}
			registerVendorEvent = &e
			return nil
		})

		event, err := registry.RegisterVendor("foobar", "Foobar Software", "healthcare")
		if !assert.NoError(t, err) {
			return
		}
		// Verify RegisterVendor event emitted
		if !assert.NotNil(t, registerVendorEvent) {
			return
		}
		assert.NotNil(t, event)
		// Verify CA Certificate issued
		key, err := crypto.MapToJwk(registerVendorEvent.Keys[0].(map[string]interface{}))
		if err != nil {
			panic(err)
		}
		certType, _ := key.Get("ct")
		assert.Equal(t, string(cert.VendorCACertificate), certType)
		chain := key.X509CertChain()
		if !assert.NotNil(t, chain) {
			return
		}
		if !assert.Len(t, chain, 1) {
			return
		}
		assert.Equal(t, "Foobar Software CA", chain[0].Subject.CommonName)
	})
	t.Run("error: unable to publish event", func(t *testing.T) {
		registry.EventSystem.RegisterEventHandler(events.RegisterVendor, func(event events.Event) error {
			return errors.New("unit test error")
		})
		_, err := registry.RegisterVendor("foobar", "Foobar Software", "healthcare")
		assert.Contains(t, err.Error(), "unit test error")
	})
}


func createRegistry(repo test.TestRepo) Registry {
	registry := Registry{
		Config: RegistryConfig{
			Mode:     core.ServerEngineMode,
			Datadir:  repo.Directory,
			SyncMode: "fs",
		},
		EventSystem: events.NewEventSystem(),
	}
	err := registry.Configure()
	cryptoBackend, _ := storage.NewFileSystemBackend(repo.Directory)
	registry.crypto = &crypto.Crypto{
		Storage: cryptoBackend,
		Config: crypto.CryptoConfig{
			Keysize: 2048,
		},
	}
	if err != nil {
		panic(err)
	}
	return registry
}

func Test_marshalJwk(t *testing.T) {
	t.Run("invalid certificate", func(t *testing.T) {
		_, err := marshalJwk([]byte{1, 2, 3}, cert.VendorCACertificate)
		assert.Contains(t, err.Error(), "asn1: structure error")
	})
}