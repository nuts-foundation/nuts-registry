package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestRegisterVendorEvent(t *testing.T) {
	t.Run("check default domain fallback", func(t *testing.T) {
		event := events.CreateEvent(RegisterVendor, RegisterVendorEvent{})
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var registerVendorEvent = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&registerVendorEvent)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", registerVendorEvent.Domain)
	})
	t.Run("invalid JWK", func(t *testing.T) {
		event := events.CreateEvent(RegisterVendor, RegisterVendorEvent{
			Keys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var payload = RegisterVendorEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestRegisterVendorEvent_PostProcessUnmarshal(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest("abc", "Vendor", "CA", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := RegisterVendorEvent{
			Identifier: Identifier("abc"),
			Keys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event))
		assert.NoError(t, err)
	})
	t.Run("ok - fallback to healthcare domain", func(t *testing.T) {
		event := RegisterVendorEvent{}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event))
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "healthcare", event.Domain)
	})
	t.Run("certificate vendor doesn't match", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest("def", "Vendor", "CA", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := RegisterVendorEvent{
			Identifier: Identifier("abc"),
			Keys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(RegisterVendor, event))
		assert.EqualError(t, err, "vendor ID in certificate (def) doesn't match event (abc)")
	})
}

func TestVendorClaimEvent(t *testing.T) {
	t.Run("invalid JWK", func(t *testing.T) {
		event := events.CreateEvent(VendorClaim, VendorClaimEvent{
			VendorIdentifier: "v1",
			OrgKeys: []interface{}{
				map[string]interface{}{
					"kty": "EC",
				},
			},
		})
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var payload = VendorClaimEvent{}
		err := unmarshalledEvent.Unmarshal(&payload)
		assert.Contains(t, err.Error(), "invalid JWK")
	})
}

func TestVendorClaimEvent_PostProcessUnmarshal(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", "abc", "Org", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event))
		assert.NoError(t, err)
	})
	t.Run("ok - key does not contain certificate", func(t *testing.T) {
		key, _ := rsa.GenerateKey(rand.Reader, 1024)
		keyAsJwk, _ := jwk.New(key)
		jwkAsMap, _ := crypto.JwkToMap(keyAsJwk)
		jwkAsMap["kty"] = string(jwkAsMap["kty"].(jwa.KeyType))
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{jwkAsMap},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event))
		assert.NoError(t, err)
	})
	t.Run("error - certificate organization doesn't match", func(t *testing.T) {
		csr, _ := cert2.OrganisationCertificateRequest("Vendor", "def", "Org", HealthcareDomain)
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		event := VendorClaimEvent{
			OrgIdentifier: Identifier("abc"),
			OrgKeys:       []interface{}{certToMap(cert)},
		}
		err := event.PostProcessUnmarshal(events.CreateEvent(VendorClaim, event))
		assert.EqualError(t, err, "organization ID in certificate (def) doesn't match event (abc)")
	})
}

func TestRegisterEndpointEvent(t *testing.T) {
	t.Run("unmarshal event with no post processors", func(t *testing.T) {
		event := events.CreateEvent(RegisterEndpoint, RegisterEndpointEvent{})
		data := event.Marshal()
		unmarshalledEvent, _ := events.EventFromJSON(data)
		var registerEndpointEvent = RegisterEndpointEvent{}
		err := unmarshalledEvent.Unmarshal(&registerEndpointEvent)
		if !assert.NoError(t, err) {
			return
		}
	})
}

func TestGetEventTypes(t *testing.T) {
	assert.NotEmpty(t, GetEventTypes())
	for _, eventType := range GetEventTypes() {
		assert.NotEqual(t, "", string(eventType))
	}
}

func TestNewTrustStore(t *testing.T) {
	ts := NewTrustStore().(*trustStore)
	assert.NotNil(t, ts.certPool)
}

