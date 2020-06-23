package engine

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
)

func TestRegisterVendor(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor("name", "domain").Return(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}, nil), nil)
		command.SetArgs([]string{"register-vendor", "name", "domain"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - no domain (default fallback to 'healthcare')", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor("name", "healthcare").Return(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}, nil), nil)
		command.SetArgs([]string{"register-vendor", "name"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RegisterVendor(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
		command.SetArgs([]string{"register-vendor", "name", "domain"})
		command.Execute()
	}))
}

func TestVendorClaim(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.VendorClaim, domain.RegisterVendorEvent{}, nil)
		client.EXPECT().VendorClaim("orgId", "orgName", nil).Return(event, nil)
		command.SetArgs([]string{"vendor-claim", "orgId", "orgName"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().VendorClaim(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
		command.SetArgs([]string{"vendor-claim", "orgId", "orgName"})
		command.Execute()
	}))
}

func TestRefreshVendorCertificate(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{Keys: []interface{}{generateCertificate()}}, nil)
		client.EXPECT().RefreshVendorCertificate().Return(event, nil)
		command.SetArgs([]string{"refresh-vendor-cert"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - no certs", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}, nil)
		client.EXPECT().RefreshVendorCertificate().Return(event, nil)
		command.SetArgs([]string{"refresh-vendor-cert"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RefreshVendorCertificate().Return(nil, errors.New("failed"))
		command.SetArgs([]string{"refresh-vendor-cert"})
		command.Execute()
	}))
}

func TestRefreshOrganizationCertificate(t *testing.T) {
	command := cmd()
	t.Run("ok", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{OrgKeys: []interface{}{generateCertificate()}}, nil)
		client.EXPECT().RefreshOrganizationCertificate("123").Return(event, nil)
		command.SetArgs([]string{"refresh-organization-cert", "123"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - no certs", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil)
		client.EXPECT().RefreshOrganizationCertificate("123").Return(event, nil)
		command.SetArgs([]string{"refresh-organization-cert", "123"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().RefreshOrganizationCertificate("123").Return(nil, errors.New("failed"))
		command.SetArgs([]string{"refresh-organization-cert", "123"})
		command.Execute()
	}))
}

func TestVerify(t *testing.T) {
	t.Run("ok - fix data", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().Verify(true).Return(nil, false, nil)
		command := cmd()
		command.SetArgs([]string{"verify", "-f"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - nothing to do", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().Verify(false).Return(nil, false, nil)
		command := cmd()
		command.SetArgs([]string{"verify"})
		err := command.Execute()
		assert.NoError(t, err)
	}))

	t.Run("ok - data needs fixing", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().Verify(false).Return(nil, true, nil)
		command := cmd()
		command.SetArgs([]string{"verify"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - events emitted", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().Verify(true).Return([]events.Event{events.CreateEvent("foobar", struct{}{}, nil)}, true, nil)
		command := cmd()
		command.SetArgs([]string{"verify", "-f"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("error", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		client.EXPECT().Verify(false).Return(nil, false, errors.New("failed"))
		command := cmd()
		command.SetArgs([]string{"verify"})
		err := command.Execute()
		assert.Error(t, err)
	}))
}

func TestRegisterEndpoint(t *testing.T) {
	command := cmd()
	t.Run("ok - bare minimum parameters", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{}, nil)
		client.EXPECT().RegisterEndpoint("orgId", "", "url", "type", db.StatusActive, map[string]string{}).Return(event, nil)
		command.SetArgs([]string{"register-endpoint", "orgId", "type", "url"})
		err := command.Execute()
		assert.NoError(t, err)
	}))
	t.Run("ok - all parameters", withMock(func(t *testing.T, client *mock.MockRegistryClient) {
		event := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{}, nil)
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

func Test_flagSet(t *testing.T) {
	assert.NotNil(t, flagSet())
}

func TestNewRegistryEngine(t *testing.T) {

	t.Run("instance", func(t *testing.T) {
		assert.NotNil(t, NewRegistryEngine())
	})

	t.Run("configuration", func(t *testing.T) {
		e := NewRegistryEngine()
		cfg := core.NutsConfig()
		cfg.RegisterFlags(e.Cmd, e)
		assert.NoError(t, cfg.InjectIntoEngine(e))
	})
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

func generateCertificate() map[string]interface{} {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	certAsBytes := test.GenerateCertificateEx(time.Now(), 1, privateKey)
	certificate, _ := x509.ParseCertificate(certAsBytes)
	certAsJWK, _ := cert.CertificateToJWK(certificate)
	certAsMap, _ := cert.JwkToMap(certAsJWK)
	return certAsMap
}
