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

package events

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestValidateLocation(t *testing.T) {
	t.Run("Location is a file", func(t *testing.T) {
		err := validateLocation("../../test_data/invalid_location")
		assert.Error(t, err)
	})

	t.Run("All files present with trailing slash", func(t *testing.T) {
		err := validateLocation("../../test_data/all_empty_files/")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Non-existing location is created", func(t *testing.T) {
		const target = "../../test_data/non-existing/"
		err := validateLocation(target)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		err = os.Remove(target)
		if !assert.NoError(t, err, "unable to remove test directory") {
			return
		}
	})
}

func TestNormalizeLocation(t *testing.T) {
	assert.Equal(t, "foo/bar/file", normalizeLocation("foo/bar/", "file"))
}
