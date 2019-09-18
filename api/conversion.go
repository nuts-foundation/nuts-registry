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

func (a Actor) fromDb(db db.Actor) Actor {
	a.Identifier = Identifier(db.Identifier)
	return a
}

func (o Organization) fromDb(db db.Organization) Organization {
	o.Actors = actorsArrayFromDb(db.Actors)
	o.Identifier = Identifier(db.Identifier)
	o.Name = db.Name
	o.PublicKey = db.PublicKey
	o.Endpoints = endpointsArrayFromDb(db.Endpoints)
	return o
}

func (eo EndpointOrganization) fromDb(db db.EndpointOrganization) EndpointOrganization {
	eo.Endpoint = Identifier(db.Endpoint)
	eo.Organization = Identifier(db.Organization)
	eo.Status = db.Status
	return eo
}

func (a Actor) toDb() db.Actor {
	return db.Actor{
		Identifier: db.Identifier(a.Identifier),
	}
}

func (a Organization) toDb() db.Organization {
	return db.Organization{
		Actors:     actorsArrayToDb(a.Actors),
		Identifier: db.Identifier(a.Identifier),
		Name:       a.Name,
		PublicKey:  a.PublicKey,
		Endpoints:  endpointsArrayToDb(a.Endpoints),
	}
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

func actorsArrayFromDb(actorsIn []db.Actor) []Actor {
	as := make([]Actor, len(actorsIn))
	for i, a := range actorsIn {
		as[i] = Actor{}.fromDb(a)
	}
	return as
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

func actorsArrayToDb(actorsIn []Actor) []db.Actor {
	as := make([]db.Actor, len(actorsIn))
	for i, a := range actorsIn {
		as[i] = a.toDb()
	}
	return as
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
