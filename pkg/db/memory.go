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
	"strings"

	"github.com/sirupsen/logrus"
)

type MemoryDb struct {
	endpointIndex               map[string]*Endpoint
	organizationIndex           map[string]*Organization
	endpointToOrganizationIndex map[string][]EndpointOrganization
	organizationToEndpointIndex map[string][]EndpointOrganization
}

// ErrDuplicateOrganization is given for an organization with an identifier that has already been stored
var ErrDuplicateOrganization = errors.New("duplicate organization")

// ErrUnknownOrganization is given when an organization does not exist for given identifier
var ErrUnknownOrganization = errors.New("unknown organization")

func (db *MemoryDb) RegisterOrganization(org Organization) error {
	o := db.organizationIndex[string(org.Identifier)]

	if o != nil {
		return fmt.Errorf("error registering organization with id %s: %w", org.Identifier, ErrDuplicateOrganization)
	}

	// also validate the keys parsed from json
	if _, err := org.KeysAsSet(); err != nil {
		return fmt.Errorf("error registering organization with id %s: %w", org.Identifier, err)
	}

	db.appendOrganization(&org)

	for _, e := range org.Endpoints {
		db.RegisterEndpoint(e)
		err := db.RegisterEndpointOrganization(
			EndpointOrganization{
				Status:       StatusActive,
				Organization: org.Identifier,
				Endpoint:     e.Identifier,
			})

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *MemoryDb) RemoveOrganization(id string) error {
	o := db.organizationIndex[id]
	eos := db.organizationToEndpointIndex[id]

	if o == nil {
		return fmt.Errorf("error removing organization with id %s: %w", id, ErrUnknownOrganization)
	}

	for _, v := range eos {
		delete(db.endpointIndex, string(v.Endpoint))
		delete(db.endpointToOrganizationIndex, string(v.Endpoint))
	}

	delete(db.organizationToEndpointIndex, id)
	delete(db.organizationIndex, id)

	return nil
}

func New() *MemoryDb {
	return &MemoryDb{
		make(map[string]*Endpoint),
		make(map[string]*Organization),
		make(map[string][]EndpointOrganization),
		make(map[string][]EndpointOrganization),
	}
}

func (db *MemoryDb) RegisterEndpointOrganization(eo EndpointOrganization) error {
	ois := eo.Organization.String()
	eis := eo.Endpoint.String()

	_, f := db.organizationToEndpointIndex[ois]
	if !f {
		return fmt.Errorf("Endpoint <> Organization mapping references unknown organization with identifier [%s]", ois)
	}

	_, f = db.endpointToOrganizationIndex[eis]
	if !f {
		return fmt.Errorf("Endpoint <> Organization mapping references unknown endpoint with identifier [%s]", eis)
	}

	db.organizationToEndpointIndex[ois] = append(db.organizationToEndpointIndex[ois], eo)
	db.endpointToOrganizationIndex[eis] = append(db.endpointToOrganizationIndex[eis], eo)

	logrus.Tracef("Added mapping between: %s <-> %s", ois, eis)

	return nil
}

func (db *MemoryDb) RegisterEndpoint(e Endpoint) {
	cp := &e
	db.endpointIndex[e.Identifier.String()] = cp

	// also create empty slice at this map R
	db.endpointToOrganizationIndex[e.Identifier.String()] = []EndpointOrganization{}

	logrus.Tracef("Added endpoint: %s", e.Identifier)
}

func (db *MemoryDb) appendOrganization(o *Organization) {
	cp := *o
	db.organizationIndex[o.Identifier.String()] = &cp

	// also create empty slice at this map R
	db.organizationToEndpointIndex[o.Identifier.String()] = []EndpointOrganization{}

	logrus.Tracef("Added Organization: %s", o.Identifier)
}

func (db *MemoryDb) FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]Endpoint, error) {

	_, exists := db.organizationIndex[organizationIdentifier]

	// not found
	if !exists {
		return nil, fmt.Errorf("Organization with identifier [%s] does not exist", organizationIdentifier)
	}

	mappings := db.organizationToEndpointIndex[organizationIdentifier]

	// filter inactive mappings
	filtered := mappings[:0]
	for _, x := range mappings {
		if x.Status == StatusActive {
			filtered = append(filtered, x)
		}
	}

	if len(filtered) == 0 {
		return []Endpoint{}, nil
	}

	// map to endpoints
	var endpoints []Endpoint
	for _, f := range filtered {
		es := db.endpointIndex[f.Endpoint.String()]

		if es.Status == StatusActive {
			if endpointType != nil {
				if *endpointType == es.EndpointType {
					endpoints = append(endpoints, *es)
				}
			} else {
				endpoints = append(endpoints, *es)
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
	for _, o := range db.organizationIndex {
		if searchRecursive(strings.Split(strings.ToLower(query), ""), strings.Split(strings.ToLower(o.Name), "")) {
			matches = append(matches, *o)
		}
	}
	return matches
}

// ErrOrganizationNotFound is returned when an organization is not found
var ErrOrganizationNotFound = errors.New("organization not found")

func (db *MemoryDb) ReverseLookup(name string) (*Organization, error) {
	for _, o := range db.organizationIndex {
		if strings.ToLower(name) == strings.ToLower(o.Name) {
			return o, nil
		}
	}

	return nil, ErrOrganizationNotFound
}

func (db *MemoryDb) OrganizationById(id string) (*Organization, error) {

	for _, o := range db.organizationIndex {
		if id == o.Identifier.String() {
			return o, nil
		}
	}
	return nil, ErrOrganizationNotFound
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

