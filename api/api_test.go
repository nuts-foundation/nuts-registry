/*
 * Nuts registry
 * Copyright (C) 2019 Nuts community
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
 */

package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"net/url"

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

func (db *MockDb) FindEndpointsByOrganization(organizationIdentifier string) ([]db.Endpoint, error) {
	if db.endpointsError != nil {
		return nil, db.endpointsError
	}

	return db.endpoints, nil
}

func (db *MockDb) Load(location string) error {
	return nil
}

func (db *MockDb) SearchOrganizations(query string) []db.Organization {
	return db.organizations
}

func (db *MockDb) OrganizationById(id string) (*db.Organization, error) {
	return &db.organizations[0], nil
}

var endpoints = []db.Endpoint{
	{
		Identifier:   db.Identifier("urn:nuts:system:value"),
		EndpointType: "type#value",
	},
}

var organizations = []db.Organization{
	{
		Identifier: db.Identifier("urn:nuts:system:value"),
		Name:       "test",
		Actors: []db.Actor{
			{
				Identifier: db.Identifier("urn:nuts:system:value"),
			},
		},
	},
}

func initEcho(db *MockDb) (*echo.Echo, *ServerInterfaceWrapper) {
	e := echo.New()
	stub := ApiWrapper{R: &pkg.Registry{Db: db}}
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

func deserializeActors(data *bytes.Buffer) ([]Actor, error) {
	var stub []Actor
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

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Organization
		result, err = deserializeOrganizations(rec.Body)

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

	t.Run("200 with empty list", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("query", "urn:nuts:system:value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		var result []Organization
		result, err = deserializeOrganizations(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})
}

func TestApiResource_OrganizationById(t *testing.T) {
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

func TestApiResource_OrganizationActors(t *testing.T) {
	t.Run("200 with single result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		q := make(url.Values)
		q.Set("actorId", "urn:nuts:system:value")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id/actors")
		c.SetParamNames("id")
		c.SetParamValues("urn:nuts:system:value")

		err := wrapper.OrganizationActors(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		result, err := deserializeActors(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if result == nil {
			t.Error("Got nil from deserialization")
		}

		if len(result) != 1 {
			t.Errorf("Got %d results want [1]", len(result))
		}
	})

	t.Run("200 with empty result", func(t *testing.T) {
		e, wrapper := initEcho(&MockDb{organizations: organizations})

		q := make(url.Values)
		q.Set("actorId", "system#value2")

		req := httptest.NewRequest(echo.GET, "/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organization/:id/actors")
		c.SetParamNames("id")
		c.SetParamValues("system#value")

		err := wrapper.OrganizationActors(c)

		if err != nil {
			t.Errorf("Got err during call: %s", err.Error())
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Got status=%d, want %d", rec.Code, http.StatusOK)
		}

		result, err := deserializeActors(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if result == nil {
			t.Error("Got nil from deserialization")
		}

		if len(result) != 0 {
			t.Errorf("Got %d results want [0]", len(result))
		}
	})
}