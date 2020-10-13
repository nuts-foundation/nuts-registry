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
	"testing"

	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationConversion(t *testing.T) {
	t.Run("JWK is converted correctly from DB", func(t *testing.T) {
		em := make([]interface{}, 1)
		em[0] = map[string]interface{}{"kty": "EC"}

		o := Organization{}.fromDb(db.Organization{Keys: em})

		assert.Len(t, *o.Keys, 1)
		assert.Equal(t, "EC", (*o.Keys)[0]["kty"].(string))
	})

	t.Run("JWK is converted correctly to DB", func(t *testing.T) {
		em := []JWK{{"kty": "EC"}}
		o := Organization{Keys: &em}.toDb()

		assert.Len(t, o.Keys, 1)
		assert.Equal(t, "EC", o.Keys[0].(JWK)["kty"].(string))
	})
}

func TestVendorConversion(t *testing.T) {
	t.Run("JWK is converted correctly from DB", func(t *testing.T) {
		em := make([]interface{}, 1)
		em[0] = map[string]interface{}{"kty": "EC"}

		o := Vendor{}.fromDb(db.Vendor{Keys: em})

		assert.Len(t, *o.Keys, 1)
		assert.Equal(t, "EC", (*o.Keys)[0]["kty"].(string))
	})

	t.Run("JWK is converted correctly to DB", func(t *testing.T) {
		em := []JWK{{"kty": "EC"}}
		o := Vendor{Keys: &em}.toDb()

		assert.Len(t, o.Keys, 1)
		assert.Equal(t, "EC", o.Keys[0].(JWK)["kty"].(string))
	})
}
