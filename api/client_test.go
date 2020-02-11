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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RoundTripFunc
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

type handler struct {
	statusCode int
	bytes      []byte
}

func (h handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(h.statusCode)
	writer.Write(h.bytes)
}

var genericError = []byte("error reason")

func TestHttpClient_OrganizationById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: 404, bytes: genericError})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.OrganizationById("id")

		expected := "registry returned 404, reason: error reason"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got [%v]", expected, err)
		}
	})

	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations[0])
		s := httptest.NewServer(handler{statusCode: 200, bytes: org})
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
		s := httptest.NewServer(handler{statusCode: 200, bytes: org})
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
		s := httptest.NewServer(handler{statusCode: 200, bytes: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		res, err := c.ReverseLookup("name")

		if assert.Nil(t, err) {
			assert.Equal(t, organizations[0], *res)
		}
	})

	t.Run("404", func(t *testing.T) {
		s := httptest.NewServer(handler{statusCode: 404})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.ReverseLookup("name")

		if assert.NotNil(t, err) {
			assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		}
	})

	t.Run("too many results", func(t *testing.T) {
		org, _ := json.Marshal(organizations)
		s := httptest.NewServer(handler{statusCode: 200, bytes: org})
		c := HttpClient{ServerAddress: s.URL, Timeout: time.Second}

		_, err := c.ReverseLookup("name")

		if assert.NotNil(t, err) {
			assert.True(t, errors.Is(err, ErrOrganizationNotFound))
		}
	})
}

func TestHttpClient_EndpointsByOrganizationAndType(t *testing.T) {
	t.Run("200", func(t *testing.T) {
		eps, _ := json.Marshal(endpoints)
		s := httptest.NewServer(handler{statusCode: 200, bytes: eps})
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
