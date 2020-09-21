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
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	errors2 "github.com/pkg/errors"
)

// RegisterVendor event type
const RegisterVendor events.EventType = "RegisterVendorEvent"

// RegisterVendorEvent event
type RegisterVendorEvent struct {
	Identifier core.PartyID  `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

func (r *RegisterVendorEvent) PostProcessUnmarshal(event events.Event) error {
	// Default fallback to 'healthcare' domain when none is set, for handling legacy data when 'domain' didn't exist.
	if r.Domain == "" {
		r.Domain = types.FallbackDomain
	}
	if err := cert.ValidateJWK(r.Keys...); err != nil {
		return err
	}
	for _, key := range r.Keys {
		var vendorID core.PartyID
		if certificate := cert.GetCertificate(key); certificate != nil {
			var err error
			if vendorID, err = cert2.NewNutsCertificate(certificate).GetVendorID(); err != nil {
				return errors2.Wrap(err, "unable to unmarshal vendor ID from certificate")
			}
		}
		if r.Identifier != vendorID {
			return fmt.Errorf("vendor ID in certificate (%s) doesn't match event (%s)", vendorID, r.Identifier)
		}
	}
	return nil
}

// VendorEventMatcher returns an EventMatcher which matches the RegisterVendorEvent for the vendor with the specified ID.
func VendorEventMatcher(vendorID core.PartyID) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != RegisterVendor {
			return false
		}
		var p = RegisterVendorEvent{}
		_ = event.Unmarshal(&p)
		return vendorID == p.Identifier
	}
}
