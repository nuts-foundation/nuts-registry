package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/storage"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVendorCACertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := VendorCACertificateRequest("abc", "def", "care", "test")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no domain", func(t *testing.T) {
		csr, err := VendorCACertificateRequest("abc", "def", "", "test")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no environment", func(t *testing.T) {
		csr, err := VendorCACertificateRequest("abc", "def", "", "")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := VendorCACertificateRequest("", "hello", "", "")
		assert.EqualError(t, err, "missing vendor identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := VendorCACertificateRequest("abc", "", "", "")
		assert.EqualError(t, err, "missing vendor name")
	})
}

func TestIssueCertificate(t *testing.T) {
	repo, err := test.NewTestRepo(t.Name())
	if !assert.NoError(t, err) {
		return
	}
	defer repo.Cleanup()
	cryptoStorage, _ := storage.NewFileSystemBackend(repo.Directory)
	crypt := crypto.Crypto{
		Storage: cryptoStorage,
		Config: crypto.CryptoConfig{
			Keysize: types.ConfigKeySizeDefault,
		},
	}
	entity := types.LegalEntity{URI: "foo"}
	logrus.SetLevel(logrus.DebugLevel)

	t.Run("ok", func(t *testing.T) {
		crypt.GenerateKeyPairFor(entity)
		_, err = IssueCertificate(&crypt, func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{Subject: pkix.Name{CommonName: "Mosselman"}}, nil
		}, entity, entity, crypto.CertificateProfile{})
		assert.NoError(t, err)
	})
	t.Run("csr template fn error", func(t *testing.T) {
		_, err = IssueCertificate(&crypt, func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{}, errors.New("oops")
		}, entity, entity, crypto.CertificateProfile{})
		assert.Contains(t, err.Error(), "unable to create CSR template")
	})
	t.Run("key pair unavailable", func(t *testing.T) {
		unknownEntity := types.LegalEntity{URI: "unknown"}
		_, err = IssueCertificate(&crypt, func() (x509.CertificateRequest, error) {
			return x509.CertificateRequest{Subject: pkix.Name{CommonName: "Mosselman"}}, nil
		}, unknownEntity, unknownEntity, crypto.CertificateProfile{})
		assert.Contains(t, err.Error(), "unable to retrieve subject private key")
	})
}
