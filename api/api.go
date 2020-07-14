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
	"fmt"
	"net/http"
	"net/url"

	"github.com/nuts-foundation/nuts-registry/pkg/events"

	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
)

// String converts an identifier to string
func (i Identifier) String() string {
	return string(i)
}

// ApiWrapper is needed to connect the implementation to the echo ServiceWrapper
type ApiWrapper struct {
	R pkg.RegistryClient
}

// To unmarshal an event EventFromJSON must be used (since we expose the interface rather than the internal struct).
// However, since we can't instruct Go to use a particular function to unmarshal list entries, we create a type alias
// for the []events.Event and implement json.Unmarshaler to allow Go to unmarshal the list.
type listOfEvents []events.Event

func (l *listOfEvents) UnmarshalJSON(data []byte) error {
	evts, err := events.EventsFromJSON(data)
	if err != nil {
		return err
	}
	*l = evts
	return nil
}

// altVerifyResponse is alternative, unmarshallable version of VerifyResponse in generated.go for client-side usage.
// Our OpenAPI code generator generates a struct for Event completely separate from our (thoroughly tested) implementation
// events.Event. Using the generator's struct would require elaborate conversion code, which would only be a potential source of bugs.
type altVerifyResponse struct {
	Events listOfEvents
	Fix    bool
}

func (apiResource ApiWrapper) Verify(ctx echo.Context, params VerifyParams) error {
	var fix = false
	if params.Fix != nil {
		fix = *params.Fix
	}
	resultingEvents, needsFixing, err := apiResource.R.Verify(fix)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, altVerifyResponse{Events: resultingEvents, Fix: needsFixing})
}

// DeprecatedVendorClaim is deprecated, use VendorClaim.
func (apiResource ApiWrapper) DeprecatedVendorClaim(ctx echo.Context, _ string) error {
	return apiResource.VendorClaim(ctx)
}

func (apiResource ApiWrapper) RefreshOrganizationCertificate(ctx echo.Context, id string) error {
	event, err := apiResource.R.RefreshOrganizationCertificate(id)
	if errors.Is(err, ErrOrganizationNotFound) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

func (apiResource ApiWrapper) RefreshVendorCertificate(ctx echo.Context) error {
	event, err := apiResource.R.RefreshVendorCertificate()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// RegisterEndpoint is the Api implementation for registering an endpoint.
func (apiResource ApiWrapper) RegisterEndpoint(ctx echo.Context, id string) error {
	unescapedID, err := url.PathUnescape(id)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	ep := Endpoint{}
	err = json.Unmarshal(bytes, &ep)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err = ep.validate(); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	event, err := apiResource.R.RegisterEndpoint(unescapedID, ep.Identifier.String(), ep.URL, ep.EndpointType, ep.Status, fromEndpointProperties(ep.Properties))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// VendorClaim is the Api implementation for registering a vendor claim.
func (apiResource ApiWrapper) VendorClaim(ctx echo.Context) error {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	org := Organization{}
	err = json.Unmarshal(bytes, &org)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err = org.validate(); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	var keys []interface{}
	if org.Keys != nil {
		keys = jwkToMap(*org.Keys)
	}
	event, err := apiResource.R.VendorClaim(org.Identifier.String(), org.Name, keys)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// RegisterVendor is the Api implementation for registering a vendor.
func (apiResource ApiWrapper) RegisterVendor(ctx echo.Context) error {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	v := Vendor{}
	if err := json.Unmarshal(bytes, &v); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err := v.validate(); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	event, err := apiResource.R.RegisterVendor(v.Name, string(v.Domain))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// OrganizationById is the Api implementation for getting an organization based on its Id.
func (apiResource ApiWrapper) OrganizationById(ctx echo.Context, id string) error {

	unescaped, err := url.PathUnescape(id)

	if err != nil {
		return err
	}

	result, err := apiResource.R.OrganizationById(unescaped)
	if result == nil {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Could not find organization with id %v", unescaped))
	}

	return ctx.JSON(http.StatusOK, Organization{}.fromDb(*result))
}

// EndpointsByOrganisationId is the Api implementation for getting all or certain types of endpoints for an organization
func (apiResource ApiWrapper) EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error {
	foundEPs := []Endpoint{}
	strict := params.Strict
	for _, id := range params.OrgIds {
		dbEndpoints, err := apiResource.R.EndpointsByOrganizationAndType(id, params.Type)

		if err != nil {
			logrus.Warning(err.Error())
		} else {
			foundEPs = append(endpointsFromDb(dbEndpoints), foundEPs...)
		}

		if strict != nil && *strict && len(dbEndpoints) == 0 {
			var t = ""
			if params.Type != nil {
				t = *params.Type
			}
			return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("organization with id %s does not have an endpoint of type %s", id, t))
		}
	}

	// filter on type
	filtered := []Endpoint{}
	if params.Type == nil {
		filtered = foundEPs
	} else {
		for _, u := range foundEPs {
			if u.EndpointType == *params.Type {
				filtered = append(filtered, u)
			}
		}
	}

	// generate output
	return ctx.JSON(http.StatusOK, filtered)
}

// SearchOrganizations is the Api implementation for finding organizations by (partial) query
func (apiResource ApiWrapper) SearchOrganizations(ctx echo.Context, params SearchOrganizationsParams) error {

	var (
		searchResult []db.Organization
		org          *db.Organization
		err          error
	)

	if params.Exact != nil && *params.Exact {
		org, err = apiResource.R.ReverseLookup(params.Query)

		if org != nil {
			searchResult = append(searchResult, *org)
		}
	} else {
		searchResult, err = apiResource.R.SearchOrganizations(params.Query)
	}

	if errors.Is(err, db.ErrOrganizationNotFound) {
		return ctx.NoContent(http.StatusNotFound)
	}

	if err != nil {
		return err
	}

	result := make([]Organization, len(searchResult))
	for i, o := range searchResult {
		result[i] = Organization{}.fromDb(o)
	}

	return ctx.JSON(http.StatusOK, result)
}

func (apiResource ApiWrapper) MTLSCAs(ctx echo.Context) error {
	var err error

	return err
}

func (apiResource ApiWrapper) MTLSCertificates(ctx echo.Context) error {
	var err error

	return err
}
