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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-registry/internal/storage"
	"github.com/nuts-foundation/nuts-registry/logging"
	"sync"
	"time"

	"github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/network"

	"github.com/nuts-foundation/nuts-crypto/client"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	core "github.com/nuts-foundation/nuts-go-core"
	networkClient "github.com/nuts-foundation/nuts-network/client"
	networkPkg "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
)

type constError string

func (err constError) Error() string {
	return string(err)
}

const (
	// ErrIncorrectLastVersionHash indicates that the DID Document can't be updated because the supplied hash doesn't
	// match with the hash of the last version (of the DID Document).
	ErrIncorrectLastVersionHash = constError("supplied hash of last DID Document version is incorrect")
	ErrInvalidDIDDocument       = constError("invalid DID document")
)

// ConfDataDir is the config name for specifiying the data location of the requiredFiles
const ConfDataDir = "datadir"

// ConfMode is the config name for the engine mode, server or client
const ConfMode = "mode"

// ConfAddress is the config name for the http server/client address
const ConfAddress = "address"

// ConfClientTimeout is the time-out for the client in seconds (e.g. when using the CLI).
const ConfClientTimeout = "clientTimeout"

// ModuleName == Registry
const ModuleName = "Registry"

// RegistryClient is an alias for the DIDService so older code can still use it.
type RegistryClient interface {
	DIDService
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
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	generatedDID := ecdsaPublicKeyToNutsDID(privateKey.PublicKey)
	publicKeyJWK, err := jwk.New(privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	if err := jwk.AssignKeyID(publicKeyJWK); err != nil {
		return nil, err
	}
	verificationMethod, err := jwkToVerificationMethod(generatedDID, publicKeyJWK)
	if err != nil {
		return nil, err
	}
	document := did.Document{
		Context:            []did.URI{did.DIDContextV1URI()},
		ID:                 generatedDID,
		VerificationMethod: []did.VerificationMethod{*verificationMethod},
		Authentication:     []did.VerificationRelationship{{VerificationMethod: verificationMethod}},
	}
	var metadata DIDDocumentMetadata

	// Send DID Document to Nuts Network
	if didDocumentAsJSON, err := json.Marshal(document); err != nil {
		return nil, fmt.Errorf("unable to marshal DID Document: %w", err)
	} else if envelope, err := r.network.AddDocumentWithContents(time.Now(), "application/json+did-document", didDocumentAsJSON); err != nil {
		return nil, fmt.Errorf("unable to register DID Document on Nuts Network: %w", err)
	} else {
		// TODO: envelope is still the old Nuts Network Document type, this should be changed to Distributed Document
		// format as specified by RFC004 and used in RFC006.
		metadata.Hash = envelope.Hash
		metadata.Version = 0
		metadata.OriginJWSHash = envelope.Hash
	}

	if err = r.DIDStore.Add(document, DIDDocumentMetadata{
		Created:       time.Time{},
		Updated:       time.Time{},
		Version:       0,
		OriginJWSHash: model.Hash{},
		Hash:          model.Hash{},
	}); err != nil {
		return nil, fmt.Errorf("unable to store created DID: %w", err)
	}

	return &document, nil
}

func (r *Registry) Get(DID did.DID) (*did.Document, *DIDDocumentMetadata, error) {
	return r.DIDStore.Get(DID)
}

func (r *Registry) GetByTag(tag string) (*did.Document, *DIDDocumentMetadata, error) {
	return r.DIDStore.GetByTag(tag)
}

func (r *Registry) Update(DID did.DID, hash model.Hash, nextVersion did.Document) (*did.Document, error) {
	current, metadata, err := r.DIDStore.Get(DID)
	if err != nil {
		return nil, err
	} else if !metadata.Hash.Equals(hash) {
		return nil, ErrIncorrectLastVersionHash
	}

	// TODO: More validation?
	if !nextVersion.ID.Equals(DID) {
		return nil, ErrInvalidDIDDocument
	}
	return r.DIDStore.Update(DID, hash, nextVersion)

}

func (r *Registry) Tag(DID did.DID, tags []string) error {
	return r.DIDStore.Tag(DID, tags)
}

var instance *Registry
var oneRegistry sync.Once

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
		Config:   config,
		crypto:   cryptoClient,
		network:  networkClient,
		DIDStore: storage.NewMemoryDIDStore(),
		_logger:  logging.Log(),
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
