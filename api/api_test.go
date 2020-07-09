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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"github.com/nuts-foundation/nuts-registry/test"
	"net/url"
	"strings"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/stretchr/testify/assert"

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

func (mdb *MockDb) OrganizationsByVendorID(id core.PartyID) []*db.Organization {
	panic("implement me")
}

func (mdb *MockDb) VendorByID(id core.PartyID) *db.Vendor {
	panic("implement me")
}

func (mdb *MockDb) RegisterEventHandlers(fn events.EventRegistrar) {

}

func (mdb *MockDb) FindEndpointsByOrganizationAndType(organizationIdentifier core.PartyID, endpointType *string) ([]db.Endpoint, error) {
	if mdb.endpointsError != nil {
		return nil, mdb.endpointsError
	}

	// only return relevant EP's
	eps := []db.Endpoint{}
	for _, u := range mdb.endpoints {
		if u.Organization == organizationIdentifier {
			eps = append(eps, u)
		}
	}

	return eps, nil
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

func (mdb *MockDb) OrganizationById(id core.PartyID) (*db.Organization, error) {
	if len(mdb.organizations) > 0 {
		return &mdb.organizations[0], nil
	}

	return nil, nil
}

var endpoints = []db.Endpoint{
	{
		Organization: test.OrganizationID("value"),
		Identifier:   types.EndpointID("urn:nuts:system:value"),
		EndpointType: "type#value",
	},
}

var multiEndpoints = []db.Endpoint{
	{
		Organization: test.OrganizationID("value"),
		Identifier:   types.EndpointID("urn:nuts:system:value"),
		EndpointType: "type#value",
	},
	{
		Organization: test.OrganizationID("value2"),
		Identifier:   types.EndpointID("urn:nuts:system:value2"),
		EndpointType: "type#value",
	},
}

var key = map[string]interface{}{"kty": "EC"}
var organizations = []db.Organization{
	{
		Identifier: test.OrganizationID("value"),
		Name:       "test",
	},
	{
		Identifier: test.OrganizationID("hidden"),
		Name:       "hidden",
		Keys: []interface{}{
			key,
		},
	},
}

func initEcho(db *MockDb) (*echo.Echo, *ServerInterfaceWrapper) {
	e := echo.New()
	stub := ApiWrapper{R: &pkg.Registry{Db: db, EventSystem: events.NewEventSystem(domain.GetEventTypes()...)}}
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

	if stub == nil {
		return nil, errors.New("got nil value")
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
		q.Set("orgIds", test.OrganizationID("value").String())

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
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, result, 1) {
			return
		}
		assert.Equal(t, string(endpoints[0].Identifier), result[0].Identifier.String())
	})

	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: multiEndpoints})

		q := make(url.Values)
		q.Set("orgIds", test.OrganizationID("value").String())
		q.Add("orgIds", test.OrganizationID("value2").String())

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		wrapper.EndpointsByOrganisationId(c)

		var result []Endpoint
		result, _ = deserializeEndpoints(rec.Body)

		if len(result) != 2 {
			t.Errorf("Got result size: %d, want 2", len(result))
		}
	})

	t.Run("by Id and type 200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{endpoints: endpoints})

		q := make(url.Values)
		q.Set("orgIds", test.OrganizationID("value").String())
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
		q.Set("orgIds", test.OrganizationID("1").String())
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
		q.Set("orgIds", test.OrganizationID("1").String())
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
		assert.Equal(t, "organization with id urn:oid:2.16.840.1.113883.2.4.6.1:1 does not have an endpoint of type otherType#value", responseText)
	})

	t.Run("by Id 200 empty result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("orgIds", test.OrganizationID("1").String())

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
		q.Set("orgIds", test.OrganizationID("1").String())

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
		q.Set("query", test.OrganizationID("value").String())

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
			assert.Equal(t, test.OrganizationID("value").String(), result[0].Identifier.String())
		}
	})

	t.Run("200 with empty list", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("query", test.OrganizationID("value").String())

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
		c.SetParamValues("urn:oid:1.2.3:value")

		err := wrapper.OrganizationById(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusNotFound {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}
	})
	t.Run("400 invalid PartyID", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: []db.Organization{}})

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id")
		c.SetParamNames("id")
		c.SetParamValues("https%3A//system%23value")

		err := wrapper.OrganizationById(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("200", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id")
		c.SetParamNames("id")
		c.SetParamValues("urn:oid:1.2.3:value")

		err := wrapper.OrganizationById(c)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Equal(t, http.StatusOK, rec.Code) {
			return
		}
		result, err := deserializeOrganization(rec.Body)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NotNil(t, result) {
			return
		}
		assert.Equal(t, "test", result.Name)
	})
}

