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

package generated

import (
	"github.com/labstack/echo"
	"net/url"

	"net/http"
	"net/http/httptest"
	"testing"
)

type RestInterfaceStub struct{}

func (e RestInterfaceStub) EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error {
	var err error

	return err
}

func (e RestInterfaceStub) SearchOrganizations(ctx echo.Context, params SearchOrganizationsParams) error {
	var err error

	return err
}


func TestServerInterfaceWrapper_EndpointsByOrganisationId(t *testing.T) {
	t.Run("200", func(t *testing.T){
		e := echo.New()
		stub:= RestInterfaceStub{}
		wrapper := &ServerInterfaceWrapper{
			Handler: stub,
		}
		e.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)

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
	})

	t.Run("400", func(t *testing.T){
		e := echo.New()
		stub:= RestInterfaceStub{}
		wrapper := &ServerInterfaceWrapper{
			Handler: stub,
		}

		e.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)

		req := httptest.NewRequest(echo.GET, "/" , nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/endpoints")

		err := wrapper.EndpointsByOrganisationId(c)

		if (err == nil) {
			t.Errorf("Didn't get expected err during call")
		}

		expected := "code=400, message=Invalid format for parameter orgIds: code=400, message=query parameter 'orgIds' is required"
		if (err != nil && err.Error() != expected) {
			t.Errorf("Got message=%s, want %s", err.Error(), expected)
		}
	})
}

func TestServerInterfaceWrapper_SearchOrganizations(t *testing.T) {
	t.Run("200", func(t *testing.T){
		e := echo.New()
		stub:= RestInterfaceStub{}
		wrapper := &ServerInterfaceWrapper{
			Handler: stub,
		}
		e.GET("/api/organizations", wrapper.SearchOrganizations)

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
	})

	t.Run("400", func(t *testing.T){
		e := echo.New()
		stub:= RestInterfaceStub{}
		wrapper := &ServerInterfaceWrapper{
			Handler: stub,
		}

		e.GET("/api/organizations", wrapper.SearchOrganizations)

		req := httptest.NewRequest(echo.GET, "/" , nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/organizations")

		err := wrapper.SearchOrganizations(c)

		if err == nil {
			t.Errorf("Didn't get expected err during call")
		}

		expected := "code=400, message=Invalid format for parameter query: code=400, message=query parameter 'query' is required"
		if err != nil && err.Error() != expected {
			t.Errorf("Got message=%s, want %s", err.Error(), expected)
		}
	})
}