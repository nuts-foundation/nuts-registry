/*
 * Nuts registry
 * Copyright (C) 2019 Nuts community
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
 */

package db

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/nuts-foundation/nuts-registry/generated"
	"strings"
)

type MemoryDb struct {
	endpointIndex               map[string]generated.Endpoint
	organizationIndex           map[string]generated.Organization
	endpointToOrganizationIndex map[string][]generated.EndpointOrganization
	organizationToEndpointIndex map[string][]generated.EndpointOrganization
}

func New() *MemoryDb {
	return &MemoryDb{
		make(map[string]generated.Endpoint),
		make(map[string]generated.Organization),
		make(map[string][]generated.EndpointOrganization),
		make(map[string][]generated.EndpointOrganization),
	}
}

func (i *MemoryDb) appendEO(eo generated.EndpointOrganization) error {
	ois := eo.OrganizationIdentifier.String()
	eis := eo.EndpointIdentifier.String()

	_, f := i.organizationToEndpointIndex[ois]
	if !f {
		return newDbError(fmt.Sprintf("Endpoint <> Organization mapping references unknown organization with identifier [%s]", ois))
	}

	_, f = i.endpointToOrganizationIndex[eis]
	if !f {
		return newDbError(fmt.Sprintf("Endpoint <> Organization mapping references unknown endpoint with identifier [%s]", ois))
	}

	i.organizationToEndpointIndex[ois] = append(i.organizationToEndpointIndex[ois], eo)
	i.endpointToOrganizationIndex[eis] = append(i.endpointToOrganizationIndex[eis], eo)

	return nil
}

func (i *MemoryDb) appendEndpoint(e generated.Endpoint) {
	i.endpointIndex[e.Identifier.String()] = e

	// also create empty slice at this map Db
	i.endpointToOrganizationIndex[e.Identifier.String()] = []generated.EndpointOrganization{}
}

func (i *MemoryDb) appendOrganization(o generated.Organization) {
	i.organizationIndex[o.Identifier.String()] = o

	// also create empty slice at this map Db
	i.organizationToEndpointIndex[o.Identifier.String()] = []generated.EndpointOrganization{}
}

func (db *MemoryDb) Load(location string) error {
	err := validateLocation(location)

	if err != nil {
		return err
	}

	err = db.loadEndpoints(location)
	if err != nil {
		return err
	}

	err = db.loadOrganizations(location)
	if err != nil {
		return err
	}

	err = db.loadEndpointsOrganizations(location)
	return err
}

func (db *MemoryDb) FindEndpointsByOrganization(organizationIdentifier string) ([]generated.Endpoint, error) {

	_, exists := db.organizationIndex[organizationIdentifier]

	// not found
	if !exists {
		return nil, newDbError(fmt.Sprintf("Organization with identifier [%s] does not exist", organizationIdentifier))
	}

	mappings := db.organizationToEndpointIndex[organizationIdentifier]

	// filter inactive mappings
	filtered := mappings[:0]
	for _, x := range mappings {
		if x.Status == generated.StatusActive {
			filtered = append(filtered, x)
		}
	}

	if len(filtered) == 0 {
		return []generated.Endpoint{}, nil
	}

	// map to endpoints
	var endpoints []generated.Endpoint
	for _, f := range filtered {
		es := db.endpointIndex[f.EndpointIdentifier.String()]

		if es.Status == generated.StatusActive {
			endpoints = append(endpoints, es)
		}
	}

	return endpoints, nil
}

func (db *MemoryDb) SearchOrganizations(query string) []generated.Organization {

	// all organization names to lowercase and to slice
	// query to slice
	// compare slices: if both match: pop both, if not pop organization slice
	// continue until one is empty
	// if query is empty, match is found
	var matches []generated.Organization
	for _, o := range db.organizationIndex {
		if searchRecursive(strings.Split(strings.ToLower(query), ""), strings.Split(strings.ToLower(o.Name), "")) {
			matches = append(matches, o)
		}
	}
	return matches
}

func (db *MemoryDb) OrganizationById(id string) (*generated.Organization, error) {

	for _, o := range db.organizationIndex {
		if id == o.Identifier.String() {
			return &o, nil
		}
	}
	return nil, newDbError("organization not found")
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

func (db *MemoryDb) loadEndpoints(location string) error {

	data, err := ReadFile(location, fileEndpoints)

	if err != nil {
		return err
	}

	var stub []generated.Endpoint
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		db.appendEndpoint(e)
	}

	glog.V(2).Infof("Added %d Endpoints to db", len(db.endpointIndex))

	return nil
}

func (db *MemoryDb) loadOrganizations(location string) error {

	data, err := ReadFile(location, fileOrganizations)

	if err != nil {
		return err
	}

	var stub []generated.Organization
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		db.appendOrganization(e)
	}

	glog.V(2).Infof("Added %d Organizations to db", len(db.organizationIndex))

	return nil
}

func (db *MemoryDb) loadEndpointsOrganizations(location string) error {
	data, err := ReadFile(location, fileEndpointsOrganizations)

	if err != nil {
		return err
	}

	var stub []generated.EndpointOrganization
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		// all map values should be present, appending
		err = db.appendEO(e)
		if err != nil {
			return err
		}
	}

	glog.V(2).Infof("Added %d mappings of endpoint <-> organization to db", len(db.organizationIndex))

	return nil
}
