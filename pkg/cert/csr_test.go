package cert

import (
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

func TestOrganizationCertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", "def", "care", "healthcare", "test")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no domain", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", "def", "care", "", "test")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no environment", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", "def", "care", "healthcare", "")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", "", "care", "healthcare", "test")
		assert.EqualError(t, err, "missing organization identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", "def", "", "healthcare", "test")
		assert.EqualError(t, err, "missing organization name")
	})
	t.Run("error - no vendor name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("", "def", "care", "healthcare", "test")
		assert.EqualError(t, err, "missing vendor name")
	})
}
