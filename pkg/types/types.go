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
package types

type EndpointID string

const (
	// HealthcareDomain is a const for domain 'healthcare'
	HealthcareDomain string = "healthcare"
	// PersonalDomain is a const for domain 'personal' (which are "PGO's")
	PersonalDomain = "personal"
	// InsuranceDomain is a const for domain 'insurance'
	InsuranceDomain = "insurance"
	// FallbackDomain is a const for the fallback domain in case there's no domain set, which can be the case for legacy data.
	FallbackDomain = HealthcareDomain
)

