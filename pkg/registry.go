package pkg

import (
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"sync"
)

// ConfDataDir is the config name for specifiying the data location of the requiredFiles
const ConfDataDir = "datadir"
// ConfMode is the config name for the engine mode
const ConfMode = "mode"
// ConfAddress is the config name for the http server/client address
const ConfAddress = "address"

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

// Registry holds the config and Db reference
type Registry struct {
	Config struct {
		Mode string
		Datadir string
		Address string
	}
	Db         db.Db
	configOnce sync.Once
}

var instance *Registry
var oneRegistry sync.Once

// RegistryInstance returns the singleton Registry
func RegistryInstance() *Registry {
	oneRegistry.Do(func() {
		instance = &Registry{}
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

