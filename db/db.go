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

import "github.com/nuts-foundation/nuts-registry/generated"

type Db interface {
	FindEndpointsByOrganisation(organizationIdentifier string) ([]generated.Endpoint, error)
	Load() error
	SearchOrganizations(query string) []generated.Organization
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
