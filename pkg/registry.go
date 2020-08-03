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
	"archive/tar"
	"compress/gzip"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nuts-foundation/nuts-crypto/client"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	core "github.com/nuts-foundation/nuts-go-core"
	networkClient "github.com/nuts-foundation/nuts-network/client"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/pkg/network"
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

// RegistryClient is the interface to be implemented by any remote or local client
type RegistryClient interface {
	// EndpointsByOrganization returns all registered endpoints for an organization
	EndpointsByOrganizationAndType(organizationIdentifier core.PartyID, endpointType *string) ([]db.Endpoint, error)

	// SearchOrganizations searches the registry for any Organization matching the given query
	SearchOrganizations(query string) ([]db.Organization, error)

	// OrganizationById returns an Organization given the Id or an error if it doesn't exist
	OrganizationById(id core.PartyID) (*db.Organization, error)

	// ReverseLookup finds an exact match on name or returns an error if not found
	ReverseLookup(name string) (*db.Organization, error)

	// RegisterEndpoint registers an endpoint for an organization
	RegisterEndpoint(organizationID core.PartyID, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error)

	// VendorClaim registers an organization under a vendor. orgKeys are the organization's keys in JWK format
	VendorClaim(orgID core.PartyID, orgName string, orgKeys []interface{}) (events.Event, error)

	// RegisterVendor registers a vendor with the given id, name for the specified domain. If the vendor with this ID
	// already exists, it functions as an update.
	RegisterVendor(certificate *x509.Certificate) (events.Event, error)

	// RefreshOrganizationCertificate issues a new certificate for the organization. The organization must be registered under the current vendor.
	// If successful it returns the resulting event.
	RefreshOrganizationCertificate(organizationID core.PartyID) (events.Event, error)

	// Verify verifies the data in the registry owned by this node.
	// If fix=true, data will be fixed/upgraded when necessary (e.g. issue certificates). Events resulting from fixing the data are returned.
	// If the returned bool=true there's data to be fixed and Verify should be run with fix=true.
	Verify(fix bool) ([]events.Event, bool, error)

	// VendorCAs returns all registered vendors as list of chains, PEM encoded. The first entry in a chain will be the leaf and the last one the root.
	VendorCAs() [][]*x509.Certificate
}

// RegistryConfig holds the config
type RegistryConfig struct {
	Mode                            string
	SyncMode                        string
	SyncAddress                     string
	SyncInterval                    int
	Datadir                         string
	Address                         string
	VendorCACertificateValidity     int
	OrganisationCertificateValidity int
	ClientTimeout                   int
}

func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		SyncMode:                        "fs",
		SyncAddress:                     "https://codeload.github.com/nuts-foundation/nuts-registry-development/tar.gz/master",
		SyncInterval:                    30,
		Datadir:                         "./data",
		Address:                         "localhost:1323",
		VendorCACertificateValidity:     1095,
		OrganisationCertificateValidity: 365,
		ClientTimeout:                   10,
	}
}

// Registry holds the config and Db reference
type Registry struct {
	Config            RegistryConfig
	Db                db.Db
	EventSystem       events.EventSystem
	OnChange          func(registry *Registry)
	crypto            crypto.Client
	networkAmbassador network.Ambassador
	configOnce        sync.Once
	_logger           *logrus.Entry
	closers           []chan struct{}
}

var instance *Registry
var oneRegistry sync.Once

func init() {
	ReloadRegistryIdleTimeout = 3 * time.Second
}

