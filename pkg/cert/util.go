package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
)

// GetVendorSubjectAltName tries to get the vendor ID from the certificate SAN.
// If it doesn't exist, an empty string is returned without error.
// If an error occurs, an empty string is returned alongside the error.
func GetVendorSubjectAltName(certificate *x509.Certificate) (string, error) {
	return getOtherSubjectAltName(certificate, oidNutsVendor)
}

// GetOrganizationSubjectAltName tries to get the organization ID (AGB-code) from the certificate SAN.
// If it doesn't exist, an empty string is returned without error.
// If an error occurs, an empty string is returned alongside the error.
func GetOrganizationSubjectAltName(certificate *x509.Certificate) (string, error) {
	return getOtherSubjectAltName(certificate, oidAgbCode)
}

// GetDomain tries to get the Nuts Domain from the certificate. If the certificate doesn't require the extension,
// an empty string is returned. If something else goes wrong, the error is returned.
func GetDomain(certificate *x509.Certificate) (string, error) {
	for _, extension := range certificate.Extensions {
		if extension.Id.Equal(oidNutsDomain) {
			return cert.UnmarshalNutsDomain(extension.Value)
		}
	}
	return "", nil
}

func getOtherSubjectAltName(certificate *x509.Certificate, oid asn1.ObjectIdentifier) (string, error) {
	if certificate == nil {
		return "", errors.New("certificate is nil")
	}
	for _, extension := range certificate.Extensions {
		if extension.Id.Equal(oidSubjectAltName) {
			return cert.UnmarshalOtherSubjectAltName(oid, extension.Value)
		}
	}
	return "", nil
}