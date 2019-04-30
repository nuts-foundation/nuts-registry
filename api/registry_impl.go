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
	"github.com/golang/glog"
	"github.com/labstack/echo"
	"github.com/nuts-foundation/nuts-registry/db"
	"github.com/nuts-foundation/nuts-registry/generated"
	"net/http"
)

type ApiResource struct{
	Db db.Db
}

func (apiResource ApiResource) EndpointsByOrganisationId(ctx echo.Context, params generated.EndpointsByOrganisationIdParams) error {
	var err error

	var dupEndpoints []generated.Endpoint
	var endpoints []generated.Endpoint
	endpointIds := make(map[string]bool)
	for _, id := range params.OrgIds {
		endpoints, err = apiResource.Db.FindEndpointsByOrganisation(id)

		if err != nil{
			glog.Warning(err.Error())
		} else {
			dupEndpoints = append(dupEndpoints, endpoints...)
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
	uniqFiltered := uniq[0:]
	if params.Type != nil {
		for i, u := range uniqFiltered {
			if u.EndpointType != *params.Type {
				uniqFiltered = append(uniqFiltered[:i], uniqFiltered[i+1:]...)
			}
		}
	}

	// generate output
	return ctx.JSON(http.StatusOK, uniqFiltered)
}

func (apiResource ApiResource) SearchOrganizations(ctx echo.Context, params generated.SearchOrganizationsParams) error {

	result := apiResource.Db.SearchOrganizations(params.Query)

	if result == nil {
		result = []generated.Organization{}
	}

	return ctx.JSON(http.StatusOK, result)
}