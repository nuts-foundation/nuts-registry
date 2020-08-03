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
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	test2 "github.com/nuts-foundation/nuts-crypto/test"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
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

var registerVendor1 = events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
	Identifier: test.VendorID("v1"),
	Name:       "Vendor Uno",
}, nil)
var registerVendor2 = events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
	Identifier: test.VendorID("v2"),
	Name:       "Vendor Dos",
}, nil)

var vendorClaim1 = events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
	VendorID:       test.VendorID("v1"),
	OrganizationID: test.OrganizationID("o1"),
	OrgName:        "Organization Uno",
	OrgKeys:        nil,
}, nil)

var vendorClaim2 = events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
	VendorID:       test.VendorID("v1"),
	OrganizationID: test.OrganizationID("o2"),
	OrgName:        "Organization Dos",
	OrgKeys:        nil,
}, nil)

var registerEndpoint1 = events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{
	Organization: test.OrganizationID("o1"),
	URL:          "foo:bar",
	EndpointType: "simple",
	Identifier:   "e1",
	Status:       StatusActive,
}, nil)
var registerEndpoint2 = events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{
	Organization: test.OrganizationID("o1"),
	URL:          "foo:bar",
	EndpointType: "simple",
	Identifier:   "e2",
	Status:       "inactive",
}, nil)

func TestNew(t *testing.T) {
	emptyDb := New()

	if len(emptyDb.vendors) != 0 {
		t.Errorf("Expected 0 len structure, got [%d]", len(emptyDb.vendors))
	}
}

func initDb(repo test.TestRepo) (events.EventSystem, *MemoryDb) {
	db := New()
	eventSystem := events.NewEventSystem(domain.GetEventTypes()...)
	eventSystem.Configure(repo.Directory + "/events")
	db.RegisterEventHandlers(eventSystem.RegisterEventHandler)
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
	t.Run("ok", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
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

	t.Run("ok - update", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		payload := domain.RegisterVendorEvent{}
		registerVendor1.Unmarshal(&payload)
		payload.Name = "Foobar"
		err := eventSystem.PublishEvent(events.CreateEvent(registerVendor1.Type(), payload, registerVendor1.Ref()))
		assert.NoError(t, err)
		assert.Equal(t, payload.Name, db.vendors[payload.Identifier.String()].Name)
	}))

	t.Run("errors - update - different vendorID", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		payload := domain.RegisterVendorEvent{}
		registerVendor1.Unmarshal(&payload)
		payload.Name = "Foobar"
		payload.Identifier = test.VendorID("123")
		err := eventSystem.PublishEvent(events.CreateEvent(registerVendor1.Type(), payload, registerVendor1.Ref()))
		assert.EqualError(t, err, "referred event contains a different vendor: actual vendorId (urn:oid:1.3.6.1.4.1.54851.4:v1) differs from expected (urn:oid:1.3.6.1.4.1.54851.4:123)")
	}))

	t.Run("error - vendor already exists", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		err = eventSystem.PublishEvent(registerVendor1)
		assert.EqualError(t, err, "vendor already registered (id = urn:oid:1.3.6.1.4.1.54851.4:v1)")
	}))
}

