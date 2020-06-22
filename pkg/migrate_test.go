package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockConfig struct {
	identity string
}

func (m mockConfig) ServerAddress() string {
	panic("implement me")
}

func (m mockConfig) InStrictMode() bool {
	panic("implement me")
}

func (m mockConfig) Mode() string {
	panic("implement me")
}

func (m mockConfig) Identity() string {
	return m.identity
}

func (m mockConfig) GetEngineMode(engineMode string) string {
	panic("implement me")
}

func TestRegistry_verify(t *testing.T) {
	vendorName := "vendorName"
	orgId := "orgId"
	orgName := "orgName"
	test := func(t *testing.T, autoFix bool) {
		t.Run("all is ok", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(vendorName, domain.HealthcareDomain)
			cxt.registry.VendorClaim(orgId, orgName, nil)
			resultingEvents, needsFixing, err := cxt.registry.Verify(autoFix)
			assert.Empty(t, resultingEvents)
			assert.False(t, needsFixing)
			assert.NoError(t, err)
		})
		t.Run("vendor not registered", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.verify(mockConfig{vendorId}, autoFix)
		})
		t.Run("vendor has no certificates, issued", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			err := cxt.registry.EventSystem.PublishEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
				Name:       vendorName,
				Identifier: vendorId,
			}, nil))
			if !assert.NoError(t, err) {
				return
			}
			events, needsFixing, err := cxt.registry.verify(mockConfig{vendorId}, autoFix)
			assert.NoError(t, err)
			if autoFix {
				assert.Len(t, cxt.registry.Db.VendorByID(vendorId).Keys, 1)
				assert.Len(t, events, 1)
				assert.False(t, needsFixing)
			} else {
				assert.True(t, needsFixing)
			}
		})
		t.Run("error - vendor has certificates but no key material", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(vendorName, domain.HealthcareDomain)
			cxt.empty()
			_, _, err := cxt.registry.verify(mockConfig{vendorId}, autoFix)
			assert.Error(t, err)
		})
		t.Run("org has no certificates", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(vendorName, domain.HealthcareDomain)
			cxt.registry.EventSystem.PublishEvent(events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
				VendorIdentifier: vendorId,
				OrgIdentifier:    domain.Identifier(orgId),
				OrgName:          orgName,
			}, nil))
			// Assert that the org has no keys
			org, _ := cxt.registry.Db.OrganizationById(orgId)
			assert.Len(t, org.Keys, 0)
			// Migrate
			events, needsFixing, err := cxt.registry.verify(mockConfig{vendorId}, autoFix)
			assert.NoError(t, err)
			// Assert a new certificate was issued
			org, _ = cxt.registry.Db.OrganizationById(orgId)
			if autoFix {
				assert.Len(t, org.Keys, 1)
				assert.Len(t, events, 1)
				assert.False(t, needsFixing)
			} else {
				assert.True(t, needsFixing)
			}
		})
		t.Run("error - org has certificates but no key material", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(vendorName, domain.HealthcareDomain)
			cxt.registry.VendorClaim(orgId, vendorName, nil)
			// Empty key material directory
			cxt.empty()
			cxt.registry.crypto.GenerateKeyPair(types.KeyForEntity(types.LegalEntity{URI: vendorId}))
			_, _, err := cxt.registry.verify(mockConfig{vendorId}, autoFix)
			assert.Error(t, err)
		})
	}
	t.Run("only verify", func(t *testing.T) {
		test(t, false)
	})
	t.Run("verify and migrate", func(t *testing.T) {
		test(t, true)
	})
}
