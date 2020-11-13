/*
 * Nuts registry
 * Copyright (C) 2020. Nuts community
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
package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cryptoTypes "github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	certutil "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/pkg/network"
	"github.com/nuts-foundation/nuts-registry/test"
)

var vendorId core.PartyID

const vendorName = "Test Vendor"

func init() {
	vendorId = test.VendorID("4")
}

func TestRegistryAdministration_RegisterEndpoint(t *testing.T) {
	var payload = domain.RegisterEndpointEvent{}
	orgID := test.OrganizationID("orgId")
	t.Run("ok", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterEndpoint, func(e events.Event, _ events.EventLookup) error {
			return e.Unmarshal(&payload)
		})
		_, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		if !assert.NoError(t, err) {
			return
		}
		_, err = cxt.registry.VendorClaim(orgID, "org", nil)
		if !assert.NoError(t, err) {
			return
		}
		event, err := cxt.registry.RegisterEndpoint(orgID, "endpointId", "url", "type", "status", map[string]string{"foo": "bar"})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event.Signature())
		assert.Equal(t, orgID, payload.Organization)
		assert.Equal(t, "endpointId", string(payload.Identifier))
		assert.Equal(t, "url", payload.URL)
		assert.Equal(t, "type", payload.EndpointType)
		assert.Equal(t, "status", payload.Status)
		assert.Len(t, payload.Properties, 1)
	})
	t.Run("ok - update", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterEndpoint, func(e events.Event, _ events.EventLookup) error {
			return e.Unmarshal(&payload)
		})
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		cxt.registry.VendorClaim(orgID, "org", nil)
		cxt.registry.RegisterEndpoint(orgID, "endpointId", "url", "type", "status", map[string]string{"foo": "bar"})
		// Now update endpoint
		event, err := cxt.registry.RegisterEndpoint(orgID, "endpointId", "url-updated", "type-updated", "status-updated", map[string]string{"foo": "bar-updated"})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
		assert.False(t, event.PreviousRef().IsZero())
		assert.Equal(t, orgID, payload.Organization)
		assert.Equal(t, "endpointId", string(payload.Identifier))
		assert.Equal(t, "url-updated", payload.URL)
		assert.Equal(t, "type-updated", payload.EndpointType)
		assert.Equal(t, "status-updated", payload.Status)
		assert.Len(t, payload.Properties, 1)
	})
	t.Run("ok - auto generate id", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterEndpoint, func(e events.Event, _ events.EventLookup) error {
			return e.Unmarshal(&payload)
		})
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		cxt.registry.VendorClaim(orgID, "org", nil)
		event, err := cxt.registry.RegisterEndpoint(orgID, "", "url", "type", "status", map[string]string{"foo": "bar"})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event.Signature())
		assert.Len(t, payload.Identifier, 36) // 36 = length of UUIDv4 as string
	})
	t.Run("ok - org has no certificates", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterEndpoint, func(e events.Event, _ events.EventLookup) error {
			return e.Unmarshal(&payload)
		})
		cxt.registry.EventSystem.ProcessEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
			Identifier: vendorId,
			Name:       vendorName,
		}, nil))
		cxt.registry.VendorClaim(orgID, "org", nil)
		event, err := cxt.registry.RegisterEndpoint(orgID, "", "url", "type", "status", map[string]string{"foo": "bar"})
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, event.Signature())
	})
	t.Run("error - org not found", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		endpoint, err := cxt.registry.RegisterEndpoint(orgID, "", "url", "type", "status", map[string]string{"foo": "bar"})
		assert.Nil(t, endpoint)
		assert.Error(t, err)
	})
}

func TestRegistryAdministration_VendorClaim(t *testing.T) {
	var payload = domain.VendorClaimEvent{}
	registerEventHandler := func(registry *Registry) {
		registry.EventSystem.RegisterEventHandler(domain.VendorClaim, func(e events.Event, _ events.EventLookup) error {
			return e.Unmarshal(&payload)
		})
	}
	t.Run("ok - keys generated by crypto", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		if err != nil {
			panic(err)
		}
		registerEventHandler(cxt.registry)

		event, err := cxt.registry.VendorClaim(test.OrganizationID(t.Name()), "orgName", nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event.Signature())
		assert.Equal(t, vendorId, payload.VendorID)
		assert.Equal(t, test.OrganizationID(t.Name()), payload.OrganizationID)
		assert.Equal(t, "orgName", payload.OrgName)
		if !assert.Len(t, payload.OrgKeys, 1) {
			return
		}
		keyAsJwk, err := cert.MapToJwk(payload.OrgKeys[0].(map[string]interface{}))
		if !assert.NoError(t, err) {
			return
		}
		assert.False(t, payload.Start.IsZero())
		assert.Nil(t, payload.End)
		// Check certificate
		chainInterf, _ := keyAsJwk.Get("x5c")
		chain := chainInterf.(jwk.CertificateChain).Get()
		if !assert.Len(t, chain, 1) {
			return
		}
		assert.Equal(t, "CN=orgName,O=Test Vendor,C=NL", chain[0].Subject.String())
		assert.Equal(t, "CN=Test Vendor CA,O=Test Vendor,C=NL", chain[0].Issuer.String())
		assert.Equal(t, x509.KeyUsageDigitalSignature, chain[0].KeyUsage&x509.KeyUsageDigitalSignature)
		assert.Equal(t, x509.KeyUsageKeyEncipherment, chain[0].KeyUsage&x509.KeyUsageKeyEncipherment)
		assert.Equal(t, x509.KeyUsageDataEncipherment, chain[0].KeyUsage&x509.KeyUsageDataEncipherment)
	})
	t.Run("ok - existing org keys in crypto", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		vendor, _ := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		registerVendorEvent := domain.RegisterVendorEvent{}
		vendor.Unmarshal(&registerVendorEvent)
		vendorCertChain, _ := cert.MapToX509CertChain(registerVendorEvent.Keys[0].(map[string]interface{}))
		vendorPrivKey, _ := cxt.registry.crypto.GetPrivateKey(cryptoTypes.KeyForEntity(cryptoTypes.LegalEntity{URI: vendorId.String()}))

		orgName := "Test Organization"
		org := test.OrganizationID(uuid.New().String())
		orgKey := cryptoTypes.KeyForEntity(cryptoTypes.LegalEntity{URI: org.String()})
		// "Out-of-Band" generate key material
		orgPubKey, err := cxt.registry.crypto.GenerateKeyPair(orgKey, false)
		if !assert.NoError(t, err) {
			return
		}
		// Now "Out-of-Band"-sign a certificate with it using the vendor priv. key
		csr, _ := certutil.OrganisationCertificateRequest(vendorName, org, orgName, "healthcare")
		csr.PublicKey = orgPubKey
		orgCertificate := test.SignCertificateFromCSRWithKey(csr, time.Now(), 2, vendorCertChain[0], vendorPrivKey)
		// Feed it to VendorClaim()
		orgCertAsJWK, _ := cert.CertificateToJWK(orgCertificate)
		jwkAsMap, _ := cert.JwkToMap(orgCertAsJWK)
		//jwkAsMap[jwk.X509CertChainKey] = base64.StdEncoding.EncodeToString(orgCertificate.Raw)
		event, err := cxt.registry.VendorClaim(org, orgName, []interface{}{jwkAsMap})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event.Signature())
	})

	t.Run("ok - vendor has no active certificates", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.ProcessEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
			Identifier: vendorId,
			Name:       vendorName,
		}, nil))
		org := test.OrganizationID(t.Name())
		event, err := cxt.registry.VendorClaim(org, "orgName", nil)
		assert.NoError(t, err)
		assert.NoError(t, event.Unmarshal(&payload))
		assert.Len(t, payload.OrgKeys, 1)
		// No certificate means no signature
		assert.Nil(t, event.Signature())
		// A certificate couldn't have been issued
		certChain, err := cert.MapToX509CertChain(payload.OrgKeys[0].(map[string]interface{}))
		assert.NoError(t, err)
		assert.Nil(t, certChain)
	})

	t.Run("error - vendor not found", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.VendorClaim(test.OrganizationID(t.Name()), "orgName", nil)
		assert.Contains(t, err.Error(), "vendor not found")
	})

	t.Run("error - vendor has no keys", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		cxt.empty()
		_, err := cxt.registry.VendorClaim(test.OrganizationID("org"), "orgName", nil)
		assert.Contains(t, err.Error(), crypto.ErrUnknownCA.Error())
		assert.Contains(t, err.Error(), ErrCertificateIssue.Error())
	})

	t.Run("error - while generating key", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		c := cxt.registry.crypto.(*crypto.Crypto)
		var defaultKeySize = c.Config.Keysize
		c.Config.Keysize = -1
		defer func() {
			c.Config.Keysize = defaultKeySize
		}()
		_, err := cxt.registry.VendorClaim(test.OrganizationID("org"), "orgName", nil)
		assert.Error(t, err)
	})

	t.Run("error - unable to load existing key", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		org := test.OrganizationID("org")
		entity := cryptoTypes.LegalEntity{URI: org.String()}
		_, err := cxt.registry.crypto.GenerateKeyPair(cryptoTypes.KeyForEntity(entity), false)
		if !assert.NoError(t, err) {
			return
		}
		f := getLastUpdatedFile(filepath.Join(cxt.repo.Directory, "crypto"))
		ioutil.WriteFile(f, []byte("this is not a private key"), os.ModePerm)
		_, err = cxt.registry.VendorClaim(org, "orgName", nil)
		assert.EqualError(t, err, "malformed PEM block")
	})
}

func TestRegistryAdministration_RegisterVendor(t *testing.T) {
	t.Run("ok - register", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		var registerVendorEvent *domain.RegisterVendorEvent
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterVendor, func(event events.Event, _ events.EventLookup) error {
			e := domain.RegisterVendorEvent{}
			if err := event.Unmarshal(&e); err != nil {
				return err
			}
			registerVendorEvent = &e
			return nil
		})

		event, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		if !assert.NoError(t, err) {
			return
		}
		// Verify RegisterVendor event emitted
		if !assert.NotNil(t, registerVendorEvent) {
			return
		}
		assert.NotNil(t, event)
		// Verify issued signing certificate
		key, err := cert.MapToJwk(registerVendorEvent.Keys[0].(map[string]interface{}))
		if err != nil {
			panic(err)
		}
		chain := key.X509CertChain()
		assert.Len(t, chain, 1)
		assert.Equal(t, "CN=Test Vendor CA,O=Test Vendor,C=NL", chain[0].Subject.String())
	})
	t.Run("ok - update", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())

		var registerVendorEvent *domain.RegisterVendorEvent
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterVendor, func(event events.Event, _ events.EventLookup) error {
			e := domain.RegisterVendorEvent{}
			if err := event.Unmarshal(&e); err != nil {
				return err
			}
			registerVendorEvent = &e
			return nil
		})
		updateEvent, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, registerVendorEvent) {
			return
		}
		assert.False(t, updateEvent.PreviousRef().IsZero())
	})
	t.Run("error - unable to publish event", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.EventSystem.RegisterEventHandler(domain.RegisterVendor, func(event events.Event, _ events.EventLookup) error {
			return errors.New("unit test error")
		})
		_, err := cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		assert.Contains(t, err.Error(), "unit test error")
	})
}

func TestRegistryAdministration_RefreshOrganizationCertificate(t *testing.T) {
	var org = test.OrganizationID("123")
	var orgEntity = cryptoTypes.LegalEntity{URI: org.String()}
	t.Run("ok", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		cxt.registry.VendorClaim(org, "Test Org", nil)
		publicKeyBeforeRefresh, _ := cxt.registry.crypto.GetPublicKeyAsPEM(cryptoTypes.KeyForEntity(orgEntity))
		event, err := cxt.registry.RefreshOrganizationCertificate(org)
		if !assert.NoError(t, err) {
			return
		}
		publicKeyAfterRefresh, _ := cxt.registry.crypto.GetPublicKeyAsPEM(cryptoTypes.KeyForEntity(orgEntity))
		assert.NotNil(t, event.Signature())
		org, _ := cxt.registry.Db.OrganizationById(org)
		assert.Len(t, org.Keys, 2)
		assert.NotNil(t, publicKeyBeforeRefresh)
		assert.Equal(t, publicKeyBeforeRefresh, publicKeyAfterRefresh, "refresh certificate should not generate a new key pair")
	})
	t.Run("error - vendor not found", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		event, err := cxt.registry.RefreshOrganizationCertificate(org)
		assert.Nil(t, event)
		assert.EqualError(t, err, "vendor not found (id=urn:oid:1.3.6.1.4.1.54851.4:4)")
	})
	t.Run("error - organization not found", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.RegisterVendor(cxt.issueVendorCACertificate())
		event, err := cxt.registry.RefreshOrganizationCertificate(org)
		assert.Nil(t, event)
		assert.EqualError(t, err, "organization not found")
	})
}

func TestCreateAndSubmitCSR(t *testing.T) {
	entity := cryptoTypes.LegalEntity{URI: "foo"}

	t.Run("ok", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		cxt.registry.crypto.GenerateKeyPair(cryptoTypes.KeyForEntity(entity), false)
		_, err := cxt.registry.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{Subject: pkix.Name{CommonName: "Mosselman"}}, nil
		}, entity, entity, crypto.CertificateProfile{})
		assert.NoError(t, err)
	})
	t.Run("error - csr template fn error", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{}, errors.New("oops")
		}, entity, entity, crypto.CertificateProfile{})
		assert.Contains(t, err.Error(), "unable to create CSR template")
	})
	t.Run("error - key pair unavailable", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		unknownEntity := cryptoTypes.LegalEntity{URI: "unknown"}
		_, err := cxt.registry.createAndSubmitCSR(func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{Subject: pkix.Name{CommonName: "Mosselman"}}, nil
		}, unknownEntity, unknownEntity, crypto.CertificateProfile{})
		assert.Contains(t, err.Error(), "unable to retrieve subject private key")
	})
}

func TestRegistry_signAsOrganization(t *testing.T) {
	t.Run("error - unable to create CSR", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		err := cxt.registry.EventSystem.ProcessEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{Identifier: vendorId, Name: "Vendor"}, nil))
		if !assert.NoError(t, err) {
			return
		}
		_, err = cxt.registry.signAsOrganization(test.OrganizationID("orgId"), "", []byte{1, 2, 3}, time.Now(), true)
		assert.Equal(t, "unable to create CSR for JWS signing: missing organization name", err.Error())
	})
	t.Run("error - unable to sign JWS (CA key material missing)", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		err := cxt.registry.EventSystem.ProcessEvent(events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{Identifier: vendorId, Name: "Vendor"}, nil))
		if !assert.NoError(t, err) {
			return
		}
		_, err = cxt.registry.signAsOrganization(test.OrganizationID("orgId"), "Foobar", nil, time.Now(), true)
		assert.Contains(t, err.Error(), "unable to sign JWS: unknown CA")
	})
	t.Run("error - vendor not found", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		_, err := cxt.registry.signAsOrganization(test.OrganizationID("orgId"), "Foobar", nil, time.Now(), true)
		assert.Contains(t, err.Error(), "vendor not found (id=urn:oid:1.3.6.1.4.1.54851.4:4)")
	})
}

func TestRegistry_signAndPublishEvent(t *testing.T) {
	t.Run("error - signer returns error", func(t *testing.T) {
		cxt := createTestContext(t)
		defer cxt.close()
		event, err := cxt.registry.signAndPublishEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}, nil, func([]byte, time.Time) ([]byte, error) {
			return nil, errors.New("error")
		})
		assert.Nil(t, event)
		assert.Error(t, err, "error")
	})
}

func getLastUpdatedFile(dir string) string {
	entries, _ := ioutil.ReadDir(dir)
	if len(entries) == 0 {
		panic("no entries in dir: " + dir)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ModTime().After(entries[j].ModTime())
	})
	return filepath.Join(dir, entries[0].Name())
}

func createRegistry(repo *test.TestRepo) *Registry {
	core.NutsConfig().Load(&cobra.Command{})
	return NewTestRegistryInstance(repo.Directory)
}

type testContext struct {
	identity          core.PartyID
	vendorName        string
	registry          *Registry
	networkAmbassador *network.MockAmbassador
	mockCtrl          *gomock.Controller
	repo              *test.TestRepo
}

var nutsCACertificate *x509.Certificate
var nutsCAKey *rsa.PrivateKey

func init() {
	nutsCAKey, _ = rsa.GenerateKey(rand.Reader, 1024)
	nutsCACertificate = test.SignCertificateFromCSRWithKey(x509.CertificateRequest{
		PublicKey: nutsCAKey.Public(),
		Subject: pkix.Name{
			CommonName: "Root CA",
		},
	}, time.Now(), 365*10, nil, nutsCAKey)
}

func (cxt *testContext) empty() {
	cxt.repo.Cleanup()
	os.MkdirAll(filepath.Join(cxt.repo.Directory, "crypto"), os.ModePerm)
	os.MkdirAll(filepath.Join(cxt.repo.Directory, "events"), os.ModePerm)
}

func (cxt *testContext) close() {
	defer cxt.mockCtrl.Finish()
	defer os.Unsetenv("NUTS_IDENTITY")
	defer cxt.registry.Shutdown()
	defer cxt.repo.Cleanup()
}

func (cxt *testContext) issueVendorCACertificate() *x509.Certificate {
	// "Import" Root CA
	caEntity := cryptoTypes.LegalEntity{URI: "rootca"}
	cryptoStorage := cxt.registry.crypto.(*crypto.Crypto).Storage
	if err := cryptoStorage.SaveCertificate(cryptoTypes.KeyForEntity(caEntity), nutsCACertificate.Raw); err != nil {
		panic(err)
	}
	if err := cryptoStorage.SavePrivateKey(cryptoTypes.KeyForEntity(caEntity), nutsCAKey); err != nil {
		panic(err)
	}
	cxt.registry.crypto.TrustStore().AddCertificate(nutsCACertificate)

	csr, _ := cxt.registry.crypto.GenerateVendorCACSR(vendorName)
	vendorCACertificate, err := cxt.registry.crypto.SignCertificate(cryptoTypes.KeyForEntity(cryptoTypes.LegalEntity{vendorId.String()}), cryptoTypes.KeyForEntity(caEntity), csr, crypto.CertificateProfile{
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:         true,
		NumDaysValid: 365 * 3,
	})
	if err != nil {
		panic(err)
	}
	certificate, _ := x509.ParseCertificate(vendorCACertificate)
	return certificate
}

func createTestContext(t *testing.T) testContext {
	os.Setenv("NUTS_IDENTITY", vendorId.String())
	repo, err := test.NewTestRepo(t)
	if err != nil {
		panic(err)
	}
	mockCtrl := gomock.NewController(t)
	context := testContext{
		identity:          vendorId,
		vendorName:        vendorName,
		registry:          createRegistry(repo),
		mockCtrl:          mockCtrl,
		networkAmbassador: network.NewMockAmbassador(mockCtrl),
		repo:              repo,
	}
	return context
}
