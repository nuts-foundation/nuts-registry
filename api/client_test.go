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
	"encoding/json"
	"errors"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestHttpClient_OrganizationById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusNotFound, responseData: genericError})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.OrganizationById("id")

		assert.EqualError(t, err, "registry returned HTTP 404 (expected: 200), response: error reason", "error")
	})

	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations[0])
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.OrganizationById("id")

		if err != nil {
			t.Errorf("Expected no error, got [%v]", err)
		}

		if res.Identifier != organizations[0].Identifier {
			t.Errorf("Expected return organization identifier to be [%s], got [%s]", organizations[0].Identifier, res.Identifier)
		}
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
		event := events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{})
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		key := map[string]interface{}{
			"e": 12345,
		}
		event, err := c.VendorClaim("id", "orgID", "name", []interface{}{key})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.VendorClaim("id", "orgID", "name", []interface{}{})
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
}

func TestHttpClient_RegisterVendor(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{})
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		vendor, err := c.RegisterVendor("id", "name", "")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, vendor)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RegisterVendor("id", "name", "")
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
}

func TestHttpClient_RegisterEndpoint(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{})
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: event.Marshal()})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RegisterEndpoint("orgId", "id", "url", "type", "status", "version", map[string]string{"foo": "bar"})
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, event)
	})
	t.Run("error 500", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: http.StatusInternalServerError, responseData: []byte{}})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		event, err := c.RegisterEndpoint("orgId", "id", "url", "type", "status", "version", nil)
		assert.EqualError(t, err, "registry returned HTTP 500 (expected: 200), response: ", "error")
		assert.Nil(t, event)
	})
}

func TestHttpClient_EndpointsByOrganizationAndType(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		eps, _ := json.Marshal(endpoints)
		s := httptest.NewServer(handler{statusCode: http.StatusOK, responseData: eps})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.EndpointsByOrganizationAndType("entity", nil)

		if err != nil {
			t.Errorf("Expected no error, got [%s]", err.Error())
		}

		if len(res) != 1 {
			t.Errorf("Expected 1 Endpoint in return, got [%d]", len(res))
		}
	})
}
