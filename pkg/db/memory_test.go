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
})
var registerEndpoint2 = events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
	Organization: "o1",
	URL:          "foo:bar",
	EndpointType: "simple",
	Identifier:   "e2",
	Status:       "inactive",
})

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
	t.Run("valid example", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors, 1)
		err = eventSystem.PublishEvent(registerVendor2)
		if assert.NoError(t, err) {
			assert.Len(t, db.vendors, 2)
		}
	}))

	t.Run("duplicate entry", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(registerVendor1)
		assert.Error(t, err)
	}))

	t.Run("vendor with invalid key set", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		e := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
			Identifier: "v2",
			Name:       "Foobar",
			Keys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		err := eventSystem.PublishEvent(e)
		assert.Error(t, err)
		assert.Nil(t, db.vendors["v2"])
	}))
}

func TestMemoryDb_VendorClaim(t *testing.T) {
	t.Run("valid example", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
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
	}))

	t.Run("organization with invalid key set", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
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
	}))

	t.Run("unknown vendor", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(vendorClaim1)
		assert.Error(t, err)
	}))

	t.Run("duplicate organization", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(vendorClaim1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(vendorClaim1)
		assert.Error(t, err)
	}))
}

func TestMemoryDb_VendorByID(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	eventSystem, db := initDb(*repo)
	if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
		return
	}

	t.Run("found", func(t *testing.T) {
		assert.NotNil(t, db.VendorByID("v1"))
	})
	t.Run("not found", func(t *testing.T) {
		assert.Nil(t, db.VendorByID("v2"))
	})
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {
	t.Run("Valid example", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	}))

	t.Run("Valid example with type", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		et := "simple"
		result, err := db.FindEndpointsByOrganizationAndType("o1", &et)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	}))

	t.Run("incorrect type", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		et := "unknown"
		result, err := db.FindEndpointsByOrganizationAndType("o1", &et)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 0)
	}))

	t.Run("Inactive mappings are not returned", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1, registerEndpoint2) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	}))

	t.Run("no mapping", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 0)
	}))

	t.Run("Organization unknown", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType("o1", nil)
		if !assert.Error(t, err) {
			return
		}
		assert.Len(t, result, 0)
	}))
}

func TestMemoryDb_RegisterEndpoint(t *testing.T) {
	t.Run("unknown organization", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerEndpoint1)
		assert.Error(t, err)
	}))
}

func TestMemoryDb_SearchOrganizations(t *testing.T) {
	t.Run("tests", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
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
	}))
}

func TestMemoryDb_ReverseLookup(t *testing.T) {
	t.Run("tests", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
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
	}))
}

func TestMemoryDb_OrganizationById(t *testing.T) {
	t.Run("organization is found", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		result, err := db.OrganizationById("o1")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, result)
	}))

	t.Run("organization is not found", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		_, err := db.OrganizationById("unknown")
		assert.True(t, errors.Is(err, ErrOrganizationNotFound))
	}))
}

func TestMemoryDb_OrganizationsByVendorID(t *testing.T) {
	t.Run("vendor with 2 orgs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, vendorClaim2) {
			return
		}
		orgs := db.OrganizationsByVendorID("v1")
		assert.Len(t, orgs, 2)
	}))
	t.Run("vendor not found", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem) {
			return
		}
		orgs := db.OrganizationsByVendorID("unknown vendor")
		assert.Len(t, orgs, 0)
	}))
	t.Run("vendor with no orgs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1) {
			return
		}
		orgs := db.OrganizationsByVendorID("v1")
		assert.Len(t, orgs, 0)
	}))
}

func withTestContext(fn func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb)) func(*testing.T) {
	return func(t *testing.T) {
		repo, err := test.NewTestRepo(t.Name())
		if !assert.NoError(t, err) {
			return
		}
		defer repo.Cleanup()
		eventSystem, db := initDb(*repo)
		fn(t, eventSystem, db)
	}
}
