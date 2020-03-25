package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
)

// VendorCertificateRequest creates a CertificateRequest template for issuing a vendor certificate.
//   vendorID:      URN-OID-encoded ID of the vendor
//   vendorName:    Name of the vendor
//   qualifier:     (optional) Qualifier for the certificate, which will be postfixed to Subject.CommonName
//   domain:        Domain the vendor operates in, e.g. "healthcare"
//   env:           (optional) Environment for the certificate, e.g. "Test" or "Dev", which will be postfixed to Subject.CommonName
func VendorCertificateRequest(vendorID string, vendorName string, qualifier string, domain string) (x509.CertificateRequest, error) {
	if vendorID == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor identifier")
	}
	if vendorName == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor name")
	}
	if domain == "" {
		return x509.CertificateRequest{}, errors.New("missing domain")
	}
	subjectAltName, err := cert.MarshalOtherSubjectAltName(oidNutsVendor, vendorID)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: oidSubjectAltName, Critical: false, Value: subjectAltName},
	}

	domainData, err := cert.MarshalNutsDomain(domain)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions = append(extensions, pkix.Extension{Id: oidNutsDomain, Critical: false, Value: domainData})

	commonName := vendorName
	if qualifier != "" {
		commonName += " " + qualifier
	}
	return x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"NL"},
			Organization: []string{vendorName},
			CommonName:   commonName,
		},
		// Depending on how the resulting CSR object is used (directly or marshalled to ASN.1 and then unmarshalled)
		// we need to either set Extensions or ExtraExtensions, so we'll set them both.
		ExtraExtensions: extensions,
		Extensions: extensions,
	}, nil
}

// OrganisationCertificateRequest creates a CertificateRequest template for issuing an organisation. The certificate
// should be issued by the vendor CA. Parameters 'domain' and 'env' are optional.
func OrganisationCertificateRequest(vendorName string, organisationID string, organisationName string, domain string) (x509.CertificateRequest, error) {
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
	return x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"NL"},
			Organization: []string{vendorName},
			CommonName:   commonName,
		},
		// Depending on how the resulting CSR object is used (directly or marshalled to ASN.1 and then unmarshalled)
		// we need to either set Extensions or ExtraExtensions, so we'll set them both.
		ExtraExtensions: extensions,
		Extensions: extensions,
	}, nil
}

