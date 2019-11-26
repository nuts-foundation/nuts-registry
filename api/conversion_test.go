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

import (
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrganizationConversion(t *testing.T) {
	t.Run("JWK is converted correctly from DB", func(t *testing.T) {
		em := make([]interface{}, 1)
		em[0] = map[string]interface{}{"kty": "EC"}

		o := Organization{}.fromDb(db.Organization{Keys: em})

		assert.Len(t, *o.Keys, 1)
		assert.Equal(t, "EC", (*o.Keys)[0].AdditionalProperties["kty"].(string))
	})

	t.Run("JWK is converted correctly to DB", func(t *testing.T) {
		em := []JWK{{AdditionalProperties: map[string]interface{}{"kty": "EC"}}}
		o := Organization{Keys: &em}.toDb()

		assert.Len(t, o.Keys, 1)
		assert.Equal(t, "EC", o.Keys[0].(map[string]interface{})["kty"].(string))
	})
}