func TestApiResource_Verify(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("ok - http status 200", func(t *testing.T) {
		var registryClient = mock.NewMockRegistryClient(mockCtrl)
		e, wrapper := initMockEcho(registryClient)
		registryClient.EXPECT().Verify(true)

		req := httptest.NewRequest(echo.POST, "/?fix=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/admin/verify")

		err := wrapper.Verify(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("error - http status 500", func(t *testing.T) {
		var registryClient = mock.NewMockRegistryClient(mockCtrl)
		e, wrapper := initMockEcho(registryClient)
		registryClient.EXPECT().Verify(true).Return([]events.Event{}, false, errors.New("oops"))

		req := httptest.NewRequest(echo.POST, "/?fix=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/admin/verify")

		_ = wrapper.Verify(c)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestApiResource_RegisterVendor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	certificate := test.GenerateCertificateEx(time.Now(), 2, privateKey)
	certificateAsPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate,
	})

	t.Run("register vendor", func(t *testing.T) {
		t.Run("200", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().RegisterVendor(gomock.Any())

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(certificateAsPEM))
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

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader([]byte{1, 2, 3}))
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
		orgID := test.OrganizationID("abc")
		t.Run("deprecated still works", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().VendorClaim(orgID, "def", gomock.Any())

			b, _ := json.Marshal(Organization{
				Identifier: Identifier(orgID.String()),
				Name:       "def",
				Keys:       &[]JWK{{AdditionalProperties: map[string]interface{}{}}},
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/vendor/:id/claim")
			c.SetParamNames("id")
			c.SetParamValues("1")

			err := wrapper.DeprecatedVendorClaim(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
		t.Run("204", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			registryClient.EXPECT().VendorClaim(orgID, "def", gomock.Any())
			b, _ := json.Marshal(Organization{
				Identifier: Identifier(orgID.String()),
				Name:       "def",
				Keys:       &[]JWK{{AdditionalProperties: map[string]interface{}{}}},
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization")

			err := wrapper.VendorClaim(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("400 - Invalid JSON", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{{[[][}{"))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization")

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
			c.SetPath("/api/organization")

			err := wrapper.VendorClaim(c)

			if err != nil {
				t.Errorf("Got err during call: %s", err.Error())
			}

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Got status=%d, want %d", rec.Code, http.StatusBadRequest)
			}
		})

		t.Run("400 - invalid org ID", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)

			b, _ := json.Marshal(Organization{Identifier: "foobar", Name: "test"})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization")

			err := wrapper.VendorClaim(c)
			assert.NoError(t, err)
			assert.Equal(t, rec.Code, http.StatusBadRequest)
		})
	})
}

func TestApiResource_RefreshOrganizationCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("refresh organization certificate", func(t *testing.T) {
		t.Run("200", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			orgID := test.OrganizationID("1234")
			registryClient.EXPECT().RefreshOrganizationCertificate(orgID)

			req := httptest.NewRequest(echo.POST, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/refresh-cert")
			c.SetParamNames("id")
			c.SetParamValues(orgID.String())

			err := wrapper.RefreshOrganizationCertificate(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
		t.Run("400", func(t *testing.T) {
			var registryClient = mock.NewMockRegistryClient(mockCtrl)
			e, wrapper := initMockEcho(registryClient)
			orgID := test.OrganizationID("1234")
			registryClient.EXPECT().RefreshOrganizationCertificate(orgID).Return(nil, ErrOrganizationNotFound)

			req := httptest.NewRequest(echo.POST, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/refresh-cert")
			c.SetParamNames("id")
			c.SetParamValues(orgID.String())

			err := wrapper.RefreshOrganizationCertificate(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, ErrOrganizationNotFound.Error(), rec.Body.String())
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
			orgID := test.OrganizationID("1234")
			registryClient.EXPECT().RegisterEndpoint(orgID, "", "foo:bar", "fhir", "", map[string]string{"key": "value"})

			props := map[string]string{}
			props["key"] = "value"
			b, _ := json.Marshal(Endpoint{
				Identifier:   "",
				URL:          "foo:bar",
				EndpointType: "fhir",
				Properties:   &EndpointProperties{AdditionalProperties: props},
			})

			req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/organization/:id/endpoints")
			c.SetParamNames("id")
			c.SetParamValues(orgID.String())

			err := wrapper.RegisterEndpoint(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
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

func TestApiResource_MTLSCAs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pk1, _ := rsa.GenerateKey(rand.Reader, 1024)
	pk2, _ := rsa.GenerateKey(rand.Reader, 1024)
	pk3, _ := rsa.GenerateKey(rand.Reader, 1024)
	certBytes := test.GenerateCertificateEx(time.Now().AddDate(0, 0, -1), 2, pk1)
	root, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Intermediate CA", root, pk2, pk1)
	ca, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Vendor CA 1", ca, pk3, pk2)
	vca1, _ := x509.ParseCertificate(certBytes)
	certBytes = test.GenerateCertificateCA("Vendor CA 2", ca, pk3, pk2)
	vca2, _ := x509.ParseCertificate(certBytes)
	combined := fmt.Sprintf("%s\n%s\n", certificateToPEM(root), certificateToPEM(ca))

	t.Run("ok - http status 200 - single pem", func(t *testing.T) {
		var registryClient = mock.NewMockRegistryClient(mockCtrl)
		e, wrapper := initMockEcho(registryClient)
		registryClient.EXPECT().VendorCAs().Return([][]*x509.Certificate{{vca1, ca, root}, {vca2, ca, root}})

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/mtls/cas")

		err := wrapper.MTLSCAs(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Equal(t, 0, strings.Index(body, combined))
		assert.True(t, strings.Contains(body, certificateToPEM(vca1)))
		assert.True(t, strings.Contains(body, certificateToPEM(vca2)))
		assert.Equal(t, "application/x-pem-file", rec.Result().Header.Get("Content-Type"))
	})

	t.Run("ok - http status 200 - json", func(t *testing.T) {
		var registryClient = mock.NewMockRegistryClient(mockCtrl)
		e, wrapper := initMockEcho(registryClient)
		registryClient.EXPECT().VendorCAs().Return([][]*x509.Certificate{{vca1, ca, root}, {vca2, ca, root}})

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Accept", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/mtls/cas")

		err := wrapper.MTLSCAs(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		cAListWithChain := CAListWithChain{}
		json.Unmarshal(rec.Body.Bytes(), &cAListWithChain)
		assert.Len(t, cAListWithChain.Chain, 2)
		assert.Equal(t, certificateToPEM(root), cAListWithChain.Chain[0])
		assert.Len(t, cAListWithChain.CAList, 2)
	})
}

func Test_listOfEvents(t *testing.T) {
	t.Run("ok - unmarshal", func(t *testing.T) {
		input := []events.Event{
			events.CreateEvent("foobar", struct{}{}, nil),
			events.CreateEvent("foobar", struct{}{}, nil),
		}
		data, _ := json.Marshal(input)
		list := listOfEvents{}
		err := json.Unmarshal(data, &list)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, list, 2)
	})
	t.Run("error - unmarshal", func(t *testing.T) {
		list := listOfEvents{}
		err := json.Unmarshal([]byte("{}"), &list)
		assert.Error(t, err)
		assert.Empty(t, list)
	})
}