func Test_trustStore_HandleEvent(t *testing.T) {
	t.Run("ok - register vendor", func(t *testing.T) {
		ts := NewTrustStore().(*trustStore)
		csr, _ := cert2.VendorCertificateRequest("vendorId", "vendorName", "CA", "healthcare")
		cert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 1)
		key, _ := crypto.CertificateToJWK(cert)
		jwkAsMap, _ := crypto.JwkToMap(key)
		event := RegisterVendorEvent{
			Identifier: Identifier("vendorId"),
			Keys: []interface{}{jwkAsMap},
		}
		err := ts.handleEvent(events.CreateEvent(RegisterVendor, event))
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, ts.certPool.Subjects(), 1)
	})
	t.Run("error - register vendor - not self-signed", func(t *testing.T) {
		ts := NewTrustStore().(*trustStore)
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1Data, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1Data)
		key, _ := crypto.CertificateToJWK(cert)
		key.Set(jwk.X509CertChainKey, []string{base64cert, base64cert})
		jwkAsMap, _ := crypto.JwkToMap(key)
		event := RegisterVendorEvent{
			Keys: []interface{}{jwkAsMap},
		}
		err := ts.handleEvent(events.CreateEvent(RegisterVendor, event))
		assert.Equal(t, err.Error(), "unexpected X.509 certificate chain length for vendor (it should be self-signed)")
		assert.Len(t, ts.certPool.Subjects(), 0)
	})
	t.Run("error - vendor claim - organization certificate not trusted", func(t *testing.T) {
		ts := NewTrustStore().(*trustStore)
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1Data, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1Data)
		key, _ := crypto.CertificateToJWK(cert)
		jwkAsMap, _ := crypto.JwkToMap(key)
		event := VendorClaimEvent{
			OrgKeys: []interface{}{jwkAsMap},
		}
		err := ts.handleEvent(events.CreateEvent(VendorClaim, event))
		assert.Equal(t, "organization certificate is not trusted (issued by untrusted vendor certificate?): x509: certificate signed by unknown authority", err.Error())
		assert.Len(t, ts.certPool.Subjects(), 0)
	})
}

func Test_trustStore_RegisterEventHandlers(t *testing.T) {
	ts := NewTrustStore()
	called := false
	ts.RegisterEventHandlers(func(eventType events.EventType, handler events.EventHandler) {
		called = true
	})
	assert.True(t, called)
}

