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

package client

import (
	"time"

	"github.com/nuts-foundation/nuts-registry/pkg"
)

// NewRegistryClient creates a new Local- or RemoteClient for the nuts registry
func NewRegistryClient() pkg.RegistryClient {
	return initialize(pkg.RegistryInstance())
}

func initialize(registry *pkg.Registry) pkg.RegistryClient {
	return HttpClient{
		ServerAddress: registry.Config.Address,
		Timeout:       time.Duration(registry.Config.ClientTimeout) * time.Second,
	}
}
