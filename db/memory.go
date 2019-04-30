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
	"github.com/spf13/viper"
	"strings"
)

type index struct {
	endpointIndex map[string]generated.Endpoint
	organizationIndex map[string]generated.Organization
	endpointToOrganizationIndex map[string][]generated.EndpointOrganization
	organizationToEndpointIndex map[string][]generated.EndpointOrganization
}

type dbError struct {
	s string
}

func newDbError(text string) error {
	return &dbError{text}
}

func (e *dbError) Error() string {
	return e.s
}

func (i index) appendEO(eo generated.EndpointOrganization) error {
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

func (i index) appendEndpoint(e generated.Endpoint) {
	i.endpointIndex[e.Identifier.String()] = e

	// also create empty slice at this map index
	i.endpointToOrganizationIndex[e.Identifier.String()] = []generated.EndpointOrganization{}
}

func (i index) appendOrganization(o generated.Organization) {
	i.organizationIndex[o.Identifier.String()] = o

	// also create empty slice at this map index
	i.organizationToEndpointIndex[o.Identifier.String()] = []generated.EndpointOrganization{}
}

var memoryDB index

func Load() error {
	location := viper.GetString(CONF_DATA_DIR)

	err := ValidateLocation(location)

	if err != nil {
		return err
	}

	memoryDB = index{
		make(map[string]generated.Endpoint),
		make(map[string]generated.Organization),
		make(map[string][]generated.EndpointOrganization),
		make(map[string][]generated.EndpointOrganization),
	}

	err = loadEndpoints()
	if err != nil {
		return err
	}

	err = loadOrganizations()
	if err != nil {
		return err
	}

	err = loadEndpointsOrganizations()
	return err
}

func FindEndpointsByOrganisation(organizationIdentifier string) ([]generated.Endpoint, error) {

	_, exists := memoryDB.organizationIndex[organizationIdentifier]

	// not found
	if !exists {
		return nil, newDbError(fmt.Sprintf("Organization with identifier [%s] does not exist", organizationIdentifier))
	}

	mappings, exists := memoryDB.organizationToEndpointIndex[organizationIdentifier]

	// not found
	if !exists {
		return nil, newDbError(fmt.Sprintf("Organization with identifier [%s] does not have any endpoints", organizationIdentifier))
	}

	// filter inactive mappings
	filtered := mappings[:0]
	for _, x := range mappings {
		if x.Status == generated.STATUS_ACTIVE {
			filtered = append(filtered, x)
		}
	}

	if len(filtered) == 0 {
		return nil, newDbError(fmt.Sprintf("Organization with identifier [%s] does not have any active endpoints", organizationIdentifier))
	}

	// map to endpoints
	var endpoints []generated.Endpoint
	for _, f := range filtered {
		es := memoryDB.endpointIndex[f.EndpointIdentifier.String()]

		if es.Status == generated.STATUS_ACTIVE {
			endpoints = append(endpoints, es)
		}
	}

	return endpoints, nil
}

func SearchOrganizations(query string) []generated.Organization {

	// all organization names to lowercase and to slice
	// query to slice
	// compare slices: if both match: pop both, if not pop organization slice
	// continue until one is empty
	// if query is empty, match is found
	var matches []generated.Organization
	for _, o := range memoryDB.organizationIndex {
		if searchRecursive(strings.Split(strings.ToLower(query), ""), strings.Split(strings.ToLower(o.Name), "")) {
			matches = append(matches, o)
		}
	}
	return matches
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

func loadEndpoints() error {

	data, err := ReadFile(viper.GetString(CONF_DATA_DIR), FILE_ENDPOINTS)

	if err != nil {
		return err
	}

	var stub []generated.Endpoint
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		memoryDB.appendEndpoint(e)
	}

	glog.V(2).Infof("Added %d Endpoints to db", len(memoryDB.endpointIndex))

	return nil
}

func loadOrganizations() error {

	data, err := ReadFile(viper.GetString(CONF_DATA_DIR), FILE_ORGANIZATIONS)

	if err != nil {
		return err
	}

	var stub []generated.Organization
	err = json.Unmarshal(data, &stub)

	if err != nil {
		return err
	}

	for _, e := range stub {
		memoryDB.appendOrganization(e)
	}

	glog.V(2).Infof("Added %d Organizations to db", len(memoryDB.organizationIndex))

	return nil
}

func loadEndpointsOrganizations() error {
	data, err := ReadFile(viper.GetString(CONF_DATA_DIR), FILE_ENDPOINTS_ORGANIZATIONS)

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
		err = memoryDB.appendEO(e)
		if err != nil {
			return err
		}
	}

	glog.V(2).Infof("Added %d mappings of endpoint <-> organization to db", len(memoryDB.organizationIndex))

	return nil
}
