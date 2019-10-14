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

package db

import (
	"errors"
	"testing"
)

func TestValidateLocation(t *testing.T) {
	t.Run("Missing files gives err", func(t *testing.T) {
		err := validateLocation("../../test_data/missing_files")

		if err == nil {
			t.Errorf("Expected error: %v", ErrMissingRequiredFiles)
		} else if !errors.Is(err, ErrMissingRequiredFiles) {
			t.Errorf("Expected error: [%v], got [%v]", ErrMissingRequiredFiles, err)
		}
	})

	t.Run("All files present", func(t *testing.T) {
		err := validateLocation("../../test_data/all_empty_files")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("All files present with trailing slash", func(t *testing.T) {
		err := validateLocation("../../test_data/all_empty_files/")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestReadFile(t *testing.T) {
	data, err := ReadFile("../../test_data/all_empty_files", "endpoints.json")

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
