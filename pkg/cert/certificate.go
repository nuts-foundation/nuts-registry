package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
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
			if value, err := cert.UnmarshalOtherSubjectAltName(oid, extension.Value); err != nil {
				return core.PartyID{}, err
			} else {
				return core.NewPartyID(oid.String(), value)
			}
		}
	}
	return core.PartyID{}, nil
}
