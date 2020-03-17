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
	"crypto/x509"
	"github.com/lestrrat-go/jwx/jwk"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"time"
)

// StatusActive represents the "active" status
const StatusActive = "active"

// Endpoint defines component schema for Endpoint.
type Endpoint struct {
	URL          string            `json:"URL"`
	EndpointType string            `json:"endpointType"`
	Identifier   Identifier        `json:"identifier"`
	Status       string            `json:"status"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// Identifier defines component schema for Identifier.
type Identifier string

// String converts an identifier to string
func (i Identifier) String() string {
	return string(i)
}

// Organization defines component schema for Organization.
type Organization struct {
	Identifier Identifier    `json:"identifier"`
	Name       string        `json:"name"`
	PublicKey  *string       `json:"publicKey,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
	Endpoints  []Endpoint
}

func (o Organization) GetActiveCertificates() []*x509.Certificate {
	return cert.GetActiveCertificates(o.Keys, time.Now())
}

// Vendor defines component schema for Vendor.
type Vendor struct {
	Identifier Identifier    `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

// GetActiveCertificates looks up the vendor's certificates and returns them sorted, longest valid certificate first.
// Expired certificates aren't returned.
func (v Vendor) GetActiveCertificates() []*x509.Certificate {
	return cert.GetActiveCertificates(v.Keys, time.Now())
}

// KeysAsSet transforms the raw map in Keys to a jwk.Set. If no keys are present, it'll return an empty set
func (o Organization) KeysAsSet() (*jwk.Set, error) {
	var maps []map[string]interface{}
	for _, key := range o.Keys {
		maps = append(maps, key.(map[string]interface{}))
	}
	return crypto.MapsToJwkSet(maps)
}

// CurrentPublicKey returns the first current active public key. If a JWK set is registered, it'll search in the keys there.
// If none are valid it'll return an error.
// If no JWK Set is set, it'll always return the (deprecated) PublicKey
// TODO: In a later stage the certificate capabilities of the JWK will be used. For now the first JWK is returned
func (o Organization) CurrentPublicKey() (jwk.Key, error) {
	if len(o.Keys) > 0 {
		set, err := o.KeysAsSet()
		if err != nil {
			return nil, err
		}
		key := set.Keys[0]

		// check for certificate validity at a later stage.
		return key, nil
	}

	key, err := crypto.PemToPublicKey([]byte(*o.PublicKey))
	if err != nil {
		return nil, err
	}

	return jwk.New(key)
}

type Db interface {
	RegisterEventHandlers(system events.EventSystem)
	FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]Endpoint, error)
	SearchOrganizations(query string) []Organization
	OrganizationById(id string) (*Organization, error)
	VendorByID(id string) *Vendor
	OrganizationsByVendorID(id string) []*Organization
	ReverseLookup(name string) (*Organization, error)
}
