package pkg

import (
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

func TestRegistry_verifyAndMigrateRegistry(t *testing.T) {
	vendorId := "vendorId"
	vendorName := "vendorName"
	orgId := "orgId"
	orgName := "orgName"
	t.Run("vendor not registered", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.verifyAndMigrateRegistry(mockConfig{vendorId})
	})
	t.Run("vendor has no certificates", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		err := cxt.registry.EventSystem.PublishEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
			Name:       vendorName,
			Identifier: domain.Identifier(vendorId),
		}))
		if !assert.NoError(t, err) {
			return
		}
		cxt.registry.verifyAndMigrateRegistry(mockConfig{vendorId})
	})
	t.Run("vendor has certificates but no key material", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.RegisterVendor(vendorId, vendorName, domain.HealthcareDomain)
		if !assert.NoError(t, err) {
			return
		}
		cxt.empty()
		cxt.registry.verifyAndMigrateRegistry(mockConfig{vendorId})
	})
	t.Run("org has no certificates", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(vendorId, vendorName, domain.HealthcareDomain)
		cxt.registry.EventSystem.PublishEvent(events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
			VendorIdentifier: domain.Identifier(vendorId),
			OrgIdentifier:    domain.Identifier(orgId),
			OrgName:          orgName,
		}))
		cxt.registry.verifyAndMigrateRegistry(mockConfig{vendorId})
	})
	t.Run("org has certificates but no key material", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(vendorId, vendorName, domain.HealthcareDomain)
		cxt.registry.VendorClaim(vendorId, orgId, vendorName, nil)
		// Empty key material directory
		cxt.empty()
		cxt.registry.verifyAndMigrateRegistry(mockConfig{vendorId})
	})
}
