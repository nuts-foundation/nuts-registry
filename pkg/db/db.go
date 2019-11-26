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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/lestrrat-go/jwx/jwk"
)

// StatusActive represents the "active" status
const StatusActive = "active"

// Endpoint defines component schema for Endpoint.
type Endpoint struct {
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// EndpointOrganization defines component schema for EndpointOrganization.
type EndpointOrganization struct {
	Endpoint     Identifier `json:"endpoint"`
	Organization Identifier `json:"organization"`
	Status       string     `json:"status"`
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

// KeysAsSet transforms the raw map in Keys to a jwk.Set. If no keys are present, it'll return an empty set
func (o Organization) KeysAsSet() (*jwk.Set, error) {
	if len(o.Keys) == 0 {
		return &jwk.Set{}, nil
	}

	var set jwk.Set
	m := make(map[string]interface{})
	m["keys"] = o.Keys
	if err := set.ExtractMap(m); err != nil {
		return nil, err
	}
	return &set, nil
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

	key, err := pemToPublicKey([]byte(*o.PublicKey))
	if err != nil {
		return nil, err
	}

	return jwk.New(key)
}

// temporary func for converting pem public keys to rsaPublicKey
func pemToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key, key is of the wrong type")
	}

	b := block.Bytes
	key, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, err
	}
	finalKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Unable to convert public key to RSA public key")
	}

	return finalKey, nil
}

// todo: Db temporary abstraction
type Db interface {
	FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]Endpoint, error)
	Load(location string) error
	SearchOrganizations(query string) []Organization
	OrganizationById(id string) (*Organization, error)
	RemoveOrganization(id string) error
	RegisterOrganization(org Organization) error
	ReverseLookup(name string) (*Organization, error)
}
