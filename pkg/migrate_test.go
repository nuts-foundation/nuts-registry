package pkg

import (
	cryptoTypes "github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	test2 "github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockConfig struct {
	identity core.PartyID
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
	return m.identity.String()
}

func (m mockConfig) VendorID() core.PartyID {
	return m.identity
}

func (m mockConfig) GetEngineMode(engineMode string) string {
	panic("implement me")
}

func TestRegistry_verify(t *testing.T) {
	vendorName := "vendorName"
	orgId := test2.OrganizationID("orgId")
	orgName := "orgName"
	test := func(t *testing.T, autoFix bool) {
		t.Run("all is ok", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
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
		t.Run("vendor has no certificates", func(t *testing.T) {
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
			assert.Len(t, cxt.registry.Db.VendorByID(vendorId).Keys, 0)
			assert.Len(t, events, 0)
			assert.False(t, needsFixing)
		})
		t.Run("error - vendor has certificates but no key material", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
			cxt.empty()
			_, _, err := cxt.registry.verify(mockConfig{vendorId}, autoFix)
			assert.Error(t, err)
		})
		t.Run("org has no certificates", func(t *testing.T) {
			cxt := createTestContext(t)
			defer cxt.close()
			cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
			cxt.registry.EventSystem.PublishEvent(events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
				VendorID:       vendorId,
				OrganizationID: orgId,
				OrgName:        orgName,
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
			cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
			cxt.registry.VendorClaim(orgId, vendorName, nil)
			// Empty key material directory
			cxt.empty()
			cxt.registry.crypto.GenerateKeyPair(cryptoTypes.KeyForEntity(cryptoTypes.LegalEntity{URI: vendorId.String()}))
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
