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
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sort"
)

type fileError struct {
	s string
}

func newFileError(text string) error {
	return &fileError{text}
}

func (e *fileError) Error() string {
	return e.s
}

// Validate location of data files, checks if following files can be found in given directory:
// - endpoints.json
// - organisations.json
// - endpoints_organisations.json
func validateLocation(location string) error {
	sLocation := sanitizeLocation(location)

	if _, err := os.Stat(location); os.IsNotExist(err) {
		// create and return
		os.Mkdir(location, os.ModePerm)
		return newFileError(fmt.Sprintf("%s is missing required files: ", sLocation))
	}

	files, err := ioutil.ReadDir(sLocation)

	if err != nil {
		return err
	}

	m := make(map[string]bool)

	for _, f := range requiredFiles {
		m[f] = false
	}

	for _, f := range files {
		_, prs := m[f.Name()]
		if prs {
			delete(m, f.Name())
		}
	}

	if len(m) != 0 {
		var keys []string
		for key := range m {
			keys = append(keys, key)
		}

		// sort to get consistent test results
		sort.Strings(keys)

		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "%s is missing required files: ", sLocation)
		for i, k := range keys {
			if i == len(keys)-1 {
				fmt.Fprintf(buf, "%s", k)
			} else {
				fmt.Fprintf(buf, "%s, ", k)
			}
		}

		return newFileError(buf.String())
	}

	return nil
}

// Readfile reads a file relative to datadir
func ReadFile(location string, file string) ([]byte, error) {
	finalLocation := fmt.Sprintf("%s/%s", sanitizeLocation(location), file)
	logrus.Debugf("Reading file from %s", finalLocation)

	return ioutil.ReadFile(finalLocation)
}

func sanitizeLocation(dirty string) string {
	iLast := len(dirty) - 1
	if dirty[iLast:] == "/" {
		return dirty[:iLast]
	}
	return dirty
}
