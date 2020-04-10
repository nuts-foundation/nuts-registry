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
	"crypto/rsa"
	"errors"
	"fmt"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"strings"
)

type MemoryDb struct {
	vendors map[string]*vendor
}

type vendor struct {
	domain.RegisterVendorEvent
	orgs map[string]*org
}

type org struct {
	domain.VendorClaimEvent
	endpoints map[string]*endpoint
}

type endpoint struct {
	domain.RegisterEndpointEvent
}

func (o org) toDb() Organization {
	result := Organization{
		Identifier: toDbIdentifier(o.OrgIdentifier),
		Name:       o.OrgName,
		Keys:       o.OrgKeys,
		Endpoints:  o.toDbEndpoints(),
	}
	// Backwards compatibility for deprecated PublicKey property: fill with first RSA key we can find
	for _, k := range o.OrgKeys {
		keyAsJwk, _ := crypto.MapToJwk(k.(map[string]interface{}))
		if keyAsJwk != nil {
			matKey, _ := keyAsJwk.Materialize()
			pubKey, ok := matKey.(*rsa.PublicKey)
			if ok {
				p, _ := crypto.PublicKeyToPem(pubKey)
				result.PublicKey = &p
			}
		}
	}
	return result
}

func (v vendor) toDb() Vendor {
	return Vendor{
		Identifier: toDbIdentifier(v.Identifier),
		Name:       v.Name,
		Domain:     v.Domain,
		Keys:       v.Keys,
	}
}

func (e endpoint) toDb() Endpoint {
	return Endpoint{
		URL:          e.URL,
		Organization: toDbIdentifier(e.Organization),
		EndpointType: e.EndpointType,
		Identifier:   toDbIdentifier(e.Identifier),
		Status:       e.Status,
		Properties:   e.Properties,
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
func (db *MemoryDb) RegisterEventHandlers(fn events.EventRegistrar) {
	fn(domain.RegisterVendor, func(e events.Event) error {
		// Unmarshal
		event := domain.RegisterVendorEvent{}
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
	fn(domain.VendorClaim, func(e events.Event) error {
		// Unmarshal
		event := domain.VendorClaimEvent{}
		if err := e.Unmarshal(&event); err != nil {
			return err
		}
		// Validate
		orgID := string(event.OrgIdentifier)
		vendorID := string(event.VendorIdentifier)
		_, err := db.OrganizationById(orgID)
		if !errors.Is(err, ErrOrganizationNotFound) {
			return fmt.Errorf("organization already registered (id = %s)", event.OrgIdentifier)
		}
		if db.vendors[vendorID] == nil {
			return fmt.Errorf("vendor is not registered (id = %s)", event.VendorIdentifier)
		}
		// Process
		db.vendors[vendorID].orgs[orgID] = &org{
			VendorClaimEvent: event,
			endpoints:        make(map[string]*endpoint),
		}
		return nil
	})
	fn(domain.RegisterEndpoint, func(e events.Event) error {
		// Unmarshal
		event := domain.RegisterEndpointEvent{}
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

func toDbIdentifier(identifier domain.Identifier) Identifier {
	return Identifier(string(identifier))
}

func New() *MemoryDb {
	return &MemoryDb{
		make(map[string]*vendor),
	}
}

// VendorByID looks up the vendor by the given ID.
func (db *MemoryDb) VendorByID(id string) *Vendor {
	if db.vendors[id] == nil {
		return nil
	}
	result := db.vendors[id].toDb()
	return &result
}

func (db *MemoryDb) OrganizationsByVendorID(id string) []*Organization {
	vendor := db.vendors[id]
	if vendor == nil {
		return nil
	}
	orgs := make([]*Organization, 0, len(vendor.orgs))
	for _, org := range vendor.orgs {
		o := org.toDb()
		orgs = append(orgs, &o)
	}
	return orgs
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
