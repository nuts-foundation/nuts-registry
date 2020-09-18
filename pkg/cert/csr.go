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
	"crypto/x509/pkix"
	"errors"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
)

// VendorCertificateRequest creates a CertificateRequest template for issuing a vendor certificate.
//   vendorID:      URN-OID-encoded ID of the vendor
//   vendorName:    Name of the vendor
//   qualifier:     (optional) Qualifier for the certificate, which will be postfixed to Subject.CommonName
//   domain:        Domain the vendor operates in, e.g. "healthcare"
//   env:           (optional) Environment for the certificate, e.g. "Test" or "Dev", which will be postfixed to Subject.CommonName
func VendorCertificateRequest(vendorID core.PartyID, vendorName string, qualifier string, domain string) (x509.CertificateRequest, error) {
	if vendorID.IsZero() {
		return x509.CertificateRequest{}, errors.New("missing vendor identifier")
	}
	if vendorName == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor name")
	}
	if domain == "" {
		return x509.CertificateRequest{}, errors.New("missing domain")
	}
	subjectAltName, err := cert.MarshalOtherSubjectAltName(cert.OIDNutsVendor, vendorID.Value())
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: cert.OIDSubjectAltName, Critical: false, Value: subjectAltName},
	}

	domainData, err := cert.MarshalNutsDomain(domain)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions = append(extensions, pkix.Extension{Id: cert.OIDNutsDomain, Critical: false, Value: domainData})

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
func OrganisationCertificateRequest(vendorName string, organisationID core.PartyID, organisationName string, domain string) (x509.CertificateRequest, error) {
	if vendorName == "" {
		return x509.CertificateRequest{}, errors.New("missing vendor name")
	}
	if organisationID.IsZero() {
		return x509.CertificateRequest{}, errors.New("missing organization identifier")
	}
	if organisationName == "" {
		return x509.CertificateRequest{}, errors.New("missing organization name")
	}
	subjectAltName, err := cert.MarshalOtherSubjectAltName(OIDAGBCode, organisationID.Value())
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: cert.OIDSubjectAltName, Critical: false, Value: subjectAltName},
	}
	if domain != "" {
		domainData, err := cert.MarshalNutsDomain(domain)
		if err != nil {
			return x509.CertificateRequest{}, err
		}
		extensions = append(extensions, pkix.Extension{Id: cert.OIDNutsDomain, Critical: false, Value: domainData})
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

