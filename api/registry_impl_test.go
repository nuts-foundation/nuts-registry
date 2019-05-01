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
	"github.com/labstack/echo"
	"github.com/nuts-foundation/nuts-registry/generated"
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
	endpoints []generated.Endpoint
	organizations []generated.Organization
	endpointsError error
}

func (db *MockDb) FindEndpointsByOrganization(organizationIdentifier string) ([]generated.Endpoint, error) {
	if db.endpointsError != nil {
		return nil, db.endpointsError
	}

	return db.endpoints, nil
}

func (db *MockDb) Load(location string) error {
	return nil
}

func (db *MockDb) SearchOrganizations(query string) []generated.Organization {
	return db.organizations
}

var endpoints = []generated.Endpoint{
	{
		Identifier:generated.Identifier{System:"system", Value:"value"},
		EndpointType: "type#value",
	},
}

var organizations = []generated.Organization{
	{
		Identifier:generated.Identifier{System:"system", Value:"value"},
	},
}

func initEcho(db *MockDb) (*echo.Echo, *generated.ServerInterfaceWrapper) {
	e := echo.New()
	stub:= ApiResource{Db: db}
	wrapper := &generated.ServerInterfaceWrapper{
		Handler: stub,
	}
	e.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)
	e.GET("/api/organizations", wrapper.SearchOrganizations)

	return e, wrapper
}

func deserializeEndpoints(data *bytes.Buffer) ([]generated.Endpoint, error) {
	var stub []generated.Endpoint
	err := json.Unmarshal(data.Bytes(), &stub)

	if err != nil {
		return nil, err
	}

	return stub, err
}

func deserializeOrganizations(data *bytes.Buffer) ([]generated.Organization, error) {
	var stub []generated.Organization
	err := json.Unmarshal(data.Bytes(), &stub)

	if err != nil {
		return nil, err
	}

	return stub, err
}

func TestApiResource_EndpointsByOrganisationId(t *testing.T) {
	t.Run("200", func(t *testing.T){
		e, wrapper := initEcho(&MockDb{endpoints:endpoints})

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

		var result []generated.Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}

		if result[0].Identifier.String() != "system#value" {
			t.Errorf("Got result with Identifier: [%s], want [system#value]", result[0].Identifier.String())
		}
	})

	t.Run("by Id and type 200", func(t *testing.T){
		e, wrapper := initEcho(&MockDb{endpoints:endpoints})

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

		var result []generated.Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}

		if result[0].Identifier.String() != "system#value" {
			t.Errorf("Got result with Identifier: [%s], want [system#value]", result[0].Identifier.String())
		}
	})

	t.Run("by Id and type 200 empty result", func(t *testing.T){
		e, wrapper := initEcho(&MockDb{endpoints:endpoints})

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

		var result []generated.Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})

	t.Run("by Id 200 empty result", func(t *testing.T){
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

		var result []generated.Endpoint
		result, err = deserializeEndpoints(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})

	t.Run("internal error on missing organization returns empty 200 result", func(t *testing.T){
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

		var result []generated.Endpoint
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
	t.Run("200", func(t *testing.T){
		e, wrapper := initEcho(&MockDb{organizations:organizations})

		q := make(url.Values)
		q.Set("query", "system#value")

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

		var result []generated.Organization
		result, err = deserializeOrganizations(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 1 {
			t.Errorf("Got result size: %d, want 1", len(result))
		}

		if result[0].Identifier.String() != "system#value" {
			t.Errorf("Got result with Identifier: [%s], want [system#value]", result[0].Identifier.String())
		}
	})

	t.Run("200 with empty list", func(t *testing.T){
		e, wrapper := initEcho(&MockDb{})

		q := make(url.Values)
		q.Set("query", "system#value")

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

		var result []generated.Organization
		result, err = deserializeOrganizations(rec.Body)

		if err != nil {
			t.Errorf("Got err during deserialization: %s", err.Error())
		}

		if len(result) != 0 {
			t.Errorf("Got result size: %d, want 0", len(result))
		}
	})
}