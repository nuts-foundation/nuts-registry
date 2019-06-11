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
	"testing"
)

var endpoint = Endpoint{
	Identifier:   Identifier("urn:nuts:system:value"),
	EndpointType: "urn:nuts:endpoint::type",
	Status:       StatusActive,
}

var organization = Organization{
	Identifier: Identifier("urn:nuts:system:value"),
	Name:       "test",
}

var mapping = EndpointOrganization{
	Endpoint:     Identifier("urn:nuts:system:value"),
	Organization: Identifier("urn:nuts:system:value"),
	Status:       StatusActive,
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
	t.Run("Complete valid example", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("../../test_data/valid_files")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}
	})

	t.Run("from invalid location does not give error", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("non-existing")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}
	})

	t.Run("Loading from location with missing files gives err", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("../../test_data/missing_files/")

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "../../test_data/missing_files is missing required files: endpoints.json, endpoints_organizations.json"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("Loading from location with invalid endpoints json gives err", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("../../test_data/invalid_files/invalid_endpoints")

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "invalid character '[' looking for beginning of object key string"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("Loading from location with invalid organization json gives err", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("../../test_data/invalid_files/invalid_organizations")

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "invalid character '{' looking for beginning of object key string"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("Loading from location with missing mappings json gives err", func(t *testing.T) {
		validDb := New()
		err := validDb.Load("../../test_data/invalid_files/invalid_mappings")

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "invalid character '{' looking for beginning of object key string"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_RegisterOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()

		err := validDb.RegisterOrganization(organization)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(validDb.organizationIndex) != 1 {
			t.Errorf("Expected 1 entry in db, got: %d", len(validDb.organizationIndex))
		}
	})

	t.Run("duplicate entry", func(t *testing.T) {
		validDb := New()

		validDb.RegisterOrganization(organization)
		err := validDb.RegisterOrganization(organization)

		if err == nil {
			t.Errorf("Expected error, got nothing")
			return
		}

		expected := "Duplicate organization for id urn:nuts:system:value"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got: [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_RemoveOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		err := validDb.RemoveOrganization(string(organization.Identifier))

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(validDb.organizationIndex) != 0 {
			t.Errorf("Expected 0 entry in db, got: %d", len(validDb.organizationIndex))
		}
	})

	t.Run("unknown entry", func(t *testing.T) {
		validDb := New()

		err := validDb.RemoveOrganization(string(organization.Identifier))

		if err == nil {
			t.Errorf("Expected error, got nothing")
			return
		}

		expected := "Unknown organization with id urn:nuts:system:value"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got: [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)
		validDb.appendEndpoint(&endpoint)
		validDb.appendEO(mapping)

		result, err := validDb.FindEndpointsByOrganization("urn:nuts:system:value")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("Inactive mappings are not returned", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)
		validDb.appendEndpoint(&endpoint)
		mappingCopy := EndpointOrganization(mapping)
		mappingCopy.Status = "inactive"
		validDb.appendEO(mappingCopy)

		result, err := validDb.FindEndpointsByOrganization("urn:nuts:system:value")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})

	t.Run("Inactive organizations are not returned", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)
		endpointCopy := Endpoint(endpoint)
		endpointCopy.Status = "inactive"
		validDb.appendEndpoint(&endpointCopy)
		validDb.appendEO(mapping)

		result, err := validDb.FindEndpointsByOrganization("urn:nuts:system:value")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})
}

func TestMemoryDb_appendEO(t *testing.T) {
	t.Run("Appending mappings for unknown endpoint gives err", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)
		validDb.appendEO(mapping)

		err := validDb.appendEO(mapping)

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "Endpoint <> Organization mapping references unknown endpoint with identifier [urn:nuts:system:value]"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
		}
	})

	t.Run("Appending mappings for unknown organization gives err", func(t *testing.T) {
		validDb := New()
		validDb.appendEndpoint(&endpoint)
		validDb.appendEO(mapping)

		err := validDb.appendEO(mapping)

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "Endpoint <> Organization mapping references unknown organization with identifier [urn:nuts:system:value]"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_FindEndpointsByOrganizationNoMapping(t *testing.T) {
	validDb := New()
	validDb.appendOrganization(&organization)
	validDb.appendEndpoint(&endpoint)

	result, err := validDb.FindEndpointsByOrganization("urn:nuts:system:value")

	if err != nil {
		t.Errorf("Expected no error, got: %s", err.Error())
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 result, got: %d", len(result))
	}
}

func TestMemoryDb_FindEndpointsByOrganizationUnknown(t *testing.T) {
	validDb := New()

	_, err := validDb.FindEndpointsByOrganization("urn:nuts:system:value")

	if err == nil {
		t.Errorf("Expected error")
	}

	expected := "Organization with identifier [urn:nuts:system:value] does not exist"
	if err.Error() != expected {
		t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
	}
}

func TestMemoryDb_SearchOrganizations(t *testing.T) {
	t.Run("complete valid example", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		result := validDb.SearchOrganizations("test")

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("partial match returns organization", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		result := validDb.SearchOrganizations("ts")

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("searching for unknown organization returns empty list", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		result := validDb.SearchOrganizations("tset")

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})
}

func TestMemoryDb_OrganizationById(t *testing.T) {
	t.Run("organization is found", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		result, err := validDb.OrganizationById("urn:nuts:system:value")

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if result.Name != "test" {
			t.Errorf("Expected 1 result with name test, got: %s", result.Name)
		}
	})

	t.Run("organization is not found", func(t *testing.T) {
		validDb := New()
		validDb.appendOrganization(&organization)

		_, err := validDb.OrganizationById("test")

		expected := "organization not found"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
		}
	})
}
