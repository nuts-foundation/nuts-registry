package cert

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
