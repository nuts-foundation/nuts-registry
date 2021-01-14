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

package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"github.com/nuts-foundation/nuts-registry/internal/api"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/test"

	"github.com/stretchr/testify/assert"
)

type handler struct {
	statusCode   int
	responseData []byte
}

func (h handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(h.statusCode)
	writer.Write(h.responseData)
}

var genericError = []byte("error reason")

type TestCase struct {
	Name          string
	ResponseCode  int
	ResponseData  []byte
	ExpectedError error
	ExpectedResult interface{}
}

func OKTestCase(name string, responseCode int, responseData string) TestCase {
	return TestCase{
		Name:         name,
		ResponseCode: responseCode,
		ResponseData: []byte(responseData),
	}
}

func TestHttpClient_Create(t *testing.T) {
	testCases := []TestCase{
		OKTestCase("ok", http.StatusOK, "{}"),
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			server := httptest.NewServer(handler{statusCode: testCase.ResponseCode, responseData: testCase.ResponseData})
			client := HttpClient{ServerAddress: server.URL, Timeout: time.Second}

			createdDID, err := client.Create()
			if testCase.ExpectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, createdDID)
			} else {
				assert.EqualError(t, err, testCase.ExpectedError.Error())
				assert.Nil(t, createdDID)
			}
		})
	}
}

func TestHttpClient_OrganizationById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusNotFound, responseData: genericError})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.OrganizationById(test.OrganizationID("id"))

		assert.EqualError(t, err, "registry returned HTTP 404 (expected: 200), response: error reason", "error")
	})

	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations[0])
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.OrganizationById(test.OrganizationID("id"))

		if err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if res.Identifier != organizations[0].Identifier {
			t.Errorf("Expected return organization identifier to be [%s], got [%s]", organizations[0].Identifier, res.Identifier)
		}
	})

	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.OrganizationById(test.OrganizationID("id"))
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}

func TestHttpClient_VendorById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusNotFound, responseData: genericError})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.VendorById(test.VendorID("id"))

		assert.EqualError(t, err, "registry returned HTTP 404 (expected: 200), response: error reason", "error")
	})

	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(vendors[0])
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.VendorById(test.VendorID("id"))

		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, vendors[0].Identifier, res.Identifier)
	})
}

func TestHttpClient_SearchOrganizations(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.SearchOrganizations("query")

		if assert.Nil(t, err) {
			assert.Equal(t, 2, len(res))
		}
	})
}

func TestHttpClient_ReverseLookup(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations[0:1])
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.ReverseLookup("name")

		if assert.Nil(t, err) {
			assert.Equal(t, organizations[0], *res)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusNotFound})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.ReverseLookup("name")

		if assert.NotNil(t, err) {
			assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		}
	})

	t.Run("too many results", func(t *testing.T) {
		org, _ := json.Marshal(organizations)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.ReverseLookup("name")

		if assert.NotNil(t, err) {
			assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		}
	})
}

func TestHttpClient_VendorClaim(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		key := map[string]interface{}{
			"e": 12345,
		}
		event, err := c.VendorClaim(test.OrganizationID("orgID"), "name", []interface{}{key})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.VendorClaim(test.OrganizationID("orgID"), "name", []interface{}{})
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.VendorClaim(test.OrganizationID("orgID"), "name", []interface{}{})
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}

func TestHttpClient_VendorCAs(t *testing.T) {
	t.Run("not implemented", func(t *testing.T) {
		c := HttpClient{ServerAddress: "", Timeout: time.Second}
		vendorCAs := c.VendorCAs()
		assert.Empty(t, vendorCAs)
	})
}

func TestHttpClient_RegisterVendor(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	certificateAsDER := test.GenerateCertificateEx(time.Now(), 2, privateKey)
	certificate, _ := x509.ParseCertificate(certificateAsDER)
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{}, nil)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}
		vendor, err := c.RegisterVendor(certificate)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, vendor)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}
		event, err := c.RegisterVendor(certificate)
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.RegisterVendor(certificate)
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}

func TestHttpClient_RefreshOrganizationCertificate(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RefreshOrganizationCertificate(test.OrganizationID("1234"))
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RefreshOrganizationCertificate(test.OrganizationID("1234"))
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.RefreshOrganizationCertificate(test.OrganizationID("1234"))
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}

func TestHttpClient_RegisterEndpoint(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{}, nil)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RegisterEndpoint(test.OrganizationID("orgId"), "id", "url", "type", "status", map[string]string{"foo": "bar"})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RegisterEndpoint(test.OrganizationID("orgId"), "id", "url", "type", "status", nil)
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.RegisterEndpoint(test.OrganizationID("orgId"), "id", "url", "type", "status", map[string]string{"foo": "bar"})
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}

func TestHttpClient_Verify(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		response := api.altVerifyResponse{Fix: false, Events: []events.Event{
			events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil),
			events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil),
			events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{}, nil),
		}}
		responseData, _ := json.Marshal(response)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: responseData})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}
		evts, fix, err := c.Verify(true)
		assert.NoError(t, err)
		assert.False(t, fix)
		assert.Len(t, evts, 3)
	})
	t.Run("error - http status 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		evts, fix, err := c.Verify(true)
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, evts)
		assert.False(t, fix)
	})
	t.Run("error - invalid response", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: []byte("foobar")})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		evts, fix, err := c.Verify(true)
		assert.EqualError(t, err, "invalid character 'o' in literal false (expecting 'a')")
		assert.Nil(t, evts)
		assert.False(t, fix)
	})
}

func TestHttpClient_EndpointsByOrganizationAndType(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		eps, _ := json.Marshal(endpoints)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: eps})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.EndpointsByOrganizationAndType(test.OrganizationID("entity"), nil)

		if err != nil {
			t.Errorf("Expected no error, got [%s]", err.Error())
		}

		if len(res) != 1 {
			t.Errorf("Expected 1 Endpoint in return, got [%d]", len(res))
		}
	})
	t.Run("http execution error", func(t *testing.T) {
		c := HttpClient{ServerAddress: "localhost:9876", Timeout: time.Second}
		event, err := c.EndpointsByOrganizationAndType(test.OrganizationID("entity"), nil)
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, event)
	})
}
