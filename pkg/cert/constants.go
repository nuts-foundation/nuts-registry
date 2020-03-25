package cert

import (
	"encoding/asn1"
	asn1util "github.com/nuts-foundation/nuts-crypto/pkg/asn1"
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
