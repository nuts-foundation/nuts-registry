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

package api

import (
	"bytes"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"

	"net/http"
	"net/http/httptest"
	"testing"
)

type testError struct {
	s string
}

func newTestError(text string) error {
	return &testError{text}
}

func (e *testError) Error() string {
	return e.s
}

type MockDb struct {
	endpoints      []db.Endpoint
	organizations  []db.Organization
	endpointsError error
}

func (mdb *MockDb) VendorByID(id string) *db.Vendor {
	panic("implement me")
}

func (mdb *MockDb) RegisterEventHandlers(system events.EventSystem) {

}

func (mdb *MockDb) FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]db.Endpoint, error) {
	if mdb.endpointsError != nil {
		return nil, mdb.endpointsError
	}

	return mdb.endpoints, nil
}

func (mdb *MockDb) SearchOrganizations(query string) []db.Organization {
	return mdb.organizations
}

func (mdb *MockDb) ReverseLookup(name string) (*db.Organization, error) {
	if len(mdb.organizations) > 0 {
		return &mdb.organizations[0], nil
	}
	return nil, db.ErrOrganizationNotFound
}

func (mdb *MockDb) OrganizationById(id string) (*db.Organization, error) {
	if len(mdb.organizations) > 0 {
		return &mdb.organizations[0], nil
	}

	return nil, nil
}

var endpoints = []db.Endpoint{
	{
		Identifier:   db.Identifier("urn:nuts:system:value"),
		EndpointType: "type#value",
	},
}

var multiEndpoints = []db.Endpoint{
	{
		Identifier:   db.Identifier("urn:nuts:system:value"),
		EndpointType: "type#value",
	},
	{
		Identifier:   db.Identifier("urn:nuts:system:value2"),
		EndpointType: "type#value",
	},
}

var key = map[string]interface{}{"kty": "EC"}
var organizations = []db.Organization{
	{
		Identifier: db.Identifier("urn:nuts:system:value"),
		Name:       "test",
	},
	{
		Identifier: db.Identifier("urn:nuts:hidden"),
		Name:       "hidden",
		Keys: []interface{}{
			key,
		},
	},
}

func initEcho(db *MockDb) (*echo.Echo, *ServerInterfaceWrapper) {
	e := echo.New()
	stub := ApiWrapper{R: &pkg.Registry{Db: db, EventSystem: events.NewEventSystem()}}
	wrapper := &ServerInterfaceWrapper{
		Handler: stub,
	}

	return e, wrapper
}

func initMockEcho(registryClient *mock.MockRegistryClient) (*echo.Echo, *ServerInterfaceWrapper) {
	e := echo.New()
	stub := ApiWrapper{R: registryClient}
	wrapper := &ServerInterfaceWrapper{
		Handler: stub,
	}

	return e, wrapper
}

func deserializeEndpoints(data *bytes.Buffer) ([]Endpoint, error) {
	var stub []Endpoint
	err := json.Unmarshal(data.Bytes(), &stub)

	if err != nil {
		return nil, err
	}

	return stub, err
}

func deserializeOrganizations(data *bytes.Buffer) ([]Organization, error) {
	var stub []Organization
	err := json.Unmarshal(data.Bytes(), &stub)

	if err != nil {
		return nil, err
	}

	return stub, err
}

func deserializeOrganization(data *bytes.Buffer) (*Organization, error) {
	stub := &Organization{}
	err := json.Unmarshal(data.Bytes(), stub)

	if err != nil {
		return nil, err
	}

	return stub, err
}

func TestIdentifier_String(t *testing.T) {
	i := Identifier("urn:nuts:system:value")

	if i.String() != "urn:nuts:system:value" {
		t.Errorf("Expected [urn:nuts:system:value], got [%s]", i.String())
	}
}

func TestApiResource_EndpointsByOrganisationId(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: endpoints})

		q := make(url.Values)
		q.Set("orgIds", "1")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}

		if result[0].Identifier.String() != "urn:nuts:system:value" {
			t.Errorf("Got result with Identifier: [%s], want [urn:nuts:system:value]", result[0].Identifier.String())
		}
	})

	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: multiEndpoints})

		q := make(url.Values)
		q.Set("orgIds", "1")
		q.Add("orgIds", "2")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		wrapper.EndpointsByOrganisationId(c)

		var result []Endpoint
		result, _ = deserializeEndpoints(rec.Body)

		if len(result) != 2 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}
	})

	t.Run("by Id and type 200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: endpoints})

		q := make(url.Values)
		q.Set("orgIds", "1")
		q.Set("type", "type#value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}

		if result[0].Identifier.String() != "urn:nuts:system:value" {
			t.Errorf("Got result with Identifier: [%s], want [urn:nuts:system:value]", result[0].Identifier.String())
		}
	})

	t.Run("by Id and type 200 empty result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: endpoints})

		q := make(url.Values)
		q.Set("orgIds", "1")
		q.Set("type", "otherType#value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})

	t.Run("by Id and type 400 strict", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("orgIds", "1")
		q.Set("type", "otherType#value")
		q.Set("strict", "true")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var responseText = ""
		err = json.Unmarshal(rec.Body.Bytes(), &responseText)
		assert.NoError(t, err)
		assert.Equal(t, "organization with id 1 does not have an endpoint of type otherType#value", responseText)
	})

	t.Run("by Id 200 empty result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("orgIds", "1")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})

	t.Run("internal error on missing organization returns empty 200 result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpointsError: newTestError("error")})

		q := make(url.Values)
		q.Set("orgIds", "1")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})
}

