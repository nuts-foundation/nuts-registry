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
	"testing"

	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

var endpoint = Endpoint{
	Identifier:   Identifier("urn:nuts:system:value"),
	EndpointType: "urn:nuts:endpoint:type",
	Status:       StatusActive,
}

var organization = Organization{
	Identifier: Identifier("urn:nuts:system:value"),
	Name:       "test",
}

var hiddenOrganization = Organization{
	Identifier: Identifier("urn:nuts:hidden"),
	Name:       "hidden",
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

func TestMemoryDb_RegisterOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()

		err := validDb.RegisterOrganization(organization)

		if assert.NoError(t, err) {
			assert.Len(t, validDb.organizationIndex, 1)
		}
	})

	t.Run("organization with invalid key set", func(t *testing.T) {
		validDb := New()

		o := Organization{
			Identifier: Identifier(random.String(8)),
			Name:       random.String(8),
			Keys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		}
		err := validDb.RegisterOrganization(o)

		assert.Error(t, err)
	})

	t.Run("duplicate entry", func(t *testing.T) {
		validDb := New()

		assert.NoError(t, validDb.RegisterOrganization(organization))
		err := validDb.RegisterOrganization(organization)

		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, ErrDuplicateOrganization))
		}
	})
}

func TestMemoryDb_RemoveOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))

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

		if !errors.Is(err, ErrUnknownOrganization) {
			t.Errorf("Expected error [%v], got: [%v]", ErrUnknownOrganization, err)
		}
	})
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {

	t.Run("Valid example", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		validDb.RegisterEndpoint(endpoint)
		assert.NoError(t, validDb.RegisterEndpointOrganization(mapping))

		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", nil)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("Valid example with type", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		validDb.RegisterEndpoint(endpoint)
		assert.NoError(t, validDb.RegisterEndpointOrganization(mapping))

		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", &endpoint.EndpointType)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("incorrect type", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		validDb.RegisterEndpoint(endpoint)
		assert.NoError(t, validDb.RegisterEndpointOrganization(mapping))

		unknown := "unknown type"
		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", &unknown)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})

	t.Run("Inactive mappings are not returned", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		validDb.RegisterEndpoint(endpoint)
		mappingCopy := EndpointOrganization(mapping)
		mappingCopy.Status = "inactive"
		assert.NoError(t, validDb.RegisterEndpointOrganization(mappingCopy))

		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", nil)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})

	t.Run("Inactive organizations are not returned", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		endpointCopy := Endpoint(endpoint)
		endpointCopy.Status = "inactive"
		validDb.RegisterEndpoint(endpointCopy)
		assert.NoError(t, validDb.RegisterEndpointOrganization(mapping))

		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", nil)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})

	t.Run("no mapping", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		validDb.RegisterEndpoint(endpoint)

		result, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", nil)

		if err != nil {
			t.Errorf("Expected no error, got: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})

	t.Run("Organization unknown", func(t *testing.T) {
		validDb := New()

		_, err := validDb.FindEndpointsByOrganizationAndType("urn:nuts:system:value", nil)

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "organization with identifier [urn:nuts:system:value] does not exist"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_RegisterEndpointOrganization(t *testing.T) {
	t.Run("Appending mappings for unknown endpoint gives err", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))

		err := validDb.RegisterEndpointOrganization(mapping)

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
		validDb.RegisterEndpoint(endpoint)

		err := validDb.RegisterEndpointOrganization(mapping)

		if err == nil {
			t.Errorf("Expected error")
		}

		expected := "Endpoint <> Organization mapping references unknown organization with identifier [urn:nuts:system:value]"
		if err.Error() != expected {
			t.Errorf("Expected [%s], got: [%s]", expected, err.Error())
		}
	})
}

func TestMemoryDb_SearchOrganizations(t *testing.T) {
	t.Run("complete valid example", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		assert.NoError(t, validDb.RegisterOrganization(hiddenOrganization))

		result := validDb.SearchOrganizations("test")

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("partial match returns organization", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))

		result := validDb.SearchOrganizations("ts")

		if len(result) != 1 {
			t.Errorf("Expected 1 result, got: %d", len(result))
		}
	})

	t.Run("wide match returns 2 organization", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))
		assert.NoError(t, validDb.RegisterOrganization(hiddenOrganization))

		result := validDb.SearchOrganizations("e")

		if len(result) != 2 {
			t.Errorf("Expected 2 result, got: %d", len(result))
		}

		if result[0].Name == result[1].Name {
			t.Error("Expected 2 unique results")
		}
	})

	t.Run("searching for unknown organization returns empty list", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))

		result := validDb.SearchOrganizations("tset")

		if len(result) != 0 {
			t.Errorf("Expected 0 result, got: %d", len(result))
		}
	})
}

func TestMemoryDb_ReverseLookup(t *testing.T) {
	validDb := New()
	assert.NoError(t, validDb.RegisterOrganization(organization))

	t.Run("finds exact match", func(t *testing.T) {
		result, err := validDb.ReverseLookup("test")

		assert.Nil(t, err)
		assert.NotNil(t, result)
	})

	t.Run("finds exact match, case insensitive", func(t *testing.T) {
		result, err := validDb.ReverseLookup("TEST")

		assert.Nil(t, err)
		assert.NotNil(t, result)
	})

	t.Run("does not fn partial match", func(t *testing.T) {
		result, err := validDb.ReverseLookup("tst")

		assert.Nil(t, result)
		if assert.NotNil(t, err) {
			assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		}
	})
}

func TestMemoryDb_OrganizationById(t *testing.T) {
	t.Run("organization is found", func(t *testing.T) {
		validDb := New()
		assert.NoError(t, validDb.RegisterOrganization(organization))

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
		assert.NoError(t, validDb.RegisterOrganization(organization))

		_, err := validDb.OrganizationById("test")
		assert.True(t, errors.Is(err, ErrOrganizationNotFound))
	})
}
