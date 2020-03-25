package cert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVendorCertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := VendorCertificateRequest("abc", "def", "xyz", "care")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - optional params", func(t *testing.T) {
		csr, err := VendorCertificateRequest("abc", "def", "", "healthcare")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("err - no domain", func(t *testing.T) {
		_, err := VendorCertificateRequest("abc", "def", "", "")
		assert.EqualError(t, err, "missing domain")
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := VendorCertificateRequest("", "hello", "", "healthcare")
		assert.EqualError(t, err, "missing vendor identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := VendorCertificateRequest("abc", "", "", "healthcare")
		assert.EqualError(t, err, "missing vendor name")
	})
}

func TestOrganizationCertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", "def", "care", "healthcare")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no domain", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", "def", "care", "")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", "", "care", "healthcare")
		assert.EqualError(t, err, "missing organization identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", "def", "", "healthcare")
		assert.EqualError(t, err, "missing organization name")
	})
	t.Run("error - no vendor name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("", "def", "care", "healthcare")
		assert.EqualError(t, err, "missing vendor name")
	})
}