func TestMemoryDb_VendorClaim(t *testing.T) {
	t.Run("ok", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerVendor1)
		if !assert.NoError(t, err) {
			return
		}
		vendorID := test.VendorID("v1").String()
		assert.Len(t, db.vendors[vendorID].orgs, 0)
		err = eventSystem.PublishEvent(vendorClaim1)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors[vendorID].orgs, 1)
		err = eventSystem.PublishEvent(vendorClaim2)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, db.vendors[vendorID].orgs, 2)
	}))

	t.Run("ok - update", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		payload := domain.VendorClaimEvent{}
		vendorClaim1.Unmarshal(&payload)
		payload.OrgName = "Foobar"
		err := eventSystem.PublishEvent(events.CreateEvent(vendorClaim1.Type(), payload, vendorClaim1.Ref()))
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, payload.OrgName, db.lookupOrg(payload.OrganizationID).OrgName)
	}))

	t.Run("error - update - org ID differs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		payload := domain.VendorClaimEvent{}
		vendorClaim1.Unmarshal(&payload)
		payload.OrganizationID = test.OrganizationID("1234")
		payload.OrgName = "Foobar"
		err := eventSystem.PublishEvent(events.CreateEvent(vendorClaim1.Type(), payload, vendorClaim1.Ref()))
		assert.EqualError(t, err, "can't change organization ID: actual organizationId (urn:oid:2.16.840.1.113883.2.4.6.1:o1) differs from expected (urn:oid:2.16.840.1.113883.2.4.6.1:1234)")
	}))
	t.Run("error - update - vendor ID differs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(registerVendor2)
		eventSystem.PublishEvent(vendorClaim1)
		payload := domain.VendorClaimEvent{}
		vendorClaim1.Unmarshal(&payload)
		payload.VendorID = test.VendorID("v2")
		payload.OrgName = "Foobar"
		err := eventSystem.PublishEvent(events.CreateEvent(vendorClaim1.Type(), payload, vendorClaim1.Ref()))
		assert.EqualError(t, err, "can't change organization's vendor: actual vendorId (urn:oid:1.3.6.1.4.1.54851.4:v1) differs from expected (urn:oid:1.3.6.1.4.1.54851.4:v2)")
	}))

	t.Run("error - update - org ID differs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		payload := domain.VendorClaimEvent{}
		vendorClaim1.Unmarshal(&payload)
		payload.OrganizationID = test.OrganizationID("1234")
		payload.OrgName = "Foobar"
		err := eventSystem.PublishEvent(events.CreateEvent(vendorClaim1.Type(), payload, vendorClaim1.Ref()))
		assert.EqualError(t, err, "can't change organization ID: actual organizationId (urn:oid:2.16.840.1.113883.2.4.6.1:o1) differs from expected (urn:oid:2.16.840.1.113883.2.4.6.1:1234)")
	}))

	t.Run("error - unknown vendor", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(vendorClaim1)
		assert.Error(t, err)
	}))

	t.Run("error - duplicate organization", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
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
		assert.NotNil(t, db.VendorByID(test.VendorID("v1")))
	})
	t.Run("not found", func(t *testing.T) {
		assert.Nil(t, db.VendorByID(test.VendorID("v2")))
	})
}

func TestMemoryDb_FindEndpointsByOrganization(t *testing.T) {
	t.Run("Valid example", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
		assert.NotEmpty(t, result[0].Organization)
	}))

	t.Run("Valid example with type", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1) {
			return
		}

		et := "simple"
		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), &et)
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
		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), &et)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 0)
	}))

	t.Run("Inactive mappings are not returned", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, registerEndpoint1, registerEndpoint2) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, result, 1)
	}))

	t.Run("no mapping", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), nil)
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

		result, err := db.FindEndpointsByOrganizationAndType(test.OrganizationID("o1"), nil)
		if !assert.Error(t, err) {
			return
		}
		assert.Len(t, result, 0)
	}))
}

