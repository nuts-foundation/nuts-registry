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

package domain

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
)

// RegisterDataClientCertificate event type
const RegisterDataClientCertificate events.EventType = "RegisterDataClientCertificate"

// RegisterDataClientCertificateEvent event holds the certificate(s) which is to be used as client certificate in data connections
// the X509 certificates are encoded as a JWKSet in the Keys property
type RegisterDataClientCertificateEvent struct {
	VendorID core.PartyID  `json:"vendorIdentifier"`
	Keys     []interface{} `json:"keys,omitempty"`
}
