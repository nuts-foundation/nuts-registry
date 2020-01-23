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
}

func TestReadFile(t *testing.T) {
	data, err := readFile("../../test_data/all_empty_files", "endpoints.json")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expected := "[]"
	got := string(data)
	if got != expected {
		t.Errorf("Expected empty json file with: [[]], got [%s]", got)
	}
}
