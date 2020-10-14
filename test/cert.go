package test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/nuts-foundation/nuts-crypto/test"

	"github.com/lestrrat-go/jwx/jws"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
)

func GenerateCertificateEx(notBefore time.Time, validityInDays int, privKey *rsa.PrivateKey) []byte {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Unit Test",
		},
		PublicKey:             privKey.PublicKey,
		NotBefore:             notBefore,
		NotAfter:              notBefore.AddDate(0, 0, validityInDays),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	data, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.Public(), privKey)
	if err != nil {
		panic(err)
	}
	return data
}

func GenerateCertificateCA(name string, signer *x509.Certificate, privKey *rsa.PrivateKey, signerKey *rsa.PrivateKey) []byte {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: name,
		},
		PublicKey:             privKey.PublicKey,
		NotBefore:             time.Now().AddDate(0, 0, -1),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	data, err := x509.CreateCertificate(rand.Reader, &template, signer, privKey.Public(), signerKey)
	if err != nil {
		panic(err)
	}
	return data
}

func SignCertificateFromCSRWithKey(csr x509.CertificateRequest, notBefore time.Time, validityInDays int, ca *x509.Certificate, caPrivKey crypto.Signer) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               csr.Subject,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		NotBefore:             notBefore,
		NotAfter:              notBefore.AddDate(0, 0, validityInDays),
		ExtraExtensions:       csr.ExtraExtensions,
		EmailAddresses:        csr.EmailAddresses,
		PublicKey:             csr.PublicKey,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	if ca == nil {
		ca = template
	}
	data, err := x509.CreateCertificate(rand.Reader, template, ca, csr.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}
	certificate, err := x509.ParseCertificate(data)
	if err != nil {
		panic(err)
	}
	return certificate
}

func SelfSignCertificateFromCSR(csr x509.CertificateRequest, notBefore time.Time, validityInDays int) (*x509.Certificate, *rsa.PrivateKey) {
	key := test.GenerateRSAKey()
	csr.PublicKey = &key.PublicKey
	return SignCertificateFromCSRWithKey(csr, notBefore, validityInDays, nil, key), key
}

// NoopJwsVerifier is a JwsVerifier that just parses the JWS without verifying the signatures
var NoopJwsVerifier = func(signature []byte, signingTime time.Time, verifier cert.Verifier) ([]byte, error) {
	msg, err := jws.Parse(bytes.NewReader(signature))
	if err != nil {
		return nil, err
	}
	return msg.Payload(), nil
}

var NoopCertificateVerifier cert.Verifier = &noopCertificateVerifier{}

type noopCertificateVerifier struct{}

func (n noopCertificateVerifier) Verify(certificate *x509.Certificate, t time.Time) error {
	return nil
}

func (n noopCertificateVerifier) VerifiedChain(certificate *x509.Certificate, t time.Time) ([][]*x509.Certificate, error) {
	return nil, nil
}
