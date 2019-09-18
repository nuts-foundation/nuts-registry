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
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"regexp"
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

// ModuleName == Registry
const ModuleName = "Registry"

// RegistryClient is the interface to be implemented by any remote or local client
type RegistryClient interface {

	// EndpointsByOrganization returns all registered endpoints for an organization
	EndpointsByOrganization(organizationIdentifier string) ([]db.Endpoint, error)

	// SearchOrganizations searches the registry for any Organization mathing the given query
	SearchOrganizations(query string) ([]db.Organization, error)

	// OrganizationById returns an Organization given the Id or an error if it doesn't exist
	OrganizationById(id string) (*db.Organization, error)

	// RemoveOrganization removes the organization identified by id from the registry or returns an error if the organization does not exist
	RemoveOrganization(id string) error

	// RegisterOrganization adds the organization identified by id to the registry or returns an error if the organization already exists
	RegisterOrganization(db.Organization) error
}

// RegistryConfig holds the config
type RegistryConfig struct {
	Mode        string
	SyncMode    string
	SyncAddress string
	Datadir     string
	Address     string
}

// Registry holds the config and Db reference
type Registry struct {
	Config       RegistryConfig
	Db           db.Db
	configOnce   sync.Once
	watcher      *watcher.Watcher
	_logger       *logrus.Entry
}

var instance *Registry
var oneRegistry sync.Once

// RegistryInstance returns the singleton Registry
func RegistryInstance() *Registry {
	oneRegistry.Do(func() {
		instance = &Registry{
			_logger: logrus.StandardLogger().WithField("module", ModuleName),
		}
	})

	return instance
}

// Configure initializes the db, but only when in server mode
func (r *Registry) Configure() error {
	var err error

	r.configOnce.Do(func() {
		if r.Config.Mode == "server" {
			// load static Db
			r.Db = db.New()
			err = r.Db.Load(r.Config.Datadir)
		}
	})
	return err
}

// EndpointsByOrganization is a wrapper for sam func on DB
func (r *Registry) EndpointsByOrganization(organizationIdentifier string) ([]db.Endpoint, error) {
	return r.Db.FindEndpointsByOrganization(organizationIdentifier)
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
	return r.Db.RemoveOrganization(id)
}

// RegisterOrganization is a wrapper for sam func on DB
func (r *Registry) RegisterOrganization(org db.Organization) error {
	return r.Db.RegisterOrganization(org)
}

// Start initiates the routines for auto-updating the data
func (r *Registry) Start() error {
	if r.Config.Mode == "server" {
		switch cm := r.Config.SyncMode; cm {
		case "fs":
			return r.startFileSystemWatcher()
		case "github":
		default:
			return errors.New(fmt.Sprintf("invalid syncMode: %s", cm))
		}
	}
	return nil
}

// Shutdown cleans up any leftover go routines
func (r *Registry) Shutdown() error {
	if r.Config.Mode == "server" && r.watcher != nil {
		r.logger().Debug("Closing File watcher")
		r.watcher.Closed <- struct{}{}
		r.watcher.Close()
		r.logger().Info("File watcher closed")
	}
	return nil
}

func (r *Registry) startFileSystemWatcher() error {
	r.watcher = watcher.New()
	r.watcher.SetMaxEvents(1)
	regex := regexp.MustCompile("^.*\\.json")
	r.watcher.AddFilterHook(watcher.RegexFilterHook(regex, false))

	go func() {
		for {
			select {
			case event := <-r.watcher.Event:
				r.logger().Debugf("Received file watcher event: %s", event.String())
				if r.Db != nil {
					if err := r.Db.Load(r.Config.Datadir); err != nil {
						r.logger().Errorf("error during reloading of files: %v", err)
					}
				}
			case err := <-r.watcher.Error:
				r.logger().Errorf("Received file watcher error: %v", err)
			case <- r.watcher.Closed:
				return
			}
		}
	}()

	if err := r.watcher.Add(r.Config.Datadir); err != nil {
		return err
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	for path, _ := range r.watcher.WatchedFiles() {
		r.logger().Debugf("Watching %s for changes", path)
	}

	go func() {
		if err := r.watcher.Start(time.Millisecond * 100); err != nil {
			r.logger().Error(err)
		}
	}()

	return nil
}

func (r *Registry) logger() *logrus.Entry {
	if r._logger == nil {
		r._logger = logrus.StandardLogger().WithField("module", ModuleName)
	}
	return r._logger
}
