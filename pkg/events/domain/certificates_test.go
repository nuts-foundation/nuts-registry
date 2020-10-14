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
package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/nuts-foundation/nuts-registry/pkg/types"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
)

func TestNewCertificateEventHandler(t *testing.T) {
	ts := NewCertificateEventHandler(nil)
	assert.NotNil(t, ts)
}

func Test_CertificateEventHandler_HandleEvent(t *testing.T) {
	t.Run("ok - register vendor", func(t *testing.T) {
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("vendorId"), "vendorName", "CA", "healthcare")
		certificate, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 1)
		key, _ := cert.CertificateToJWK(certificate)
		jwkAsMap, _ := cert.JwkToMap(key)
		event := RegisterVendorEvent{
			Identifier: test.VendorID("vendorId"),
			Keys:       []interface{}{jwkAsMap},
		}
		err := handler.handleEvent(events.CreateEvent(RegisterVendor, event, nil), nil)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, handler.trustStore.Pool().Subjects(), 1)
	})
	t.Run("ok - register vendor - self-signed", func(t *testing.T) {
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1Data, _ := base64.StdEncoding.DecodeString(base64cert)
		certificate, _ := x509.ParseCertificate(asn1Data)
		key, _ := cert.CertificateToJWK(certificate)
		key.Set(jwk.X509CertChainKey, []string{base64cert, base64cert})
		jwkAsMap, _ := cert.JwkToMap(key)
		event := RegisterVendorEvent{
			Keys: []interface{}{jwkAsMap},
		}
		err := handler.handleEvent(events.CreateEvent(RegisterVendor, event, nil), nil)
		assert.NoError(t, err)
		assert.Len(t, handler.trustStore.Pool().Subjects(), 1)
	})
	t.Run("ok - vendor claim", func(t *testing.T) {
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		caCsr, _ := cert2.VendorCertificateRequest(test.VendorID("vendorId"), "vendorName", "CA", "healthcare")
		caCert, caKey := test.SelfSignCertificateFromCSR(caCsr, time.Now(), 1)
		handler.trustStore.AddCertificate(caCert)

		orgID := test.OrganizationID("orgID")
		csr, _ := cert2.OrganisationCertificateRequest("vendorName", orgID, "orgName", "healthcare")
		csr.PublicKey = &caKey.PublicKey // strange but ok for this test
		certificate := test.SignCertificateFromCSRWithKey(csr, time.Now(), 1, caCert, caKey)

		key, _ := cert.CertificateToJWK(certificate)
		jwkAsMap, _ := cert.JwkToMap(key)
		event := VendorClaimEvent{
			OrganizationID: orgID,
			OrgKeys:        []interface{}{jwkAsMap},
		}
		err := handler.handleEvent(events.CreateEvent(VendorClaim, event, nil), nil)
		assert.NoError(t, err)
	})
	t.Run("ok - vendor claim - no certs", func(t *testing.T) {
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		keyAsJWK, _ := jwk.New(privateKey)
		jwkAsMap, _ := cert.JwkToMap(keyAsJWK)
		err := handler.handleEvent(events.CreateEvent(VendorClaim, VendorClaimEvent{OrgKeys: []interface{}{jwkAsMap}}, nil), nil)
		assert.NoError(t, err)
	})
	t.Run("error - vendor claim - organization certificate not trusted", func(t *testing.T) {
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		orgID := test.OrganizationID("abc")
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", orgID, "Org Name", types.HealthcareDomain)
		certificate, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		key, _ := cert.CertificateToJWK(certificate)
		jwkAsMap, _ := cert.JwkToMap(key)
		event := VendorClaimEvent{
			OrganizationID: orgID,
			OrgKeys:        []interface{}{jwkAsMap},
		}
		err := handler.handleEvent(events.CreateEvent(VendorClaim, event, nil), nil)
		assert.EqualError(t, err, "certificate problem in VendorClaim event: certificate not trusted: CN=Org Name,O=Vendor,C=NL (issuer: CN=Org Name,O=Vendor,C=NL, serial: 1): x509: certificate signed by unknown authority")
		assert.Len(t, handler.trustStore.Pool().Subjects(), 0)
	})
}

func Test_CertificateEventHandler_RegisterEventHandlers(t *testing.T) {
	handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()})
	called := false
	handler.RegisterEventHandlers(func(eventType events.EventType, handler events.EventHandler) {
		called = true
	})
	assert.True(t, called)
}

