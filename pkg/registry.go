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
	"sync"
	"time"

	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-registry/logging"

	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/network"

	"github.com/nuts-foundation/nuts-crypto/client"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	core "github.com/nuts-foundation/nuts-go-core"
	networkClient "github.com/nuts-foundation/nuts-network/client"
	networkPkg "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
)

// ConfDataDir is the config name for specifiying the data location of the requiredFiles
const ConfDataDir = "datadir"

// ConfMode is the config name for the engine mode, server or client
const ConfMode = "mode"

// ConfAddress is the config name for the http server/client address
const ConfAddress = "address"

// ConfSyncMode is the config name for the used SyncMode
const ConfSyncMode = "syncMode"

// ConfSyncAddress is the config name for the remote address used to fetch updated registry files
const ConfSyncAddress = "syncAddress"

// ConfSyncInterval is the config name for the interval in minutes to look for new registry files online
const ConfSyncInterval = "syncInterval"

// ConfOrganisationCertificateValidity is the config name for the number of days organisation certificates are valid
const ConfOrganisationCertificateValidity = "organisationCertificateValidity"

// ConfVendorCACertificateValidity is the config name for the number of days vendor CA certificates are valid
const ConfVendorCACertificateValidity = "vendorCACertificateValidity"

// ConfClientTimeout is the time-out for the client in seconds (e.g. when using the CLI).
const ConfClientTimeout = "clientTimeout"

// ModuleName == Registry
const ModuleName = "Registry"

// ReloadRegistryIdleTimeout defines the cooling down period after receiving a file watcher notification, before
// the registry is reloaded (from disk).
var ReloadRegistryIdleTimeout time.Duration

// RegistryClient is an alias for the DIDStore so older code can still use it.
type RegistryClient interface {
	DIDStore
}

// RegistryConfig holds the config
type RegistryConfig struct {
	Mode          string
	Datadir       string
	Address       string
	ClientTimeout int
}

func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		Datadir:       "./data",
		Address:       "localhost:1323",
		ClientTimeout: 10,
	}
}

// Registry holds the config and Db reference
type Registry struct {
	Config            RegistryConfig
	network           networkPkg.NetworkClient
	crypto            crypto.Client
	OnChange          func(registry *Registry)
	networkAmbassador network.Ambassador
	configOnce        sync.Once
	DIDStore          DIDStore
	_logger           *logrus.Entry
	closers           []chan struct{}
}

func (r *Registry) Search(onlyOwn bool, tags []string) ([]did.Document, error) {
	logging.Log().Debugf("Search called (onlyOwn: %s, tags: %v)", onlyOwn, tags)
	return r.DIDStore.Search(onlyOwn, tags)
}

func (r *Registry) Create() (*did.Document, error) {
	return r.DIDStore.Create()
}

func (r *Registry) Get(DID did.DID) (*did.Document, *DIDDocumentMetadata, error) {
	return r.DIDStore.Get(DID)
}

func (r *Registry) GetByTag(tag string) (*did.Document, *DIDDocumentMetadata, error) {
	return r.DIDStore.GetByTag(tag)
}

func (r *Registry) Update(DID did.DID, hash model.Hash, nextVersion did.Document) (*did.Document, error) {
	return r.DIDStore.Update(DID, hash, nextVersion)

}

func (r *Registry) Tag(DID did.DID, tags []string) error {
	return r.DIDStore.Tag(DID, tags)
}

var instance *Registry
var oneRegistry sync.Once

func init() {
	ReloadRegistryIdleTimeout = 3 * time.Second
}

// RegistryInstance returns the singleton Registry
func RegistryInstance() *Registry {
	if instance != nil {
		return instance
	}
	oneRegistry.Do(func() {
		instance = NewRegistryInstance(DefaultRegistryConfig(), client.NewCryptoClient(), networkClient.NewNetworkClient())
	})

	return instance
}

func NewRegistryInstance(config RegistryConfig, cryptoClient crypto.Client, networkClient pkg.NetworkClient) *Registry {
	return &Registry{
		Config:  config,
		crypto:  cryptoClient,
		network: networkClient,
		_logger: logging.Log(),
	}
}

// Configure initializes the db, but only when in server mode
func (r *Registry) Configure() error {
	var err error

	r.configOnce.Do(func() {
		cfg := core.NutsConfig()
		r.Config.Mode = cfg.GetEngineMode(r.Config.Mode)
		if r.Config.Mode == core.ServerEngineMode {
			if r.networkAmbassador == nil {
				r.networkAmbassador = network.NewAmbassador(r.network, r.crypto)
			}
		}
	})
	return err
}

// Start initiates the routines for auto-updating the data
func (r *Registry) Start() error {
	if r.Config.Mode == core.ServerEngineMode {
		r.networkAmbassador.Start()
	}
	return nil
}

// Shutdown cleans up any leftover go routines
func (r *Registry) Shutdown() error {
	if r.Config.Mode == core.ServerEngineMode {
		logging.Log().Debug("Sending close signal to all routines")
		for _, ch := range r.closers {
			ch <- struct{}{}
		}
		logging.Log().Info("All routines closed")
	}
	return nil
}

func (r *Registry) Diagnostics() []core.DiagnosticResult {
	return []core.DiagnosticResult{}
}

func (r *Registry) getEventsDir() string {
	return r.Config.Datadir + "/events"
}
