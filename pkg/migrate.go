package pkg

import (
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// verify verifies the data in the registry, migrating data whenever required (e.g. issue missing certificates) when autoFix=true.
// The events that result from fixing tge data are returned.
func (r *Registry) verify(config core.NutsConfigValues, autoFix bool) ([]events.Event, bool, error) {
	r.logger().Infof("Verifying registry integrity (autofix issues=%v)...", autoFix)
	resultingEvents := make([]events.Event, 0)
	// Assert vendor is registered
	identity := config.VendorID()
	vendor := r.Db.VendorByID(identity)
	fixRequired := false
	var event events.Event
	var err error
	if vendor == nil {
		err = fmt.Errorf("configured vendor (%s) is not registered, please register it using the 'register-vendor' CLI command", identity)
	} else {
		if event, fixRequired, err = r.verifyVendorCertificate(vendor, identity); event != nil {
			resultingEvents = append(resultingEvents, event)
		}
		if err != nil {
			return resultingEvents, fixRequired, err
		}
		for _, org := range r.Db.OrganizationsByVendorID(vendor.Identifier) {
			if event, fixRequired, err = r.verifyOrganisation(org, autoFix); event != nil {
				resultingEvents = append(resultingEvents, event)
			}
			if err != nil {
				return resultingEvents, fixRequired, err
			}
		}
	}
	if fixRequired {
		r.logger().Warn("Your registry data needs fixing/upgrading. Please run the following administrative command: `registry verify -f`")
	} else {
		if len(resultingEvents) > 0 {
			r.logger().Infof("Registry data fixed/upgraded (%d events were emitted).", len(resultingEvents))
		} else {
			r.logger().Info("Registry verification done.")
		}
	}
	return resultingEvents, fixRequired, err
}

func (r *Registry) verifyVendorCertificate(vendor *db.Vendor, identity core.PartyID) (events.Event, bool, error) {
	certificates := vendor.GetActiveCertificates()
	if len(certificates) == 0 {
		r.logger().Warn("No active certificates found for configured vendor.")
		return nil, false, nil
	} else {
		if !r.crypto.PrivateKeyExists(types.KeyForEntity(types.LegalEntity{URI: identity.String()})) {
			return nil, false, errors.New("active certificates were found for configured vendor, but there's no private key available for cryptographic operations. Please recover your key material")
		}
	}
	return nil, false, nil
}

func (r *Registry) verifyOrganisation(org *db.Organization, autoFix bool) (events.Event, bool, error) {
	certificates := org.GetActiveCertificates()
	if len(certificates) == 0 {
		logrus.Warnf("No active certificates found for organisation (id = %s).", org.Identifier)
		if autoFix {
			event, err := r.RefreshOrganizationCertificate(org.Identifier)
			if err != nil {
				return nil, false, errors2.Wrap(err, "couldn't issue organization certificate")
			}
			return event, false, nil
		}
		return nil, true, nil
	} else {
		if !r.crypto.PrivateKeyExists(types.KeyForEntity(types.LegalEntity{URI: org.Identifier.String()})) {
			return nil, false, fmt.Errorf("active certificates were found for organisation (id = %s), but there's no private key available for cryptographic operations. Please recover your key material", org.Identifier)
		}
	}
	return nil, false, nil
}
