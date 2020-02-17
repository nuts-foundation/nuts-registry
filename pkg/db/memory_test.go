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
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test data:
// v1 Vendor Uno
//   o1 Organization Uno
//     e1 Endpoint Uno
//     e2 Endpoint Dos
//   o2 Organization Dos
// v2 Vendor Dos

var registerVendor1 = events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
	Identifier: "v1",
	Name:       "Vendor Uno",
})
var registerVendor2 = events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
	Identifier: "v2",
	Name:       "Vendor Dos",
})

var vendorClaim1 = events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{
	VendorIdentifier: "v1",
	OrgIdentifier:    "o1",
	OrgName:          "Organization Uno",
	OrgKeys:          nil,
})

var vendorClaim2 = events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{
	VendorIdentifier: "v1",
	OrgIdentifier:    "o2",
	OrgName:          "Organization Dos",
	OrgKeys:          nil,
})

var registerEndpoint1 = events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
	Organization: "o1",
	URL:          "foo:bar",
	EndpointType: "simple",
	Identifier:   "e1",
	Status:       StatusActive,
	Version:      "1.0",
})
var registerEndpoint2 = events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
	Organization: "o1",
	URL:          "foo:bar",
	EndpointType: "simple",
	Identifier:   "e2",
	Status:       "inactive",
	Version:      "1.0",
})

//var endpoint = Endpoint{
//	Identifier:   Identifier("urn:nuts:system:value"),
//	EndpointType: "urn:nuts:endpoint:type",
//	Status:       StatusActive,
//}
//
//var organization = Organization{
//	Identifier: Identifier("urn:nuts:system:value"),
//	Name:       "test",
//}
//
//var hiddenOrganization = Organization{
//	Identifier: Identifier("urn:nuts:hidden"),
//	Name:       "hidden",
//}
//
//var mapping = EndpointOrganization{
//	Endpoint:     Identifier("urn:nuts:system:value"),
//	Organization: Identifier("urn:nuts:system:value"),
//	Status:       StatusActive,
//}

func TestNew(t *testing.T) {
	emptyDb := New()

	if len(emptyDb.vendors) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.vendors))
	}
}

func initDb(repo test.TestRepo) (events.EventSystem, *MemoryDb) {
	db := New()
	eventSystem := events.NewEventSystem()
	eventSystem.Configure(repo.Directory + "/events")
	db.RegisterEventHandlers(eventSystem)
	return eventSystem, db
}

func pub(t *testing.T, eventSystem events.EventSystem, events ...events.Event) bool {
	for _, e := range events {
		err := eventSystem.PublishEvent(e)
		if !assert.NoError(t, err) {
			return false
		}
	}
	return true
}

func TestMemoryDb_RegisterVendor(t *testing.T) {
	t.Run("valid example", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors, 1)
		err = eventSystem.PublishEvent(registerVendor2)
		if assert.NoError(t, err) {
			assert.Len(t, db.vendors, 2)
		}
	})

	t.Run("duplicate entry", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, _ := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(registerVendor1)
		assert.Error(t, err)
	})
}

func TestMemoryDb_VendorClaim(t *testing.T) {
	t.Run("valid example", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors["v1"].orgs, 0)
		err = eventSystem.PublishEvent(vendorClaim1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors["v1"].orgs, 1)
		err = eventSystem.PublishEvent(vendorClaim2)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors["v1"].orgs, 2)
	})

	t.Run("organization with invalid key set", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}

		e := events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{
			VendorIdentifier: "v1",
			OrgKeys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		err = eventSystem.PublishEvent(e)
		assert.Error(t, err)
		assert.Len(t, db.vendors["v1"].orgs, 0)
	})

	t.Run("unknown vendor", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, _ := initDb(*repo)
		err = eventSystem.PublishEvent(vendorClaim1)
		assert.Error(t, err)
	})

	t.Run("duplicate organization", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, _ := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(vendorClaim1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(vendorClaim1)
		assert.Error(t, err)
	})
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {
	t.Run("Valid example", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	})

	t.Run("Valid example with type", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		et := "simple"
		result, err := db.FindEndpointsByOrganizationAndType("o1", &et)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	})

	t.Run("incorrect type", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		et := "unknown"
		result, err := db.FindEndpointsByOrganizationAndType("o1", &et)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 0)
	})

	t.Run("Inactive mappings are not returned", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1, registerEndpoint2) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	})

	t.Run("no mapping", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 0)
	})

	t.Run("Organization unknown", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		err = eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.Error(t, err) {
			return
		}
		assert.Len(t, result, 0)
	})
}

func TestMemoryDb_RegisterEndpoint(t *testing.T) {
	t.Run("unknown organization", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, _ := initDb(*repo)
		err = eventSystem.PublishEvent(registerEndpoint1)
		assert.Error(t, err)
	})
}

func TestMemoryDb_SearchOrganizations(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	eventSystem, db := initDb(*repo)
	if !pub(t, eventSystem, registerVendor1, vendorClaim1, vendorClaim2) {
		return
	}

	t.Run("complete valid example", func(t *testing.T) {
		result := db.SearchOrganizations("organization uno")
		assert.Len(t, result, 1)
	})

	t.Run("partial match returns organization", func(t *testing.T) {
		result := db.SearchOrganizations("uno")
		assert.Len(t, result, 1)
	})

	t.Run("wide match returns 2 organization", func(t *testing.T) {
		result := db.SearchOrganizations("organization")
		assert.Len(t, result, 2)
	})

	t.Run("searching for unknown organization returns empty list", func(t *testing.T) {
		result := db.SearchOrganizations("organization tres")
		assert.Len(t, result, 0)
	})
}

func TestMemoryDb_ReverseLookup(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	eventSystem, db := initDb(*repo)
	if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
		return
	}

	t.Run("finds exact match", func(t *testing.T) {
		result, err := db.ReverseLookup("organization uno")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("finds exact match, case insensitive", func(t *testing.T) {
		result, err := db.ReverseLookup("ORGANIZATION UNO")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("does not fn partial match", func(t *testing.T) {
		result, err := db.ReverseLookup("uno")
		assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		assert.Nil(t, result)
	})
}

func TestMemoryDb_OrganizationById(t *testing.T) {
	t.Run("organization is found", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		result, err := db.OrganizationById("o1")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, result)
	})

	t.Run("organization is not found", func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		eventSystem, db := initDb(*repo)
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		_, err = db.OrganizationById("unknown")
		assert.True(t, errors.Is(err, ErrOrganizationNotFound))
	})
}
