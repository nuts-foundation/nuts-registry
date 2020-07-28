package cert

import (
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVendorCertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := VendorCertificateRequest(test.VendorID("abc"), "def", "xyz", "care")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - optional params", func(t *testing.T) {
		csr, err := VendorCertificateRequest(test.VendorID("abc"), "def", "", "healthcare")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("err - no domain", func(t *testing.T) {
		_, err := VendorCertificateRequest(test.VendorID("abc"), "def", "", "")
		assert.EqualError(t, err, "missing domain")
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := VendorCertificateRequest(test.VendorID(""), "hello", "", "healthcare")
		assert.EqualError(t, err, "missing vendor identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := VendorCertificateRequest(test.VendorID("abc"), "", "", "healthcare")
		assert.EqualError(t, err, "missing vendor name")
	})
}

func TestOrganizationCertificateRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", test.OrganizationID("def"), "care", "healthcare")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("ok - no domain", func(t *testing.T) {
		csr, err := OrganisationCertificateRequest("abc", test.OrganizationID("def"), "care", "")
		if !assert.NoError(t, err) {
			return
		}
		assert.NotNil(t, csr)
	})
	t.Run("error: no ID", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", test.OrganizationID(""), "care", "healthcare")
		assert.EqualError(t, err, "missing organization identifier")
	})
	t.Run("error: no name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("abc", test.OrganizationID("def"), "", "healthcare")
		assert.EqualError(t, err, "missing organization name")
	})
	t.Run("error - no vendor name", func(t *testing.T) {
		_, err := OrganisationCertificateRequest("", test.OrganizationID("def"), "care", "healthcare")
		assert.EqualError(t, err, "missing vendor name")
	})
}
