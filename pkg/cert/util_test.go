package cert

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetDomain(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := VendorCertificateRequest("VendorID", "VendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		domain, err := GetDomain(cert)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", domain)
	})
	t.Run("no domain", func(t *testing.T) {
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1cert, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1cert)
		domain, err := GetDomain(cert)
		assert.Empty(t, domain)
		assert.NoError(t, err)
	})
}
func TestGetOrganizationSubjectAltName(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := OrganisationCertificateRequest("VendorID", "VendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		altName, err := GetOrganizationSubjectAltName(cert)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "VendorName", altName)
	})
}

func TestGetVendorSubjectAltName(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := VendorCertificateRequest("VendorID", "VendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		altName, err := GetVendorSubjectAltName(cert)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "VendorID", altName)
	})
}

func Test_getOtherSubjectAltName(t *testing.T) {
	type args struct {
		certificate *x509.Certificate
		oid         asn1.ObjectIdentifier
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getOtherSubjectAltName(tt.args.certificate, tt.args.oid)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOtherSubjectAltName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getOtherSubjectAltName() got = %v, want %v", got, tt.want)
			}
		})
	}
}