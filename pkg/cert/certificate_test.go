package cert

import (
	"crypto/x509"
	"encoding/base64"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNutsCertificate_GetDomain(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := VendorCertificateRequest(test.VendorID("VendorID"), "VendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		domain, err := NutsCertificate(*cert).GetDomain()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", domain)
	})
	t.Run("no domain", func(t *testing.T) {
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1cert, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1cert)
		domain, err := NutsCertificate(*cert).GetDomain()
		assert.Empty(t, domain)
		assert.NoError(t, err)
	})
}
func TestNutsCertificate_GetOrganizationID(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expected := test.OrganizationID("OrganizationID")
		csr, _ := OrganisationCertificateRequest("VendorName", expected, "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		altName, err := NutsCertificate(*cert).GetOrganizationID()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, expected, altName)
	})
}

func TestNutsCertificate_GetVendorID(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		vendorID := test.VendorID("VendorID")
		csr, _ := VendorCertificateRequest(vendorID, "VendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		altName, err := NutsCertificate(*cert).GetVendorID()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, vendorID, altName)
	})
}
