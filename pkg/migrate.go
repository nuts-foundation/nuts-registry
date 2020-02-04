package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
)

// verifyAndMigrateRegistry verifies the data in the registry, migrating data whenever required (e.g. issue missing certificates).
// When a failed verification cannot be migrated, an error is returned.
func (r *Registry) verifyAndMigrateRegistry(config core.NutsConfigValues) {
	r.logger().Info("Verifying registry integrity...")
	// Assert vendor is registered
	identity := config.Identity()
	vendor := r.Db.VendorByID(identity)
	if vendor == nil {
		logrus.Warnf("Configured vendor (%s) is not registered, please register it using the 'register-vendor' CLI command.", identity)
	} else {
		r.verifyAndMigrateVendorCertificates(vendor, identity)
		for _, org := range r.Db.OrganizationsByVendorID(vendor.Identifier.String()) {
			r.verifyAndMigrateOrganisation(org)
		}
	}
	r.logger().Info("Registry verification done.")
}

func (r *Registry) verifyAndMigrateVendorCertificates(vendor *db.Vendor, identity string) {
	certificates := vendor.GetActiveCertificates()
	if len(certificates) == 0 {
		logrus.Info("No active certificates found for configured vendor, this will be auto-migrated in the next version.")
	} else {
		if !r.crypto.KeyExistsFor(types.LegalEntity{URI: identity}) {
			logrus.Error("Active certificates were found for configured vendor, but there's no private key available for cryptographic operations. Please recover your key material.")
		}
	}
}

func (r *Registry) verifyAndMigrateOrganisation(org *db.Organization) {
	certificates := org.GetActiveCertificates()
	if len(certificates) == 0 {
		logrus.Infof("No active certificates found for organisation (id = %s), this will be auto-migrated in the next version.", org.Identifier)
	} else {
		if !r.crypto.KeyExistsFor(types.LegalEntity{URI: org.Identifier.String()}) {
			logrus.Errorf("Active certificates were found for organisation (id = %s), but there's no private key available for cryptographic operations. Please recover your key material.", org.Identifier)
		}
	}
}
