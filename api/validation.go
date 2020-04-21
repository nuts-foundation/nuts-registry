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

import "errors"

func (o Organization) validate() error {
	if err := nonEmptyString(o.Identifier.String(), "identifier"); err != nil {
		return err
	}
	if err := nonEmptyString(o.Name, "name"); err != nil {
		return err
	}
	return nil
}

func (v Vendor) validate() error {
	if err := nonEmptyString(v.Name, "name"); err != nil {
		return err
	}
	return nil
}

func (e Endpoint) validate() error {
	if err := nonEmptyString(e.URL, "url"); err != nil {
		return err
	}
	if err := nonEmptyString(e.EndpointType, "endpoint type"); err != nil {
		return err
	}
	return nil
}

func nonEmptyString(value string, name string) error {
	if len(value) == 0 {
		return errors.New("missing " + name)
	}
	return nil
}
