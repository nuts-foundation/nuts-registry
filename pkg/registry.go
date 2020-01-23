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
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
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

// ModuleName == Registry
const ModuleName = "Registry"

// RegistryClient is the interface to be implemented by any remote or local client
type RegistryClient interface {
	// EndpointsByOrganization returns all registered endpoints for an organization
	EndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]db.Endpoint, error)

	// SearchOrganizations searches the registry for any Organization mathing the given query
	SearchOrganizations(query string) ([]db.Organization, error)

	// OrganizationById returns an Organization given the Id or an error if it doesn't exist
	OrganizationById(id string) (*db.Organization, error)

	// RemoveOrganization removes the organization identified by id from the registry or returns an error if the organization does not exist
	RemoveOrganization(id string) error

	// RegisterOrganization adds the organization identified by id to the registry or returns an error if the organization already exists
	RegisterOrganization(db.Organization) error

	// ReverseLookup finds an exact match on name or returns an error if not found
	ReverseLookup(name string) (*db.Organization, error)
}

// RegistryConfig holds the config
type RegistryConfig struct {
	Mode         string
	SyncMode     string
	SyncAddress  string
	SyncInterval int
	Datadir      string
	Address      string
}

// Registry holds the config and Db reference
type Registry struct {
	Config      RegistryConfig
	Db          db.Db
	eventSystem events.EventSystem
	configOnce  sync.Once
	_logger     *logrus.Entry
	closers     []chan struct{}
	OnChange    func(registry *Registry)
}

var instance *Registry
var oneRegistry sync.Once

// RegistryInstance returns the singleton Registry
func RegistryInstance() *Registry {
	oneRegistry.Do(func() {
		instance = &Registry{
			eventSystem: events.NewEventSystem(),
			_logger:     logrus.StandardLogger().WithField("module", ModuleName),
		}
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (r *Registry) Configure() error {
	var err error

	r.configOnce.Do(func() {
		if r.Config.Mode == "server" {
			r.registerEventHandlers()
			// Apply stored events
			r.Db = db.New()
			if err := r.eventSystem.LoadAndApplyEvents(r.Config.Datadir); err != nil {
				r.logger().WithError(err).Warn("unable to load registry files")
			}
		}
	})
	return err
}

func (r *Registry) registerEventHandlers() {
	// TODO: We should receive a struct here, not a pointer to it
	r.eventSystem.RegisterEventHandler(reflect.TypeOf(&events.RegisterOrganizationEvent{}), func(e interface{}) error {
		event := e.(*events.RegisterOrganizationEvent)
		return r.Db.RegisterOrganization(event.Payload)
	})
	r.eventSystem.RegisterEventHandler(reflect.TypeOf(&events.RegisterEndpointEvent{}), func(e interface{}) error {
		event := e.(*events.RegisterEndpointEvent)
		r.Db.RegisterEndpoint(event.Payload)
		return nil
	})
	r.eventSystem.RegisterEventHandler(reflect.TypeOf(&events.RegisterEndpointOrganizationEvent{}), func(e interface{}) error {
		event := e.(*events.RegisterEndpointOrganizationEvent)
		return r.Db.RegisterEndpointOrganization(event.Payload)
	})
	r.eventSystem.RegisterEventHandler(reflect.TypeOf(&events.RemoveOrganizationEvent{}), func(e interface{}) error {
		event := e.(*events.RemoveOrganizationEvent)
		return r.Db.RemoveOrganization(event.Payload)
	})
}

// EndpointsByOrganization is a wrapper for sam func on DB
func (r *Registry) EndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]db.Endpoint, error) {
	return r.Db.FindEndpointsByOrganizationAndType(organizationIdentifier, endpointType)
}

// SearchOrganizations is a wrapper for sam func on DB
func (r *Registry) SearchOrganizations(query string) ([]db.Organization, error) {
	return r.Db.SearchOrganizations(query), nil
}

// OrganizationById is a wrapper for sam func on DB
func (r *Registry) OrganizationById(id string) (*db.Organization, error) {
	return r.Db.OrganizationById(id)
}

// RemoveOrganization is a wrapper for sam func on DB
func (r *Registry) RemoveOrganization(id string) error {
	return r.eventSystem.PublishEvent(&events.RemoveOrganizationEvent{Payload: id})
}

// RegisterOrganization is a wrapper for sam func on DB
func (r *Registry) RegisterOrganization(org db.Organization) error {
	return r.eventSystem.PublishEvent(&events.RegisterOrganizationEvent{Payload: org})
}

func (r *Registry) ReverseLookup(name string) (*db.Organization, error) {
	return r.Db.ReverseLookup(name)
}

// Start initiates the routines for auto-updating the data
func (r *Registry) Start() error {
	if r.Config.Mode == "server" {
		switch cm := r.Config.SyncMode; cm {
		case "fs":
			return r.startFileSystemWatcher()
		case "github":
			return r.startGithubSync()
		default:
			return errors.New(fmt.Sprintf("invalid syncMode: %s", cm))
		}
	}
	return nil
}

// Shutdown cleans up any leftover go routines
func (r *Registry) Shutdown() error {
	if r.Config.Mode == "server" {
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
	if err := r.eventSystem.LoadAndApplyEvents(r.Config.Datadir); err != nil {
		return err
	}

	if r.OnChange != nil {
		r.OnChange(r)
	}

	return nil
}

func (r *Registry) startFileSystemWatcher() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	closer := make(chan struct{})

	go func() {
		orgs := false
		ends := false
		eos := false

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}

				// we need to receive all 3 events before reloading the files
				r.logger().Debugf("Received file watcher event: %s", event.String())
				if strings.Contains(event.Name, "/organizations.json") && event.Op&fsnotify.Write == fsnotify.Write {
					orgs = true
				}
				if strings.Contains(event.Name, "/endpoints.json") && event.Op&fsnotify.Write == fsnotify.Write {
					ends = true
				}
				if strings.Contains(event.Name, "/endpoints_organizations.json") && event.Op&fsnotify.Write == fsnotify.Write {
					eos = true
				}

				if orgs && ends && eos {
					if r.Db != nil {
						if err := r.Load(); err != nil {
							r.logger().WithError(err).Error("error during reloading of registry files")
						}
					}
					orgs = false
					ends = false
					eos = false
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

	if err := w.Add(r.Config.Datadir); err != nil {
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

			r.logger().Debugf("Downloading registry files from %s to %s", r.Config.SyncAddress, r.Config.Datadir)
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

		name := header.Name
		nameParts := strings.Split(name, "/")
		name = nameParts[len(nameParts)-1]

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if strings.Index(name, ".json") > 0 {
				targetPath := fmt.Sprintf("%s/%s", r.Config.Datadir, name)

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

	// remove file
	return os.Remove(tarGzFile)
}

func (r *Registry) logger() *logrus.Entry {
	if r._logger == nil {
		r._logger = logrus.StandardLogger().WithField("module", ModuleName)
	}
	return r._logger
}
