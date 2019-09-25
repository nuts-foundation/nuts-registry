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

package client

import (
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/sirupsen/logrus"
	"time"
)

// NewRegistryClient creates a new Local- or RemoteClient for the nuts registry
func NewRegistryClient() pkg.RegistryClient {
	registry := pkg.RegistryInstance()

	if registry.Config.Mode == "server" {
		if err := registry.Configure(); err != nil {
			logrus.Panic(err)
		}

		return registry
	} else {
		return api.HttpClient{
			ServerAddress: registry.Config.Address,
			Timeout:       time.Second,
		}
	}
}
