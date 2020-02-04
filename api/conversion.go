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

import "github.com/nuts-foundation/nuts-registry/pkg/db"

func (e Endpoint) fromDb(db db.Endpoint) Endpoint {
	e.URL = db.URL
	e.EndpointType = db.EndpointType
	e.Identifier = Identifier(db.Identifier)
	e.Status = db.Status
	e.Version = db.Version
	return e
}

func (o Organization) fromDb(db db.Organization) Organization {
	e := endpointsArrayFromDb(db.Endpoints)
	o.Identifier = Identifier(db.Identifier)
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
	org := db.Organization{
		Identifier: db.Identifier(o.Identifier),
		Name:       o.Name,
		PublicKey:  o.PublicKey,
	}

	if o.Keys != nil {
		ks := *o.Keys
		em := make([]interface{}, len(ks))
		for i, k := range ks {
			em[i] = k.AdditionalProperties
		}
		org.Keys = em
	}

	if o.Endpoints != nil {
		org.Endpoints = endpointsArrayToDb(*o.Endpoints)
	}

	return org
}

func (a Endpoint) toDb() db.Endpoint {
	return db.Endpoint{
		URL:          a.URL,
		EndpointType: a.EndpointType,
		Identifier:   db.Identifier(a.Identifier),
		Status:       a.Status,
		Version:      a.Version,
	}
}

func organizationsArrayFromDb(organizationsIn []db.Organization) []Organization {
	os := make([]Organization, len(organizationsIn))
	for i, a := range organizationsIn {
		os[i] = Organization{}.fromDb(a)
	}
	return os
}

func endpointsArrayFromDb(endpointsIn []db.Endpoint) []Endpoint {
	es := make([]Endpoint, len(endpointsIn))
	for i, a := range endpointsIn {
		es[i] = Endpoint{}.fromDb(a)
	}
	return es
}

func organizationsToFromDb(organizationsIn []Organization) []db.Organization {
	os := make([]db.Organization, len(organizationsIn))
	for i, a := range organizationsIn {
		os[i] = a.toDb()
	}
	return os
}

func endpointsArrayToDb(endpointsIn []Endpoint) []db.Endpoint {
	es := make([]db.Endpoint, len(endpointsIn))
	for i, a := range endpointsIn {
		es[i] = a.toDb()
	}
	return es
}