func Test_trustStore_Verify(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, _ := cert2.VendorCertificateRequest("vendorId", "vendorName", "CA", "healthcare")
		caCert, _ := test.SelfSignCertificateFromCSR(csr, time.Now(), 2)
		ts := NewTrustStore().(*trustStore)
		ts.certPool.AddCert(caCert)
		err := ts.Verify(caCert)
		assert.NoError(t, err)
	})
	t.Run("error - incorrect domain", func(t *testing.T) {
		caCsr, _ := cert2.VendorCertificateRequest("vendorId", "vendorName", "CA", "healthcare")
		caCert, caPrivKey := test.SelfSignCertificateFromCSR(caCsr, time.Now(), 2)
		csr, _ := cert2.VendorCertificateRequest("vendorId", "vendorName", "", "somethingelse")
		csr.PublicKey = &caPrivKey.PublicKey
		cert := test.SignCertificateFromCSRWithKey(csr, time.Now(), 2, caCert, caPrivKey)
		ts := NewTrustStore().(*trustStore)
		ts.certPool.AddCert(caCert)
		err := ts.Verify(cert)
		assert.EqualError(t, err, "domain (healthcare) in certificate (subject: CN=vendorName CA,O=vendorName,C=NL) differs from expected domain (somethingelse)")
	})
	t.Run("error - missing domain", func(t *testing.T) {
		base64cert := "MIIE3jCCA8agAwIBAgICAwEwDQYJKoZIhvcNAQEFBQAwYzELMAkGA1UEBhMCVVMxITAfBgNVBAoTGFRoZSBHbyBEYWRkeSBHcm91cCwgSW5jLjExMC8GA1UECxMoR28gRGFkZHkgQ2xhc3MgMiBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAeFw0wNjExMTYwMTU0MzdaFw0yNjExMTYwMTU0MzdaMIHKMQswCQYDVQQGEwJVUzEQMA4GA1UECBMHQXJpem9uYTETMBEGA1UEBxMKU2NvdHRzZGFsZTEaMBgGA1UEChMRR29EYWRkeS5jb20sIEluYy4xMzAxBgNVBAsTKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTEwMC4GA1UEAxMnR28gRGFkZHkgU2VjdXJlIENlcnRpZmljYXRpb24gQXV0aG9yaXR5MREwDwYDVQQFEwgwNzk2OTI4NzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQt1RWMnCZM7DI161+4WQFapmGBWTtwY6vj3D3HKrjJM9N55DrtPDAjhI6zMBS2sofDPZVUBJ7fmd0LJR4h3mUpfjWoqVTr9vcyOdQmVZWt7/v+WIbXnvQAjYwqDL1CBM6nPwT27oDyqu9SoWlm2r4arV3aLGbqGmu75RpRSgAvSMeYddi5Kcju+GZtCpyz8/x4fKL4o/K1w/O5epHBp+YlLpyo7RJlbmr2EkRTcDCVw5wrWCs9CHRK8r5RsL+H0EwnWGu1NcWdrxcx+AuP7q2BNgWJCJjPOq8lh8BJ6qf9Z/dFjpfMFDniNoW1fho3/Rb2cRGadDAW/hOUoz+EDU8CAwEAAaOCATIwggEuMB0GA1UdDgQWBBT9rGEyk2xF1uLuhV+auud2mWjM5zAfBgNVHSMEGDAWgBTSxLDSkdRMEXGzYcs9of7dqGrU4zASBgNVHRMBAf8ECDAGAQH/AgEAMDMGCCsGAQUFBwEBBCcwJTAjBggrBgEFBQcwAYYXaHR0cDovL29jc3AuZ29kYWRkeS5jb20wRgYDVR0fBD8wPTA7oDmgN4Y1aHR0cDovL2NlcnRpZmljYXRlcy5nb2RhZGR5LmNvbS9yZXBvc2l0b3J5L2dkcm9vdC5jcmwwSwYDVR0gBEQwQjBABgRVHSAAMDgwNgYIKwYBBQUHAgEWKmh0dHA6Ly9jZXJ0aWZpY2F0ZXMuZ29kYWRkeS5jb20vcmVwb3NpdG9yeTAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZIhvcNAQEFBQADggEBANKGwOy9+aG2Z+5mC6IGOgRQjhVyrEp0lVPLN8tESe8HkGsz2ZbwlFalEzAFPIUyIXvJxwqoJKSQ3kbTJSMUA2fCENZvD117esyfxVgqwcSeIaha86ykRvOe5GPLL5CkKSkB2XIsKd83ASe8T+5o0yGPwLPk9Qnt0hCqU7S+8MxZC9Y7lhyVJEnfzuz9p0iRFEUOOjZv2kWzRaJBydTXRE4+uXR21aITVSzGh6O1mawGhId/dQb8vxRMDsxuxN89txJx9OjxUUAiKEngHUuHqDTMBqLdElrRhjZkAzVvb3du6/KFUJheqwNTrZEjYx8WnM25sgVjOuH0aBsXBTWVU+4="
		asn1Data, _ := base64.StdEncoding.DecodeString(base64cert)
		cert, _ := x509.ParseCertificate(asn1Data)
		ts := NewTrustStore().(*trustStore)
		ts.certPool.AddCert(cert)
		err := ts.Verify(cert)
		assert.Contains(t, err.Error(), "certificate is missing domain")
	})
}

func certToMap(certificate *x509.Certificate) map[string]interface{} {
	key, _ := crypto.CertificateToJWK(certificate)
	keyAsMap, _ := crypto.JwkToMap(key)
	keyAsMap["kty"] = string(keyAsMap["kty"].(jwa.KeyType))
	return keyAsMap
}