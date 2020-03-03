package test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
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
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	data, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.Public(), privKey)
	if err != nil {
		panic(err)
	}
	return data
}

func SignCertificateFromCSRWithKey(csr x509.CertificateRequest, notBefore time.Time, validityInDays int, ca *x509.Certificate, caPrivKey crypto.Signer) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber:    big.NewInt(1),
		Subject:         csr.Subject,
		KeyUsage:        x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		NotBefore:       notBefore,
		NotAfter:        notBefore.AddDate(0, 0, validityInDays),
		ExtraExtensions: csr.ExtraExtensions,
		PublicKey:       csr.PublicKey,
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

func SelfSignCertificateFromCSR(csr x509.CertificateRequest, notBefore time.Time, validityInDays int) *x509.Certificate {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	csr.PublicKey = &key.PublicKey
	return SignCertificateFromCSRWithKey(csr, notBefore, validityInDays, nil, key)
}