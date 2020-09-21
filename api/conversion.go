/*
 * Nuts registry
 * Copyright (C) 2020. Nuts community
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
	"fmt"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/types"

	"github.com/nuts-foundation/nuts-registry/pkg/db"
)

func (e Endpoint) fromDb(db db.Endpoint) Endpoint {
	e.URL = db.URL
	e.Organization = Identifier(db.Organization.String())
	e.EndpointType = db.EndpointType
	e.Identifier = Identifier(db.Identifier)
	e.Status = db.Status
	e.Properties = toEndpointProperties(db.Properties)
	return e
}

func (o Organization) fromDb(db db.Organization) Organization {
	e := endpointsFromDb(db.Endpoints)
	o.Identifier = Identifier(db.Identifier.String())
	o.Name = db.Name
	o.PublicKey = db.PublicKey
	o.Endpoints = &e

	if len(db.Keys) == 0 {
		return o
	}

	keys := make([]JWK, len(db.Keys))

	for i, k := range db.Keys {
		keys[i] = JWK{AdditionalProperties: k.(map[string]interface{})}
	}

	o.Keys = &keys

	return o
}

func (o Organization) toDb() db.Organization {
	id, _ := core.ParsePartyID(o.Identifier.String())
	org := db.Organization{
		Identifier: id,
		Name:       o.Name,
		PublicKey:  o.PublicKey,
	}

	if o.Keys != nil {
		org.Keys = jwkToMap(*o.Keys)
	}

	if o.Endpoints != nil {
		org.Endpoints = endpointsToDb(*o.Endpoints)
	}

	return org
}

func (e Endpoint) toDb() db.Endpoint {
	organizationID, _ := core.ParsePartyID(e.Organization.String())
	return db.Endpoint{
		URL:          e.URL,
		EndpointType: e.EndpointType,
		Identifier:   types.EndpointID(e.Identifier),
		Organization: organizationID,
		Status:       e.Status,
		Properties:   fromEndpointProperties(e.Properties),
	}
}

func fromEndpointProperties(endpointProperties *EndpointProperties) map[string]string {
	props := make(map[string]string, 0)
	if endpointProperties != nil {
		for key, value := range endpointProperties.AdditionalProperties {
			props[key] = fmt.Sprintf("%s", value)
		}
	}
	return props
}

func toEndpointProperties(properties map[string]string) *EndpointProperties {
	props := map[string]string{}
	for key, value := range properties {
		props[key] = value
	}
	return &EndpointProperties{AdditionalProperties: props}
}

func jwkToMap(jwk []JWK) []interface{} {
	em := make([]interface{}, len(jwk))
	for i, k := range jwk {
		em[i] = k
	}
	return em
}

func endpointsFromDb(endpointsIn []db.Endpoint) []Endpoint {
	es := make([]Endpoint, len(endpointsIn))
	for i, a := range endpointsIn {
		es[i] = Endpoint{}.fromDb(a)
	}
	return es
}

func organizationsToDb(organizationsIn []Organization) []db.Organization {
	os := make([]db.Organization, len(organizationsIn))
	for i, a := range organizationsIn {
		os[i] = a.toDb()
	}
	return os
}

func endpointsToDb(endpointsIn []Endpoint) []db.Endpoint {
	es := make([]db.Endpoint, len(endpointsIn))
	for i, a := range endpointsIn {
		es[i] = a.toDb()
	}
	return es
}
