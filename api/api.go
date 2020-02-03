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
	"io/ioutil"
	"net/http"
	"net/url"

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
	R *pkg.Registry
}

// DeregisterOrganization is the api implementation of removing an organization from the registry
func (apiResource ApiWrapper) DeregisterOrganization(ctx echo.Context, id string) error {
	result, err := apiResource.R.OrganizationById(id)

	if result == nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	err = apiResource.R.RemoveOrganization(id)

	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusAccepted)
}

// RegisterOrganization is the api implementation for adding a new organization to the registry
func (apiResource ApiWrapper) RegisterOrganization(ctx echo.Context) error {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)

	if err != nil {
		return err
	}

	org := Organization{}

	err = json.Unmarshal(bytes, &org)

	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	if len(org.Identifier) == 0 {
		return ctx.String(http.StatusBadRequest, "missing identifier on organization")
	}

	result, err := apiResource.R.OrganizationById(string(org.Identifier))

	if result != nil {
		return ctx.String(http.StatusBadRequest, "duplicate organization for identifier")
	}

	err = apiResource.R.RegisterOrganization(org.toDb())

	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusCreated)
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
	var dupEndpoints []Endpoint
	strict := params.Strict
	endpointIds := make(map[string]bool)
	for _, id := range params.OrgIds {
		dbEndpoints, err := apiResource.R.EndpointsByOrganizationAndType(id, params.Type)

		if err != nil {
			logrus.Warning(err.Error())
		} else {
			dupEndpoints = append(endpointsArrayFromDb(dbEndpoints), dupEndpoints...)
		}

		if strict != nil && *strict && len(dbEndpoints) == 0 {
			return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("organization with id %s does not have an endpoint of type %v", id, params.Type))
		}
	}

	// deduplicate
	uniq := dupEndpoints[:0]
	for _, e := range dupEndpoints {
		_, f := endpointIds[e.Identifier.String()]
		if !f {
			endpointIds[e.Identifier.String()] = true
			uniq = append(uniq, e)
		}
	}

	// filter on type
	var uniqFiltered []Endpoint
	if params.Type == nil {
		uniqFiltered = uniq
	} else {
		for _, u := range uniq {
			if u.EndpointType == *params.Type {
				uniqFiltered = append(uniqFiltered, u)
			}
		}
	}

	// generate output
	return ctx.JSON(http.StatusOK, uniqFiltered)
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

	if result == nil {
		result = []Organization{}
	}

	return ctx.JSON(http.StatusOK, result)
}