func TestApiResource_SearchOrganizations(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		q := make(url.Values)
		q.Set("query", "urn:nuts:system:value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		if assert.Nil(t, err) && assert.Equal(t, http.StatusOK, rec.Code) {

			var result []Organization
			result, err = deserializeOrganizations(rec.Body)

			assert.Nil(t, err)
			assert.Equal(t, 2, len(result))
			assert.Equal(t, "urn:nuts:system:value", result[0].Identifier.String())
		}
	})

	t.Run("200 with empty list", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("query", "urn:nuts:system:value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		if assert.Nil(t, err) && assert.Equal(t, http.StatusOK, rec.Code) {

			var result []Organization
			result, err = deserializeOrganizations(rec.Body)

			assert.Nil(t, err)
			assert.Equal(t, 0, len(result))
		}
	})
}

func TestApiResource_ReverseLookup(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		q := make(url.Values)
		q.Set("query", "test")
		q.Set("exact", "true")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		if assert.Nil(t, err) && assert.Equal(t, http.StatusOK, rec.Code) {

			result, err := deserializeOrganizations(rec.Body)

			assert.Nil(t, err)
			assert.NotNil(t, result)
		}
	})

	t.Run("404", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("query", "que?")
		q.Set("exact", "true")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestApiResource_OrganizationById(t *testing.T) {
	t.Run("404 when not found", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: []db.Organization{}})

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id")
		c.SetParamNames("id")
		c.SetParamValues("https%3A//system%23value")

		err := wrapper.OrganizationById(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}
	})
	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id")
		c.SetParamNames("id")
		c.SetParamValues("https%3A//system%23value")

		err := wrapper.OrganizationById(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		result, err := deserializeOrganization(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if result == nil {
			t.Error("Got nil from deserialization")
		}

		if result.Name != "test" {
			t.Errorf("Got result with Name: [%s], want [test]", result.Name)
		}
	})
}

func TestApiResource_RegisterVendor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("register vendor", func(t *testing.T) {
		t.Run("200", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().RegisterVendor("abc", "def", "xyz")

			b, _ := json.Marshal(Vendor{
				Identifier: "abc",
				Name:       "def",
				Domain:     "xyz",
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendors")

			err := wrapper.RegisterVendor(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
			}
		})

		t.Run("400 - Invalid JSON", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{{[[][}{"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendors")

			err := wrapper.RegisterVendor(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})

		t.Run("400 - validation failed", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			b, _ := json.Marshal(Vendor{})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendors")

			err := wrapper.RegisterVendor(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	})
}

func TestApiResource_VendorClaim(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("vendor claim", func(t *testing.T) {
		t.Run("204", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().VendorClaim("1", "abc", "def", gomock.Any())

			b, _ := json.Marshal(Organization{
				Identifier: "abc",
				Name:       "def",
				Keys:       &[]JWK{map[string]interface{}{}},
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendor/:id/claim")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.VendorClaim(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
			}
		})

		t.Run("400 - Invalid JSON", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{{[[][}{"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendor/:id/claim")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.VendorClaim(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})

		t.Run("400 - validation failed", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			b, _ := json.Marshal(Organization{})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendor/:id/claim")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.VendorClaim(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	})
}

func TestApiResource_RegisterEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("register endpoint", func(t *testing.T) {
		t.Run("204", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().RegisterEndpoint("1", "", "foo:bar", "fhir", "", map[string]string{"key": "value"})

			props := EndpointProperties{}
			props["key"] = "value"
			b, _ := json.Marshal(Endpoint{
				Identifier:   "",
				URL:          "foo:bar",
				EndpointType: "fhir",
				Properties:   &props,
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/endpoints")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.RegisterEndpoint(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusOK {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
			}
		})

		t.Run("400 - Invalid JSON", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{{[[][}{"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/endpoints")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.RegisterEndpoint(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})

		t.Run("400 - validation failed", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			b, _ := json.Marshal(Endpoint{})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/endpoints")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.RegisterVendor(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	})
}
