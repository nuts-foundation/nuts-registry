package engine

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterVendor(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor("id", "name", "domain").Return(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}), nil)
		command.SetArgs([]string{"register-vendor", "id", "name", "domain"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - no domain (default fallback to 'healthcare')", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor("id", "name", "healthcare").Return(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}), nil)
		command.SetArgs([]string{"register-vendor", "id", "name"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
		command.SetArgs([]string{"register-vendor", "id", "name", "domain"})
		command.Execute()
	}))
}

func TestVendorClaim(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.VendorClaim, domain.RegisterVendorEvent{})
		client.EXPECT().VendorClaim("vendorId", "orgId", "orgName", nil).Return(event, nil)
		command.SetArgs([]string{"vendor-claim", "vendorId", "orgId", "orgName"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().VendorClaim(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
		command.SetArgs([]string{"vendor-claim", "vendorId", "orgId", "orgName"})
		command.Execute()
	}))
}

func TestRegisterEndpoint(t *testing.T) {
	command := cmd()
	t.Run("ok - bare minimum parameters", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{})
		client.EXPECT().RegisterEndpoint("orgId", "", "url", "type", db.StatusActive, map[string]string{}).Return(event, nil)
		command.SetArgs([]string{"register-endpoint", "orgId", "type", "url"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - all parameters", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{})
		client.EXPECT().RegisterEndpoint("orgId", "id", "url", "type", db.StatusActive, map[string]string{"k1": "v1", "k2": "v2"}).Return(event, nil)
		command.SetArgs([]string{"register-endpoint", "orgId", "type", "url", "-i", "id", "-p", "k1=v1", "-p", "k2=v2"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterEndpoint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
		command.SetArgs([]string{"register-endpoint", "orgId", "type", "url"})
		command.Execute()
	}))
}

func TestSearchOrg(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().SearchOrganizations("foo")
		command.SetArgs([]string{"search", "foo"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
}

func TestPrintVersion(t *testing.T) {
	command := cmd()
	command.SetArgs([]string{"version"})
	err := command.Execute()
	assert.NoError(t, err)
}

func withMock(test func(t *testing.T, client *mock.MockRegistryClient)) func(t *testing.T) {
	return func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		registryClient := mock.NewMockRegistryClient(mockCtrl)
		registryClientCreator = func() pkg.RegistryClient {
			return registryClient
		}
		test(t, registryClient)
	}
}

func Test_flagSet(t *testing.T) {
	assert.NotNil(t, flagSet())
}