// RegistryInstance returns the singleton Registry
func RegistryInstance() *Registry {
	oneRegistry.Do(func() {
		instance = &Registry{
			Config:  DefaultRegistryConfig(),
			_logger: logrus.StandardLogger().WithField("module", ModuleName),
		}
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (r *Registry) Configure() error {
	var err error

	r.configOnce.Do(func() {
		cfg := core.NutsConfig()
		r.Config.Mode = cfg.GetEngineMode(r.Config.Mode)
		if r.Config.Mode == core.ServerEngineMode {
			r.EventSystem = events.NewEventSystem(domain.GetEventTypes()...)
			if r.crypto == nil {
				r.crypto = client.NewCryptoClient()
			}
			if r.Config.VendorCACertificateValidity < 1 {
				err = errors.New("vendor CA certificate validity must be at least 1 day")
				return
			}
			if r.Config.OrganisationCertificateValidity < 1 {
				err = errors.New("organisation certificate validity must be at least 1 day")
				return
			}
			// Order of event processors:
			// -  TrustStore; must be first since certificates might be self-signed, and thus be added to the truststore
			//    before signature validation takes place.
			// -  Signature validator
			// -  Database, (in memory) queryable view of the registry
			// -  Network Ambassador, when all other processors succeeded the event is probably valid and can be broadcast.
			domain.NewCertificateEventHandler(r.crypto.TrustStore()).RegisterEventHandlers(r.EventSystem.RegisterEventHandler)
			signatureValidator := events.NewSignatureValidator(r.crypto.VerifyJWS, r.crypto.TrustStore())
			signatureValidator.RegisterEventHandlers(r.EventSystem.RegisterEventHandler, domain.GetEventTypes())
			r.Db = db.New()
			r.Db.RegisterEventHandlers(r.EventSystem.RegisterEventHandler)
			if r.networkAmbassador == nil {
				r.networkAmbassador = network.NewAmbassador(networkClient.NewNetworkClient(), r.crypto)
			}
			r.networkAmbassador.RegisterEventHandlers(r.EventSystem.RegisterEventHandler, domain.GetEventTypes())
			if err = r.EventSystem.Configure(r.getEventsDir()); err != nil {
				r.logger().WithError(err).Warn("Unable to configure event system")
				return
			}
			// Apply stored events
			if err = r.EventSystem.LoadAndApplyEvents(); err != nil {
				r.logger().WithError(err).Warn("Unable to load registry files")
			}
		}
	})
	return err
}

func (r *Registry) Verify(fix bool) ([]events.Event, bool, error) {
	return r.verify(core.NutsConfig(), fix)
}

// EndpointsByOrganization is a wrapper for sam func on DB
func (r *Registry) EndpointsByOrganizationAndType(organizationIdentifier core.PartyID, endpointType *string) ([]db.Endpoint, error) {
	return r.Db.FindEndpointsByOrganizationAndType(organizationIdentifier, endpointType)
}

// SearchOrganizations is a wrapper for sam func on DB
func (r *Registry) SearchOrganizations(query string) ([]db.Organization, error) {
	return r.Db.SearchOrganizations(query), nil
}

// OrganizationById is a wrapper for sam func on DB
func (r *Registry) OrganizationById(id core.PartyID) (*db.Organization, error) {
	return r.Db.OrganizationById(id)
}

func (r *Registry) ReverseLookup(name string) (*db.Organization, error) {
	return r.Db.ReverseLookup(name)
}

func (r *Registry) VendorCAs() [][]*x509.Certificate {
	now := time.Now()

	roots := r.crypto.TrustStore().GetRoots(now)
	var rootChains [][]*x509.Certificate

	for _, r := range roots {
		rootChains = append(rootChains, []*x509.Certificate{r})
	}

	intermediates := r.crypto.TrustStore().GetCertificates(rootChains, now, true)
	return r.crypto.TrustStore().GetCertificates(intermediates, now, true)
}

// Start initiates the routines for auto-updating the data
func (r *Registry) Start() error {
	if r.Config.Mode == core.ServerEngineMode {
		_, _, err := r.verify(*core.NutsConfig(), false)
		if err != nil {
			logrus.Error("Error occurred during registry data verification: ", err)
		}
		switch cm := r.Config.SyncMode; cm {
		case "fs":
			return r.startFileSystemWatcher()
		case "github":
			return r.startGithubSync()
		default:
			return fmt.Errorf("invalid syncMode: %s", cm)
		}
	}
	return nil
}

// Shutdown cleans up any leftover go routines
func (r *Registry) Shutdown() error {
	if r.Config.Mode == core.ServerEngineMode {
		r.logger().Debug("Sending close signal to all routines")
		for _, ch := range r.closers {
			ch <- struct{}{}
		}
		r.logger().Info("All routines closed")
	}
	return nil
}

// Load signals the Db to (re)load sources. On success the OnChange func is called
func (r *Registry) Load() error {
	if err := r.EventSystem.LoadAndApplyEvents(); err != nil {
		return err
	}

	if r.OnChange != nil {
		r.OnChange(r)
	}

	return nil
}

func (r *Registry) getEventsDir() string {
	return r.Config.Datadir + "/events"
}

func (r *Registry) startFileSystemWatcher() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	closer := make(chan struct{})

	go func() {
		// Timer needs to be initialized, otherwise Go panics. Reloading after an hour or so isn't a problem.
		var reloadRegistryTimer = time.NewTimer(time.Hour)
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}

				r.logger().Debugf("Received file watcher event: %s", event.String())
				if strings.HasSuffix(event.Name, ".json") &&
					(event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
					// When copying or extracting files, we have no guarantees that the filewatcher notifies us about
					// the files in natural order (of filenames), which we use for our event ordering. To circumvent
					// conflicts, we schedule an 'idle time-out' before actually reloading the registry as to wait for
					// new notifications. If a file is added while waiting, the reload timer is rescheduled.
					reloadRegistryTimer.Stop()
					reloadRegistryTimer = time.NewTimer(ReloadRegistryIdleTimeout)
				}
			case <-reloadRegistryTimer.C:
				if r.Db != nil {
					if err := r.Load(); err != nil {
						r.logger().WithError(err).Error("error during reloading of registry files")
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				r.logger().WithError(err).Error("Received file watcher error")
			case <-closer:
				r.logger().Debug("Stopping file watcher")
				return
			}
		}
	}()

	if err := w.Add(r.getEventsDir()); err != nil {
		return err
	}

	// register close channel
	r.closers = append(r.closers, closer)

	return nil
}

func (r *Registry) startGithubSync() error {
	if err := r.startFileSystemWatcher(); err != nil {
		r.logger().WithError(err).Error("Github sync not started due to file watcher problem")
		return err
	}

	close := make(chan struct{})
	go func(r *Registry, ch chan struct{}) {
		var eTag string

		for {
			var err error

			r.logger().Debugf("Downloading registry files from %s to %s", r.Config.SyncAddress, r.getEventsDir())
			if eTag, err = r.downloadAndUnzip(eTag); err != nil {
				r.logger().WithError(err).Error("Error downloading registry files")
			}

			select {
			case <-ch:
				r.logger().Debug("Stopping github download")
				return
			case <-time.After(time.Duration(int64(r.Config.SyncInterval) * time.Minute.Nanoseconds())):

			}
		}
	}(r, close)

	// register close channel
	r.closers = append(r.closers, close)

	r.logger().Info("Github sync started")

	return nil
}

func (r *Registry) downloadAndUnzip(eTag string) (string, error) {
	newTag, err := r.download(eTag)

	if err != nil {
		return eTag, err
	}

	if newTag == eTag {
		r.logger().Debug("Latest version on github is the same as local, skipping")
		return eTag, nil
	}

	return newTag, r.unzip()
}

func (r *Registry) download(eTag string) (string, error) {
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Get the data
	resp, err := httpClient.Get(r.Config.SyncAddress)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	newTag := resp.Header.Get("ETag")
	if eTag == newTag {
		return eTag, nil
	}

	tmpDir := fmt.Sprintf("%s/%s", r.Config.Datadir, "tmp")
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		// create and continue
		os.Mkdir(tmpDir, os.ModePerm)
	}

	tmpPath := fmt.Sprintf("%s/%s/%s", r.Config.Datadir, "tmp", "registry.tar.gz")

	// Create the file
	out, err := os.Create(tmpPath)
	if err != nil {
		return eTag, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return eTag, err
	}

	return newTag, nil
}

// unzip also strips the directory
func (r *Registry) unzip() error {
	tarGzFile := fmt.Sprintf("%s/%s/%s", r.Config.Datadir, "tmp", "registry.tar.gz")

	f, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// Top-level directory in tarball is skipped
		parts := strings.Split(filepath.Clean(header.Name), string(os.PathSeparator))
		if len(parts) == 1 {
			// Skip top-level entries
			continue
		}

		if header.Typeflag == tar.TypeReg {
			// Only extract JSON files
			if path.Ext(header.Name) == ".json" {
				targetPath := path.Join(r.Config.Datadir, filepath.Join(parts[1:]...))
				dst, err := os.Create(targetPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(dst, tarReader); err != nil {
					return err
				}
			}
		}
	}
	return os.Remove(tarGzFile)
}

func (r *Registry) logger() *logrus.Entry {
	if r._logger == nil {
		r._logger = logrus.StandardLogger().WithField("module", ModuleName)
	}
	return r._logger
}
