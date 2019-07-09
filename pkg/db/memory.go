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
	"github.com/sirupsen/logrus"
	"strings"
)

type MemoryDb struct {
	endpointIndex               map[string]*Endpoint
	organizationIndex           map[string]*Organization
	endpointToOrganizationIndex map[string][]EndpointOrganization
	organizationToEndpointIndex map[string][]EndpointOrganization
}

func (db *MemoryDb) RegisterOrganization(org Organization) error {
	o := db.organizationIndex[string(org.Identifier)]

	if o != nil {
		return newDbError(fmt.Sprintf("Duplicate organization for id %s", org.Identifier))
	}

	db.appendOrganization(&org)

	for _, e := range org.Endpoints {
		db.appendEndpoint(&e)
		err := db.appendEO(
			EndpointOrganization{
				Status:StatusActive,
				Organization:org.Identifier,
				Endpoint:e.Identifier,
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
		return newDbError(fmt.Sprintf("Unknown organization with id %s", id))
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



func (i *MemoryDb) appendEO(eo EndpointOrganization) error {
	ois := eo.Organization.String()
	eis := eo.Endpoint.String()

	_, f := i.organizationToEndpointIndex[ois]
	if !f {
		return newDbError(fmt.Sprintf("Endpoint <> Organization mapping references unknown organization with identifier [%s]", ois))
	}

	_, f = i.endpointToOrganizationIndex[eis]
	if !f {
		return newDbError(fmt.Sprintf("Endpoint <> Organization mapping references unknown endpoint with identifier [%s]", eis))
	}

	i.organizationToEndpointIndex[ois] = append(i.organizationToEndpointIndex[ois], eo)
	i.endpointToOrganizationIndex[eis] = append(i.endpointToOrganizationIndex[eis], eo)

	logrus.Tracef("Added mapping between: %s <-> %s", ois, eis)

	return nil
}

func (i *MemoryDb) appendEndpoint(e *Endpoint) {
	cp := *e
	i.endpointIndex[e.Identifier.String()] = &cp

	// also create empty slice at this map R
	i.endpointToOrganizationIndex[e.Identifier.String()] = []EndpointOrganization{}

	logrus.Tracef("Added endpoint: %s", e.Identifier)
}

func (i *MemoryDb) appendOrganization(o *Organization) {
	cp := *o
	i.organizationIndex[o.Identifier.String()] = &cp

	// also create empty slice at this map R
	i.organizationToEndpointIndex[o.Identifier.String()] = []EndpointOrganization{}

	logrus.Tracef("Added Organization: %s", o.Identifier)
}

// Load the db files from the configured datadir
func (db *MemoryDb) Load(location string) error {
	err := validateLocation(location)

	if err != nil {
		if err.Error() == fmt.Sprintf("open %s: no such file or directory", location) {
			logrus.Warnf("No database files found at %s, starting with empty registry", location)
			return nil
		}
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

func (db *MemoryDb) FindEndpointsByOrganization(organizationIdentifier string) ([]Endpoint, error) {

	_, exists := db.organizationIndex[organizationIdentifier]

	// not found
	if !exists {
		return nil, newDbError(fmt.Sprintf("Organization with identifier [%s] does not exist", organizationIdentifier))
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
			endpoints = append(endpoints, *es)
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

func (db *MemoryDb) OrganizationById(id string) (*Organization, error) {

	for _, o := range db.organizationIndex {
		if id == o.Identifier.String() {
			return o, nil
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

	var stub []Endpoint
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		db.appendEndpoint(&e)
	}

	logrus.Infof("Added %d Endpoints to db", len(db.endpointIndex))

	return nil
}

func (db *MemoryDb) loadOrganizations(location string) error {

	data, err := ReadFile(location, fileOrganizations)

	if err != nil {
		return err
	}

	var stub []Organization
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		// replace pub key \n
		if e.PublicKey != nil {
			s := strings.ReplaceAll(*e.PublicKey, "\\n", "\n")
			e.PublicKey = &s
		}

		db.appendOrganization(&e)
	}

	logrus.Infof("Added %d Organizations to db", len(db.organizationIndex))

	return nil
}

func (db *MemoryDb) loadEndpointsOrganizations(location string) error {
	data, err := ReadFile(location, fileEndpointsOrganizations)

	if err != nil {
		return err
	}

	var stub []EndpointOrganization
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

	logrus.Infof("Added %d mappings of endpoint <-> organization to db", len(stub))

	return nil
}
