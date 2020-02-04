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

package db

import (
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"strings"
)

type MemoryDb struct {
	vendors map[string]*vendor
}

type vendor struct {
	events.RegisterVendorEvent
	orgs map[string]*org
}

type org struct {
	events.VendorClaimEvent
	endpoints map[string]*endpoint
}

type endpoint struct {
	events.RegisterEndpointEvent
}

func (o org) toDb() Organization {
	return Organization{
		Identifier: toDbIdentifier(o.OrgIdentifier),
		Name:       o.OrgName,
		Keys:       o.OrgKeys,
		Endpoints:  o.toDbEndpoints(),
	}
}

// copyKeys is needed since the jwkSet.extractMap consumes the contents
func copyKeys(src []interface{}) []interface{} {
	var keys []interface{}
	for _, k := range src {
		nk := map[string]interface{}{}
		m := k.(map[string]interface{})
		for k, v := range m {
			nk[k] = v
		}
		keys = append(keys, nk)
	}
	return keys
}

// KeysAsSet transforms the raw map in Keys to a jwk.Set. If no keys are present, it'll return an empty set
func keysAsSet(keys []interface{}) (jwk.Set, error) {
	var set jwk.Set
	if len(keys) == 0 {
		return set, nil
	}

	m := make(map[string]interface{})

	m["keys"] = copyKeys(keys)
	err := set.ExtractMap(m)
	return set, err
}

func (e endpoint) toDb() Endpoint {
	return Endpoint{
		URL:          e.URL,
		EndpointType: e.EndpointType,
		Identifier:   toDbIdentifier(e.Identifier),
		Status:       e.Status,
		Version:      e.Version,
	}
}

func (o org) toDbEndpoints() []Endpoint {
	r := make([]Endpoint, 0, len(o.endpoints))
	for _, e := range o.endpoints {
		r = append(r, e.toDb())
	}
	return r
}

// RegisterEventHandlers registers event handlers on this database
func (db *MemoryDb) RegisterEventHandlers(system events.EventSystem) {
	system.RegisterEventHandler(events.RegisterVendor, func(e events.Event) error {
		// Unmarshal
		event := events.RegisterVendorEvent{}
		if err := e.Unmarshal(&event); err != nil {
			return err
		}
		// Validate
		id := string(event.Identifier)
		if db.vendors[id] != nil {
			return fmt.Errorf("vendor already registered (id = %s)", event.Identifier)
		}
		// Process
		db.vendors[id] = &vendor{
			RegisterVendorEvent: event,
			orgs:                make(map[string]*org),
		}
		return nil
	})
	system.RegisterEventHandler(events.VendorClaim, func(e events.Event) error {
		// Unmarshal
		event := events.VendorClaimEvent{}
		if err := e.Unmarshal(&event); err != nil {
			return err
		}
		// Validate
		orgID := string(event.OrgIdentifier)
		vendorID := string(event.VendorIdentifier)
		_, err := db.OrganizationById(orgID)
		if err != ErrOrganizationNotFound {
			return fmt.Errorf("organization already registered (id = %s)", event.OrgIdentifier)
		}
		if db.vendors[vendorID] == nil {
			return fmt.Errorf("vendor is not registered (id = %s)", event.VendorIdentifier)
		}
		_, err = keysAsSet(event.OrgKeys)
		if err != nil {
			return errors2.Wrap(err, "invalid JWK")
		}
		// Process
		db.vendors[vendorID].orgs[orgID] = &org{
			VendorClaimEvent: event,
			endpoints:        make(map[string]*endpoint),
		}
		return nil
	})
	system.RegisterEventHandler(events.RegisterEndpoint, func(e events.Event) error {
		// Unmarshal
		event := events.RegisterEndpointEvent{}
		if err := e.Unmarshal(&event); err != nil {
			return err
		}
		// Validate
		orgID := string(event.Organization)
		o := db.lookupOrg(orgID)
		if o == nil {
			return fmt.Errorf("organization not registered (id = %s)", orgID)
		}
		// Process
		o.endpoints[string(event.Identifier)] = &endpoint{
			RegisterEndpointEvent: event,
		}
		return nil
	})

}

func (db *MemoryDb) lookupOrg(orgID string) *org {
	for _, vendor := range db.vendors {
		o := vendor.orgs[orgID]
		if o != nil {
			return o
		}
	}
	return nil
}

func toDbIdentifier(identifier events.Identifier) Identifier {
	return Identifier(string(identifier))
}

func New() *MemoryDb {
	return &MemoryDb{
		make(map[string]*vendor),
	}
}

func (db *MemoryDb) FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]Endpoint, error) {
	o := db.lookupOrg(organizationIdentifier)
	if o == nil {
		return nil, fmt.Errorf("organization with identifier [%s] does not exist", organizationIdentifier)
	}
	var endpoints []Endpoint
	for _, e := range o.endpoints {
		if e.Status == StatusActive {
			if endpointType == nil || *endpointType == e.EndpointType {
				endpoints = append(endpoints, e.toDb())
			}
		}
	}
	return endpoints, nil
}

func (db *MemoryDb) SearchOrganizations(query string) []Organization {

	// all organization names to lowercase and to slice
	// query to slice
	// compare slices: if both match: pop both, if not pop organization slice
	// continue until one is empty
	// if query is empty, match is found
	var matches []Organization
	for _, v := range db.vendors {
		for _, o := range v.orgs {
			if searchRecursive(strings.Split(strings.ToLower(query), ""), strings.Split(strings.ToLower(o.OrgName), "")) {
				matches = append(matches, o.toDb())
			}
		}
	}
	return matches
}

// ErrOrganizationNotFound is returned when an organization is not found
var ErrOrganizationNotFound = errors.New("organization not found")

func (db *MemoryDb) ReverseLookup(name string) (*Organization, error) {
	for _, v := range db.vendors {
		for _, o := range v.orgs {
			if strings.ToLower(name) == strings.ToLower(o.OrgName) {
				r := o.toDb()
				return &r, nil
			}
		}
	}
	return nil, fmt.Errorf("reverse lookup failed for %s: %w", name, ErrOrganizationNotFound)
}

func (db *MemoryDb) OrganizationById(id string) (*Organization, error) {
	org := db.lookupOrg(id)
	if org == nil {
		return nil, fmt.Errorf("%s: %w", id, ErrOrganizationNotFound)
	}
	r := org.toDb()
	return &r, nil
}

func searchRecursive(query []string, orgName []string) bool {
	// search string empty, return match
	if len(query) == 0 {
		return true
	}

	// no more organizations to search for
	if len(orgName) == 0 {
		return false
	}

	if query[0] == orgName[0] {
		return searchRecursive(query[1:], orgName[1:])
	} else {
		return searchRecursive(query, orgName[1:])
	}
}
