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
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitialize(t *testing.T) {
	t.Run("server mode", func(t *testing.T) {
		instance := pkg.RegistryInstance()
		instance.Config.Mode = core.ServerEngineMode
		assert.IsType(t, &pkg.Registry{}, initialize(instance))
	})
	t.Run("client mode", func(t *testing.T) {
		instance := pkg.RegistryInstance()
		instance.Config.Mode = core.ClientEngineMode
		assert.IsType(t, api.HttpClient{}, initialize(instance))
	})
}

func TestNewRegistryClient(t *testing.T) {
	client := NewRegistryClient()
	assert.NotNil(t, client)
}