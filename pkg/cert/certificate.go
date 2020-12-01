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
package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	"strings"
)

var OIDAGBCode = asn1.ObjectIdentifier{2, 16, 840, 1, 113883, 2, 4, 6, 1}

// NutsCertificate is a type alias for a regular x509.Certificate. It provides convenience methods for retrieving Nuts
// specific properties from the certificate.
type NutsCertificate x509.Certificate

// Efficiently converts the *x509.Certificate as *NutsCertificate
func NewNutsCertificate(certificate *x509.Certificate) *NutsCertificate {
	// Certificates are generally passed as pointer to avoid copying their (rather large ASN.1) in-memory representation
	// Thus this function uses casting instead of dereferencing.
	return (*NutsCertificate)(certificate)
}

// GetVendorID tries to get the vendor ID from the certificate SAN.
// If it doesn't exist, an empty identifier is returned without error.
// If an error occurs, an empty identifier is returned alongside the error.
func (c NutsCertificate) GetVendorID() (core.PartyID, error) {
	return c.getPartyIDFromSAN(cert.OIDNutsVendor)
}

// GetOrganizationID tries to get the organization ID (AGB-code) from the certificate SAN.
// If it doesn't exist, an empty identifier is returned without error.
// If an error occurs, an empty identifier is returned alongside the error.
func (c NutsCertificate) GetOrganizationID() (core.PartyID, error) {
	return c.getPartyIDFromSAN(OIDAGBCode)
}

// GetDomain tries to get the Nuts Domain from the certificate. If the certificate doesn't require the extension,
// an empty string is returned. If something else goes wrong, the error is returned.
func (c NutsCertificate) GetDomain() (string, error) {
	for _, extension := range c.Extensions {
		if extension.Id.Equal(cert.OIDNutsDomain) {
			return cert.UnmarshalNutsDomain(extension.Value)
		}
	}
	return "", nil
}

func (c NutsCertificate) getPartyIDFromSAN(oid asn1.ObjectIdentifier) (core.PartyID, error) {
	for _, extension := range c.Extensions {
		if extension.Id.Equal(cert.OIDSubjectAltName) {
			var rawValue asn1.RawValue
			if _, err := asn1.Unmarshal(extension.Value, &rawValue); err != nil {
				return core.PartyID{}, err
			}
			if value, _ := cert.UnmarshalOtherSubjectAltName(oid, extension.Value); value != "" {
				// Vendor/Organization ID values in issued up until v0.14 are fully qualified, which was incorrect since
				// the type (OID) was already specified in the ASN.1 structure (Type: Value). This is fixed starting 0.15.
				// See https://github.com/nuts-foundation/nuts-registry/issues/142
				if strings.HasPrefix(value, fmt.Sprintf("urn:oid:%s:", oid.String())) {
					// Behaviour for <= v0.14 certificates
					return core.ParsePartyID(value)
				} else {
					// Behaviour for > v0.14 certificates
					return core.NewPartyID(oid.String(), value)
				}
			}
		}
	}
	return core.PartyID{}, nil
}

// MarshalJSON marshals the ASN.1 representation of the certificate as base64 JSON string.
func (c NutsCertificate) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(c.Raw))
}

// UnmarshalJSON unmarshals a JSON string containing a base64 encoded ASN.1 certificate into a x509.Certificate.
func (c *NutsCertificate) UnmarshalJSON(bytes []byte) error {
	var str string
	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}
	asn1Bytes, err := base64.StdEncoding.DecodeString(str);
	if err != nil {
		return err
	}
	certificate, err := x509.ParseCertificate(asn1Bytes)
	if err != nil {
		return err
	}
	*c = *NewNutsCertificate(certificate)
	return nil
}