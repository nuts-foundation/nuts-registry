package cert

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

var oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var oidNuts = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 54851}
var oidNutsVendor = oid(oidNuts, 4)
var oidNutsDomain = oid(oidNuts, 3)
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
	subjectAltName, err := marshalOtherSubjectAltName(oidNutsVendor, vendorID)
	if err != nil {
		return x509.CertificateRequest{}, err
	}
	extensions := []pkix.Extension{
		{Id: oidSubjectAltName, Critical: false, Value: subjectAltName},
	}
	if domain != "" {
		domainData, err := marshalNutsDomain(domain)
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

// IssueCertificate issues an X.509 certificate to entity 'subject' through Certificate Authority 'ca'. It assumes
// the CA is under control of the application since it expects the crypto module to directly issue the certificate.
// Both the subject's and CA's key pair should be available in the crypto module. If subject and CA are equal,
// it issues a self-signed certificate. Otherwise, the CA's certificate should also be present in the crypto module.
func IssueCertificate(crypt crypto.Client, csrTemplateFn func() (x509.CertificateRequest, error),
	subject types.LegalEntity, ca types.LegalEntity, profile crypto.CertificateProfile) ([]byte, error) {
	csrTemplate, err := csrTemplateFn()
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR template")
	}

	subjectPrivKey, err := crypt.GetOpaquePrivateKey(subject)
	if err != nil {
		return nil, errors2.Wrapf(err, "unable to retrieve subject private key: %s", subject)
	}

	csrTemplate.PublicKey = subjectPrivKey.Public()
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, subjectPrivKey)
	if err != nil {
		return nil, errors2.Wrap(err, "unable to create CSR")
	}

	certASN1, err := crypt.SignCertificate(subject, ca, csr, profile)
	if err != nil {
		return nil, errors2.Wrap(err, "error while signing certificate")
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		pem.Encode(os.Stdout, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certASN1,
		})
	}
	return certASN1, nil
}

func marshalOtherSubjectAltName(valueType asn1.ObjectIdentifier, value string) ([]byte, error) {
	// The structs below look funky, but are required to marshal SubjectAlternativeName.otherName the same way OpenSSL does.
	type otherNameValue struct {
		Value asn1.RawValue
	}
	type otherNameTypeAndValue struct {
		Type  asn1.ObjectIdentifier
		Value otherNameValue `asn1:"tag:0"`
	}
	type otherName struct {
		TypeAndValue otherNameTypeAndValue `asn1:"tag:0"`
	}
	return asn1.Marshal(otherName{TypeAndValue: otherNameTypeAndValue{
		Type:  valueType,
		Value: otherNameValue{asn1.RawValue{Tag: asn1.TagUTF8String, Bytes: []byte(value)}},
	}})
}

func marshalNutsDomain(domain string) ([]byte, error) {
	return asn1.Marshal(asn1.RawValue{
		Tag:   asn1.TagUTF8String,
		Bytes: []byte(domain),
	})
}

func oid(base asn1.ObjectIdentifier, v int) asn1.ObjectIdentifier {
	r := make([]int, len(base), len(base)+1)
	copy(r, base)
	return append(r, v)
}

