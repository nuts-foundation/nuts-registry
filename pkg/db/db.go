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

package db

// StatusActive represents the "active" status
const StatusActive = "active"

// Actor defines component schema for Actor.
type Actor struct {
	Identifier Identifier `json:"identifier"`
}

// Endpoint defines component schema for Endpoint.
type Endpoint struct {
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// EndpointOrganization defines component schema for EndpointOrganization.
type EndpointOrganization struct {
	Endpoint     Identifier `json:"endpoint"`
	Organization Identifier `json:"organization"`
	Status       string     `json:"status"`
}

// Identifier defines component schema for Identifier.
type Identifier string

// String converts an identifier to string
func (i Identifier) String() string {
	return string(i)
}

// Organization defines component schema for Organization.
type Organization struct {
	Actors     []Actor    `json:"actors,omitempty"`
	Identifier Identifier `json:"identifier"`
	Name       string     `json:"name"`
	PublicKey  *string    `json:"publicKey,omitempty"`
	Endpoints  []Endpoint
}

// todo: Db temporary abstraction
type Db interface {
	FindEndpointsByOrganization(organizationIdentifier string) ([]Endpoint, error)
	Load(location string) error
	SearchOrganizations(query string) []Organization
	OrganizationById(id string) (*Organization, error)
	RemoveOrganization(id string) error
	RegisterOrganization(org Organization) error
}

type dbError struct {
	s string
}

func newDbError(text string) error {
	return &dbError{text}
}

func (e *dbError) Error() string {
	return e.s
}
