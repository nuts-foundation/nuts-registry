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
package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg"
	pkg2 "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/sirupsen/logrus"
	"path"
)

func NewTestRegistryInstance(testDirectory string) *Registry {
	config := TestRegistryConfig(testDirectory)
	newInstance := NewRegistryInstance(config, pkg.NewTestCryptoInstance(testDirectory), pkg2.NewTestNetworkInstance(testDirectory))
	if err := newInstance.Configure(); err != nil {
		logrus.Fatal(err)
	}
	instance = newInstance
	return newInstance
}

func TestRegistryConfig(testDirectory string) RegistryConfig {
	config := DefaultRegistryConfig()
	config.Datadir = path.Join(testDirectory, "registry")
	return config
}