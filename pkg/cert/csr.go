package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	asn1util "github.com/nuts-foundation/nuts-crypto/pkg/asn1"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
)

var oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var oidNuts = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 54851}
var oidNutsVendor = asn1util.OIDAppend(oidNuts, 4)
var oidNutsDomain = asn1util.OIDAppend(oidNuts, 3)
var oidAgbCode = asn1.ObjectIdentifier{2, 16, 840, 1, 113883, 2, 4, 6, 1}

// JwkCertificateType holds the JSON Web Key member name which will hold CertificateType, describing the type of the certificate.
const JwkCertificateType = "ct"

// CertificateType holds one of the certificate types as specified in the Nuts certificate specification
type CertificateType string

const (
	// VendorCACertificate specifies the CA certificate of a vendor
	VendorCACertificate CertificateType = "vendor-ca"
	// OrganisationCertificate specifies the certificate of an organisation, issued by a vendor
	OrganisationCertificate CertificateType = "org"
)

// VendorCACertificateRequest creates a CertificateRequest template for issuing a vendor CA certificate.
// Parameters 'domain' and 'env' are optional.
func VendorCACertificateRequest(vendorID string, vendorName string, domain string, env string) (x509.CertificateRequest, error) {
	if vendorID == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor identifier")
	}
	if vendorName == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor name")
	}
	subjectAltName, err := cert.MarshalOtherSubjectAltName(oidNutsVendor, vendorID)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: oidSubjectAltName, Critical: false, Value: subjectAltName},
	}
	if domain != "" {
		domainData, err := cert.MarshalNutsDomain(domain)
		if err != nil {
			return x509.CertificateRequest{}, err
		}
		extensions = append(extensions, pkix.Extension{Id: oidNutsDomain, Critical: false, Value: domainData})
	}
	commonName := vendorName + " CA"
	if env != "" {
		commonName += " " + env
	}
	return x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"NL"},
			Organization: []string{vendorName},
			CommonName:   commonName,
		},
		ExtraExtensions: extensions,
	}, nil
}

// OrganisationCertificateRequest creates a CertificateRequest template for issuing an organisation. The certificate
// should be issued by the vendor CA. Parameters 'domain' and 'env' are optional.
func OrganisationCertificateRequest(vendorName string, organisationID string, organisationName string, domain string, env string) (x509.CertificateRequest, error) {
	if vendorName == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor name")
	}
	if organisationID == "" {
		return x509.CertificateRequest{}, errors.New("missing organization identifier")
	}
	if organisationName == "" {
		return x509.CertificateRequest{}, errors.New("missing organization name")
	}
	subjectAltName, err := cert.MarshalOtherSubjectAltName(oidAgbCode, organisationID)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: oidSubjectAltName, Critical: false, Value: subjectAltName},
	}
	if domain != "" {
		domainData, err := cert.MarshalNutsDomain(domain)
		if err != nil {
			return x509.CertificateRequest{}, err
		}
		extensions = append(extensions, pkix.Extension{Id: oidNutsDomain, Critical: false, Value: domainData})
	}
	commonName := organisationName
	if env != "" {
		commonName += " " + env
	}
	return x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"NL"},
			Organization: []string{vendorName},
			CommonName:   commonName,
		},
		ExtraExtensions: extensions,
	}, nil
}