func TestMemoryDb_RegisterEndpoint(t *testing.T) {
	t.Run("ok", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		err := eventSystem.PublishEvent(registerEndpoint1)
		assert.NoError(t, err)
	}))
	t.Run("ok - update", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		payload1 := domain.RegisterEndpointEvent{}
		registerEndpoint1.Unmarshal(&payload1)
		eventSystem.PublishEvent(registerEndpoint1)
		// Create updated event
		payload2 := domain.RegisterEndpointEvent{}
		registerEndpoint1.Unmarshal(&payload2)
		payload2.Properties = map[string]string{"hello": "world"}
		payload2.EndpointType += "-updated"
		payload2.URL += "-updated"
		err := eventSystem.PublishEvent(events.CreateEvent(registerEndpoint1.Type(), payload2, registerEndpoint1.Ref()))
		if !assert.NoError(t, err) {
			return
		}
		endpoints, _ := db.FindEndpointsByOrganizationAndType(payload1.Organization, nil)
		if !assert.Len(t, endpoints, 1) {
			return
		}
		assert.Equal(t, payload2.EndpointType, endpoints[0].EndpointType)
		assert.Equal(t, payload2.URL, endpoints[0].URL)
		assert.Equal(t, payload2.Properties, endpoints[0].Properties)
	}))
	t.Run("error - can't change org for endpoint", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		eventSystem.PublishEvent(registerEndpoint1)
		eventSystem.PublishEvent(vendorClaim2)
		payload := domain.RegisterEndpointEvent{}
		registerEndpoint1.Unmarshal(&payload)
		payload.Organization = test.OrganizationID("o2") // this is org from vendorClaim2
		err := eventSystem.PublishEvent(events.CreateEvent(registerEndpoint1.Type(), payload, registerEndpoint1.Ref()))
		assert.EqualError(t, err, "can't change endpoint's organization: actual organizationId (urn:oid:2.16.840.1.113883.2.4.6.1:o1) differs from expected (urn:oid:2.16.840.1.113883.2.4.6.1:o2)")
	}))
	t.Run("error - endpoint already registered", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		eventSystem.PublishEvent(registerVendor1)
		eventSystem.PublishEvent(vendorClaim1)
		eventSystem.PublishEvent(registerEndpoint1)
		err := eventSystem.PublishEvent(registerEndpoint1)
		assert.EqualError(t, err, "endpoint already registered for this organization (id = e1)")
	}))
	t.Run("error - unknown organization", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		err := eventSystem.PublishEvent(registerEndpoint1)
		assert.EqualError(t, err, "organization not registered (id = urn:oid:2.16.840.1.113883.2.4.6.1:o1)")
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

		result, err := db.OrganizationById(test.OrganizationID("o1"))
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, result)
	}))

	t.Run("organization is not found", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1) {
			return
		}

		_, err := db.OrganizationById(test.OrganizationID("unknown"))
		assert.True(t, errors.Is(err, ErrOrganizationNotFound))
	}))
}

func TestMemoryDb_OrganizationsByVendorID(t *testing.T) {
	t.Run("vendor with 2 orgs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1, vendorClaim1, vendorClaim2) {
			return
		}
		orgs := db.OrganizationsByVendorID(test.VendorID("v1"))
		assert.Len(t, orgs, 2)
	}))
	t.Run("vendor not found", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem) {
			return
		}
		orgs := db.OrganizationsByVendorID(test.VendorID("unknown vendor"))
		assert.Len(t, orgs, 0)
	}))
	t.Run("vendor with no orgs", withTestContext(func(t *testing.T, eventSystem events.EventSystem, db *MemoryDb) {
		if !pub(t, eventSystem, registerVendor1) {
			return
		}
		orgs := db.OrganizationsByVendorID(test.VendorID("v1"))
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

func Test_org_toDb(t *testing.T) {
	t.Run("PublicKey backwards compatibility", func(t *testing.T) {
		rsaKey := test2.GenerateRSAKey()
		keyAsJWK, _ := jwk.New(&rsaKey.PublicKey)
		jwkAsMap, _ := cert.JwkToMap(keyAsJWK)
		jwkAsMap["kty"] = "RSA"
		o := org{VendorClaimEvent: domain.VendorClaimEvent{
			VendorID:       test.VendorID("v1"),
			OrganizationID: test.OrganizationID("o1"),
			OrgName:        "Organization Uno",
			OrgKeys:        []interface{}{jwkAsMap},
		}}
		publicKey := o.toDb().PublicKey
		if !assert.NotNil(t, publicKey) {
			return
		}
		pubKey, _ := cert.PemToPublicKey([]byte(*publicKey))
		assert.Equal(t, rsaKey.PublicKey, *pubKey)
	})
}
