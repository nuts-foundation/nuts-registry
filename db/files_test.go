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

import "testing"

func TestIncorrectDir(t *testing.T) {
	err := ValidateLocation("./does_not_exist")

	expected := "open ./does_not_exist: no such file or directory"
	if err == nil {
		t.Errorf("Expected error with: %s", expected)
	} else if err.Error() != expected {
		t.Errorf("Expected error with: [%s], got [%s]", expected, err.Error())
	}
}

func TestEmptyDir(t *testing.T) {
	err := ValidateLocation("../test_data/empty")

	expected := "../test_data/empty is empty"
	if err == nil {
		t.Errorf("Expected error with: %s", expected)
	} else if err.Error() != expected {
		t.Errorf("Expected error with: [%s], got [%s]", expected, err.Error())
	}
}

func TestMissingFiles(t *testing.T) {
	err := ValidateLocation("../test_data/missing_files")

	expected := "../test_data/missing_files is missing required files: endpoints.json, endpoints_organizations.json"
	if err == nil {
		t.Errorf("Expected error with: %s", expected)
	} else if err.Error() != expected {
		t.Errorf("Expected error with: [%s], got [%s]", expected, err.Error())
	}
}

func TestAllFilesPresent(t *testing.T) {
	err := ValidateLocation("../test_data/all_empty_files")

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
}

func TestAllFilesPresentTrailingSlash(t *testing.T) {
	err := ValidateLocation("../test_data/all_empty_files/")

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
}

func TestReadFile(t *testing.T) {
	data, err := ReadFile("../test_data/all_empty_files", "endpoints.json")

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	expected := "[]"
	got := string(data)
	if got != expected {
		t.Errorf("Expected empty json file with: [[]], got [%s]", got)
	}
}