func Test_CertificateEventHandler_Verify(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("vendorId"), "vendorName", "CA", "healthcare")
		caCert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		eventHandler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		eventHandler.trustStore.AddCertificate(caCert)
		chain, err := eventHandler.verify(caCert, time.Now())
		assert.NoError(t, err)
		assert.Len(t, chain, 1)
		err = eventHandler.Verify(caCert, time.Now())
		assert.NoError(t, err)
	})
	t.Run("error - incorrect domain", func(t *testing.T) {
		caCsr, _ := cert2.VendorCertificateRequest(test.VendorID("vendorId"), "vendorName", "CA", "healthcare")
		caCert, caPrivKey := test.SelfSignCertificateFromCSR(caCsr, time.Now(), 2)
		csr, _ := cert2.VendorCertificateRequest(test.VendorID("vendorId"), "vendorName", "", "somethingelse")
		csr.PublicKey = &caPrivKey.PublicKey
		cert := test.SignCertificateFromCSRWithKey(csr, time.Now(), 2, caCert, caPrivKey)
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		handler.trustStore.AddCertificate(caCert)
		_, err := handler.verify(cert, time.Now())
		// No error for now, just a warning
		assert.NoError(t, err)
	})
	t.Run("error - missing domain", func(t *testing.T) {
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1Data, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1Data)
		handler := NewCertificateEventHandler(memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		handler.trustStore.AddCertificate(cert)
		_, err := handler.verify(cert, time.Now())
		// No error for now, just a warning
		assert.NoError(t, err)
	})
	t.Run("error - certificate has expired", func(t *testing.T) {
		base64cert := "MIIDWTCCAkGgAwIBAgIIPFp5m2SmCOMwDQYJKoZIhvcNAQELBQAwPTELMAkGA1UEBhMCTkwxDjAMBgNVBAoTBU5lZGFwMR4wHAYDVQQDExVOZWRhcCBDQSBJbnRlcm1lZGlhdGUwHhcNMjAwNDAyMDgxODQwWhcNMjAwNDAyMDgxOTQwWjAtMQswCQYDVQQGEwJOTDEOMAwGA1UEChMFTmVkYXAxDjAMBgNVBAMTBU5lZGFwMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvWB8KpKRNV1RnyzkHTV4eXb8AFQPhT8DuyFsMGpYrRFV9RFXG6RjgoYVMbi9GyT/yddN92YEcDvTaLUlwxZAWeTB4uorDM69ezwrJNlFDn0T3E9Vz8tU8Yi4ZvLwD4RY5vINK8sHhVlx829/yJV0Fa8vY1t4I3M2LQMDjhPeV8QX72b6Y54fAZfAbIycrIiZOzTuoHTKjqk7nU7+aTEH7M9uXIrAEMWz7bF/rgwRhFJ2vss+ca843vtfnctOAS15ku5g661O/E/zqLyjEyUq1N4i0sI85wnopXSAiRdb4IlQm1Jht9gqL+TTx0Hio8rQk2zTpjGYnOoQ44JUGMM4KQIDAQABo20wazAOBgNVHQ8BAf8EBAMCB4AwPgYDVR0RBDcwNaAzBgkrBgEEAYOsQwSgJgwkdXJuOm9pZDoxLjMuNi4xLjQuMS41NDg1MS40OjAwMDAwMDAxMBkGCSsGAQQBg6xDAwQMDApoZWFsdGhjYXJlMA0GCSqGSIb3DQEBCwUAA4IBAQAw07XWznfAZBzhlOW9Z2/XuAsvmQMEHo8FdYV+9RdxS1YNnlIVSIlXrhuaoksJS2pKqzPJ211E0KpGx8m6YcHxKdm9Hm8kVz7GMpqRLmT1KtDoiWkt/aPGvGg8vDcq4wCeJbAmkDYmh8H2L5Asb2FRZxcvK4jP6jLTAyDDQYBJS9cLXDUQhcC6LIXKP4QRNZzuse2hgtIqRakfOSFHDudquarllvWvn2wZ5ZiqNbV586oWzMYsN8rgnK6v/UpwHkXJl48oMuQuM0N8JUQUPvgJL+p9YVMnb/Z9zlRClF52Swad+Cl6BoPa5Mt3Lwh2XYhC1a1SdsXbuNFaSjvjjxnx"
		asn1cert, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1cert)
		handler := NewCertificateEventHandler(&memoryTrustStore{certPool: x509.NewCertPool()}).(*certificateEventHandler)
		handler.trustStore.AddCertificate(cert)
		_, err := handler.verify(cert, time.Now())
		assert.Contains(t, err.Error(), "x509: certificate has expired or is not yet valid")
	})
}

func certToMap(certificate *x509.Certificate) map[string]interface{} {
	key, _ := cert.CertificateToJWK(certificate)
	keyAsMap, _ := cert.JwkToMap(key)
	keyAsMap["kty"] = string(keyAsMap["kty"].(jwa.KeyType))
	return keyAsMap
}

type memoryTrustStore struct {
	certPool *x509.CertPool
}

func (n memoryTrustStore) GetRoots(t time.Time) []*x509.Certificate {
	return nil
}

func (n memoryTrustStore) GetCertificates(i [][]*x509.Certificate, t time.Time, b bool) [][]*x509.Certificate {
	return nil
}

func (n memoryTrustStore) Pool() *x509.CertPool {
	return n.certPool
}

func (n memoryTrustStore) AddCertificate(certificate *x509.Certificate) error {
	n.certPool.AddCert(certificate)
	return nil
}

func (n memoryTrustStore) Verify(c *x509.Certificate, t time.Time) error {
	return errors.New("irrelevant func")
}

func (t memoryTrustStore) VerifiedChain(certificate *x509.Certificate, moment time.Time) ([][]*x509.Certificate, error) {
	return nil, errors.New("irrelevant func")
}

func (n memoryTrustStore) RegisterEventHandlers(func(events.EventType, events.EventHandler)) {
	// Nothing to do here
}
