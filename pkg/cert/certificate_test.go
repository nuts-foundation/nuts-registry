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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var certAsBase64 = "MIIFjjCCA3agAwIBAgIIExaSHgjVVKcwDQYJKoZIhvcNAQELBQAwRzELMAkGA1UEBhMCTkwxEzARBgNVBAoTCk5lZGFwIE4uVi4xIzAhBgNVBAMTGk5lZGFwIE4uVi4gQ0EgSW50ZXJtZWRpYXRlMB4XDTIwMDUyODA3NTgxN1oXDTIxMDUyODA3NTgxN1owRzELMAkGA1UEBhMCTkwxEzARBgNVBAoTCk5lZGFwIE4uVi4xIzAhBgNVBAMTGk5lZGFwIE4uVi4gQ0EgSW50ZXJtZWRpYXRlMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAkNsazayY2LeQrwfoeHtYaUQTF1dzW5M6kME3KMVNqzD6zS1ElqtWhhEnncrDFml2LHmhqgYjNbCvuJl8r6mUHsxk8LnfDkqJzsmhV3Lgv3iSyFYsKicTL0xldJmTacmr4VN2nQ9d/EJHTKtmz4s8b3Tc6ZG9rSXHHa0a94zdHzG2AABcHIeOh234VzYht9qnoAXtOOkr2zdA9zSbWbnzHC90MLjp/xV7AS2UI/tgbCM0dLI2OkioA4FFWg+tM20T+tEI5oi9Y9v5rNkSvR5Qwwz4nQUmOxZKiTGM+/ApDrF8/gf/eWwDny75jt1OGZ4/X2ga+3lNO9cv/iED+uJqP3bz4Qg/5oynSfP7iol4lpOWBHV30u//e2Eb1NXKogsrmfysjdN2gcSb3xzFokmDjYy7czfvLCLyO97NbIZvSYCMGNuQds7Kc63vNOBu0t/f+UICt4UVuWlZqOOZiQw9XIPLkX3VrigzGQLeFLPf8l5SQ1641lLbUE9fvP9hzcox3Z7nbm1WRNGdBHZnv1AOQSa0kBzDsXOWezfiUAjMJxS6v1E+mRrn5scdePpNrpPSFqTnsre4QrPKInGg+nnCAenNKde04SrWzk5EkkzZworwPyMtGSkoGCmkZM5XAprECQJegS+1thz/sJSt8GnKPs0ZOmUe0sFTglSreLHxxxECAwEAAaN+MHwwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wPgYDVR0RBDcwNaAzBgkrBgEEAYOsQwSgJgwkdXJuOm9pZDoxLjMuNi4xLjQuMS41NDg1MS40OjA4MDEzODM2MBkGCSsGAQQBg6xDAwQMDApoZWFsdGhjYXJlMA0GCSqGSIb3DQEBCwUAA4ICAQATYZG92fcffCGrHTMMCaWBm+SZ50yTDKyPiHe+ff+3grVvn2U2LGirP6C4hQzrdkBLr8CyBfXgQhKNfvtIRSUUriM5JpEAvrdKtcWqM8/22yMPxaFqxSW/9JGNdog0e1B/oa4dk2Nx4gHId8HAiY/HYPwMnZMuKH4m9h+RlicZkmDEIkIWBA5eWz7ESgxfEKWQDrh8Z5oV/sbRfqWprZk/8w9TSv6GPk8gN8RsEkXYWcVC4Md9B1GF6I8HJ+H9bh7eZiufriZb/2myxgv2Y7VljmESRGDkP4PV3WnK0dBDH54r3EIaB9RlB2iY00z8yaGt8c+NuOQ+T/Yniru7WL1q8QKWxZNbbmDoNZ2hQDAZ7PTOq96PuCJXbG2PgGmJ2Zrx62nnNCxvyFsA16Cj+CZ0AA4GTkj2mt+QQpU6GlMi9n0TiQzh85k3U5y1cJ5PvkQUcOlvN0ZnaqSA4zm+dHHxto3NgHQFHYTbSI4AAVizIcKQoqV1ctms5xLdsbAHGOUGkgt+Ivi0tKat8NfQwOMLK+LXc72MXrWpb25uW3qHpY8Y+lkxEu28Tram2NrAXOSxTv7peimd+PAf0AfYoV2tZTHAuCkmXouxCe9G9tjyDxIiBK1Z0s/QfwPb9DF+M6km+Bvk62sk+ILWfLLUAava9qQ7VuLaEncXabLdzdBclQ=="
var certAsBytes, _ = base64.StdEncoding.DecodeString(certAsBase64)
var certificate, _ = x509.ParseCertificate(certAsBytes)

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
	t.Run("ok - certificate contains e-mail address SAN", func(t *testing.T) {
		privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		cert := test.SignCertificateFromCSRWithKey(x509.CertificateRequest{
			PublicKey:      privateKey.Public(),
			Subject:        pkix.Name{CommonName: "Foobar"},
			EmailAddresses: []string{"foo@bar.nl"},
		}, time.Now(), 2, nil, privateKey)
		actual, err := NutsCertificate(*cert).GetVendorID()
		assert.NoError(t, err)
		assert.True(t, actual.IsZero())
	})
	t.Run("ok - v0.14 self-signed certificate", func(t *testing.T) {
		nutsCertificate := NewNutsCertificate(certificate)
		vendorID, err := nutsCertificate.GetVendorID()
		assert.NoError(t, err)
		assert.Equal(t, "urn:oid:1.3.6.1.4.1.54851.4:08013836", vendorID.String())
	})
}

func TestNutsCertificate_MarshalJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expected := NewNutsCertificate(certificate)
		var certAsJSON []byte
		var err error
		t.Run("marshal", func(t *testing.T) {
			certAsJSON, err = expected.MarshalJSON()
			assert.NoError(t, err)
		})
		t.Run("unmarshal", func(t *testing.T) {
			var actual NutsCertificate
			err = json.Unmarshal(certAsJSON, &actual)
			assert.NoError(t, err)
			assert.Equal(t, expected.Raw, actual.Raw)
		})
	})
	t.Run("error - data is not base64 encoded", func(t *testing.T) {
		var actual NutsCertificate
		err := json.Unmarshal([]byte(`"not base64"`), &actual)
		assert.Error(t, err)
		assert.Empty(t, actual.Raw)
	})
	t.Run("error - data is not a certificate", func(t *testing.T) {
		var actual NutsCertificate
		err := json.Unmarshal([]byte(`"CAFEBABE"`), &actual)
		assert.Error(t, err)
		assert.Empty(t, actual.Raw)
	})
}
