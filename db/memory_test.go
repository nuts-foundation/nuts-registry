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
	"github.com/nuts-foundation/nuts-registry/generated"
	"testing"
)

var endpoint = generated.Endpoint{
	Identifier:generated.Identifier{System:"system", Value:"value"},
	EndpointType: "type#value",
	Status: generated.STATUS_ACTIVE,
}

var organization = generated.Organization {
	Identifier:generated.Identifier{System:"system", Value:"value"},
	Name: "test",
}

var mapping = generated.EndpointOrganization {
	EndpointIdentifier: generated.Identifier{System:"system", Value:"value"},
	OrganizationIdentifier: generated.Identifier{System:"system", Value:"value"},
	Status: generated.STATUS_ACTIVE,
}

func TestNew(t *testing.T) {
	emptyDb := New()

	if len(emptyDb.organizationIndex) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.organizationIndex))
	}

	if len(emptyDb.endpointIndex) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.endpointIndex))
	}

	if len(emptyDb.endpointToOrganizationIndex) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.endpointToOrganizationIndex))
	}

	if len(emptyDb.organizationToEndpointIndex) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.organizationToEndpointIndex))
	}
}

func TestMemoryDb_Load(t *testing.T) {
	validDb := New()
	err := validDb.Load("../test_data/valid_files")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}
}

func TestMemoryDb_LoadInvalidLocation(t *testing.T) {
	validDb := New()
	err := validDb.Load("../test_data/missing_files/")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "../test_data/missing_files is missing required files: endpoints.json, endpoints_organizations.json"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got [%s]", expected, err.Error())
	}
}

func TestMemoryDb_LoadInvalidEndpoints(t *testing.T) {
	validDb := New()
	err := validDb.Load("../test_data/invalid_files/invalid_endpoints")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "invalid character '[' looking for beginning of object key string"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got [%s]", expected, err.Error())
	}
}

func TestMemoryDb_LoadInvalidOrganizations(t *testing.T) {
	validDb := New()
	err := validDb.Load("../test_data/invalid_files/invalid_organizations")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "invalid character '{' looking for beginning of object key string"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got [%s]", expected, err.Error())
	}
}

func TestMemoryDb_LoadInvalidMappings(t *testing.T) {
	validDb := New()
	err := validDb.Load("../test_data/invalid_files/invalid_mappings")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "invalid character '{' looking for beginning of object key string"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got [%s]", expected, err.Error())
	}
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)
	validDb.appendEndpoint(endpoint)
	validDb.appendEO(mapping)

	result, err := validDb.FindEndpointsByOrganization("system#value")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got: %d", len(result))
	}
}

func TestMemoryDb_FindEndpointsByOrganizationInactiveMapping(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)
	validDb.appendEndpoint(endpoint)
	mappingCopy := generated.EndpointOrganization(mapping)
	mappingCopy.Status = "inactive"
	validDb.appendEO(mappingCopy)

	result, err := validDb.FindEndpointsByOrganization("system#value")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 result, got: %d", len(result))
	}
}

func TestMemoryDb_FindEndpointsByOrganizationInactiveEndpoint(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)
	endpointCopy := generated.Endpoint(endpoint)
	endpointCopy.Status = "inactive"
	validDb.appendEndpoint(endpointCopy)
	validDb.appendEO(mapping)

	result, err := validDb.FindEndpointsByOrganization("system#value")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 result, got: %d", len(result))
	}
}

func TestMemoryDb_appendEOMissingEndpoints(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)
	validDb.appendEO(mapping)

	err := validDb.appendEO(mapping)

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "Endpoint <> Organization mapping references unknown endpoint with identifier [system#value]"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got: [%s]", expected,  err.Error())
	}
}

func TestMemoryDb_appendEOMissingOrg(t *testing.T) {
	validDb := New()
	validDb.appendEndpoint(endpoint)
	validDb.appendEO(mapping)

	err := validDb.appendEO(mapping)

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "Endpoint <> Organization mapping references unknown organization with identifier [system#value]"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got: [%s]", expected,  err.Error())
	}
}

func TestMemoryDb_FindEndpointsByOrganizationNoMapping(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)
	validDb.appendEndpoint(endpoint)

	result, err := validDb.FindEndpointsByOrganization("system#value")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 result, got: %d", len(result))
	}
}

func TestMemoryDb_FindEndpointsByOrganizationUnknown(t *testing.T) {
	validDb := New()

	_, err := validDb.FindEndpointsByOrganization("system#value")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "Organization with identifier [system#value] does not exist"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got: [%s]", expected,  err.Error())
	}
}

func TestMemoryDb_SearchOrganizations(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)

	result := validDb.SearchOrganizations("test")

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got: %d", len(result))
	}
}

func TestMemoryDb_SearchOrganizationsPartialHit(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)

	result := validDb.SearchOrganizations("ts")

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got: %d", len(result))
	}
}

func TestMemoryDb_SearchOrganizationsUnknown(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(organization)

	result := validDb.SearchOrganizations("tset")

	if len(result) != 0 {
		t.Errorf("Expected 0 result, got: %d", len(result))
	}
}