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

package events

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validate location of data files. Creates the directory if it doesn't exist.
func validateLocation(location string) error {
	sLocation := sanitizeLocation(location)
	info, err := os.Stat(sLocation)
	if err != nil {
		if os.IsNotExist(err) {
			// create and return
			return os.MkdirAll(location, os.ModePerm)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("datadir is file, expected a directory (location = %s)", location)
	}
	return nil
}

func normalizeLocation(location string, file string) string {
	return filepath.Join(sanitizeLocation(location), file)
}

func sanitizeLocation(dirty string) string {
	iLast := len(dirty) - 1
	if dirty[iLast:] == string(os.PathSeparator) {
		return dirty[:iLast]
	}
	return dirty
}
