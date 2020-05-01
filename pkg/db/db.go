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
	crypto2 "crypto"
	"crypto/x509"
	"errors"
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
	Organization Identifier        `json:"organization"`
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
	// Deprecated: use Keys or helper functions to retrieve the current key in use by the organization
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
	result, err := crypto.MapsToJwkSet(maps)
	if err != nil {
		return nil, err
	}
	// Support deprecated PublicKey
	if o.PublicKey != nil {
		key, err := crypto.PemToPublicKey([]byte(*o.PublicKey))
		if err != nil {
			return nil, err
		}
		pubKey, _ := jwk.New(key)
		result.Keys = append(result.Keys, pubKey)
	}
	return result, nil
}

// HasKey checks whether the given key is owned by the organization
func (o Organization) HasKey(key jwk.Key, validAtMoment time.Time) (bool, error) {
	// func can't return error
	keyTp, _ := key.Thumbprint(crypto2.SHA256)
	keys, err := o.KeysAsSet()
	if err != nil {
		return false, err
	}
	for _, k := range keys.Keys {
		// func can't return error
		tp, _ := k.Thumbprint(crypto2.SHA256)
		if bytes.Compare(keyTp, tp) == 0 {
			// Found the key
			chainInterf, chainExists := k.Get("x5c")
			if chainExists {
				certificate := chainInterf.([]*x509.Certificate)[0]
				// JWK has a certificate attached, check if it's valid at the specified time
				if validAtMoment.Before(certificate.NotBefore) || validAtMoment.After(certificate.NotAfter) {
					return false, nil
				}
			}
			return true, nil
		}
	}
	return false, nil
}

// CurrentPublicKey returns the public key associated with the organization certificate which has the longest validity.
// For backwards compatibility:
//  1. If the organization has no certificates, it will return the first JWK.
//  2. If the organization has no JWKs, it will return the (deprecated) PublicKey.
// If none of the above conditions are matched, an error is returned.
func (o Organization) CurrentPublicKey() (jwk.Key, error) {
	var hasCerts = false
	for _, jwkAsMap := range o.Keys {
		if (jwkAsMap.(map[string]interface{}))[jwk.X509CertChainKey] != nil {
			hasCerts = true
			break
		}
	}
	if hasCerts {
		// Organization has certificates, use those and ignore the rest
		certs := o.GetActiveCertificates()
		if len(certs) > 0 {
			return crypto.CertificateToJWK(certs[0])
		}
		return nil, errors.New("organization has no active certificates")
	} else {
		// Organization has no certificates, fallback to plain JWKs
		set, err := o.KeysAsSet()
		if err != nil {
			return nil, err
		}
		key := set.Keys[0]
		return key, nil
	}
}

type Db interface {
	RegisterEventHandlers(fn events.EventRegistrar)
	FindEndpointsByOrganizationAndType(organizationID string, endpointType *string) ([]Endpoint, error)
	SearchOrganizations(query string) []Organization
	OrganizationById(id string) (*Organization, error)
	VendorByID(id string) *Vendor
	OrganizationsByVendorID(id string) []*Organization
	ReverseLookup(name string) (*Organization, error)
}
