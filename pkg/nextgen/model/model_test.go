package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"github.com/nuts-foundation/nuts-crypto/test"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

var vendorId, _ = core.NewPartyID("1.2.3.4", "vendorId")
var organizationId, _ = core.NewPartyID("4.3.1.1", "organizationId")
var endpointType, _ = NewOIDURNValue("1.2.3.4", "http")
var location, _ = url.Parse("https://nuts.nl")
var notBefore = time.Date(2020, 10, 20, 20, 30, 0, 0, time.UTC)

func TestMarshalVendor(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	certAsBytes := test.GenerateCertificate(time.Now(), 1, privateKey)
	certificate, _ := x509.ParseCertificate(certAsBytes)

	expected := Vendor{
		Certificates: []*cert.NutsCertificate{cert.NewNutsCertificate(certificate)},
	}
	jsonAsBytes, err := json.Marshal(expected)
	if !assert.NoError(t, err) {
		return
	}
	t.Run("unmarshal", func(t *testing.T) {
		actual := Vendor{}
		err = json.Unmarshal(jsonAsBytes, &actual)
		if !assert.NoError(t, err) {
			return
		}
		// Assert struct eq
		assert.Equal(t, expected.Certificates[0].Raw, actual.Certificates[0].Raw)
	})
}

func TestMarshalEndpoint(t *testing.T) {
	expected := Endpoint{
		ID:        "7FAAFA47-F319-4C6C-A7F5-17E4026C5570",
		VendorID:  vendorId,
		NotBefore: notBefore,
		Location:  NewURL(*location),
		Type:      endpointType,
	}
	jsonAsBytes, err := json.Marshal(expected)
	if !assert.NoError(t, err) {
		return
	}
	t.Run("marshal", func(t *testing.T) {
		expectedJSON := `{"id":"7FAAFA47-F319-4C6C-A7F5-17E4026C5570","vid":"urn:oid:1.2.3.4:vendorId","nbf":"2020-10-20T20:30:00Z","loc":"https://nuts.nl","type":"urn:oid:1.2.3.4:http"}`
		assert.JSONEq(t, expectedJSON, string(jsonAsBytes))
	})
	t.Run("unmarshal", func(t *testing.T) {
		actual := Endpoint{}
		err = json.Unmarshal(jsonAsBytes, &actual)
		if !assert.NoError(t, err) {
			return
		}
		// Assert struct eq
		assert.Equal(t, expected, actual)
	})
}

func TestMarshalOrganization(t *testing.T) {
	expected := Organization{
		ID:       organizationId,
		Name:     "Care Org.",
		VendorID: vendorId,
	}
	jsonAsBytes, err := json.Marshal(expected)
	if !assert.NoError(t, err) {
		return
	}
	t.Run("marshal", func(t *testing.T) {
		expectedJSON := `{"id":"urn:oid:4.3.1.1:organizationId","name":"Care Org.","vid":"urn:oid:1.2.3.4:vendorId"}`
		assert.JSONEq(t, expectedJSON, string(jsonAsBytes))
	})
	t.Run("unmarshal", func(t *testing.T) {
		actual := Organization{}
		err = json.Unmarshal(jsonAsBytes, &actual)
		if !assert.NoError(t, err) {
			return
		}
		// Assert struct eq
		assert.Equal(t, expected, actual)
	})
}

func TestMarshalService(t *testing.T) {
	expected := Service{
		VendorID:       vendorId,
		OrganizationID: organizationId,
		Name:           "Care Org",
		Endpoints:      []types.EndpointID{"1", "2", "3"},
		NotBefore:      notBefore,
	}
	jsonAsBytes, err := json.Marshal(expected)
	if !assert.NoError(t, err) {
		return
	}
	t.Run("marshal", func(t *testing.T) {
		expectedJSON := `{"vid":"urn:oid:1.2.3.4:vendorId","oid":"urn:oid:4.3.1.1:organizationId","name":"Care Org","eps":["1","2","3"],"nbf":"2020-10-20T20:30:00Z"}`
		assert.JSONEq(t, expectedJSON, string(jsonAsBytes))
	})
	t.Run("unmarshal", func(t *testing.T) {
		actual := Service{}
		err = json.Unmarshal(jsonAsBytes, &actual)
		if !assert.NoError(t, err) {
			return
		}
		// Assert struct eq
		assert.Equal(t, expected, actual)
	})
}
