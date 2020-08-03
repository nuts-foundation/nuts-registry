package domain

import (
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"time"
)

// VendorClaim event type
const VendorClaim events.EventType = "VendorClaimEvent"

// VendorClaimEvent event
type VendorClaimEvent struct {
	VendorID       core.PartyID `json:"vendorIdentifier"`
	OrganizationID core.PartyID `json:"orgIdentifier"`
	OrgName        string       `json:"orgName"`
	// OrgKeys is a list of JWKs which are used to
	// 1. encrypt data to be decrypted by the organization,
	// 2. sign consent JWTs,
	// 3. sign organization related events (e.g. endpoint registration).
	OrgKeys []interface{} `json:"orgKeys,omitempty"`
	Start   time.Time     `json:"start"`
	End     *time.Time    `json:"end,omitempty"`
}

func (v VendorClaimEvent) PostProcessUnmarshal(event events.Event) error {
	if err := cert.ValidateJWK(v.OrgKeys...); err != nil {
		return err
	}
	for _, key := range v.OrgKeys {
		if certificate := cert.GetCertificate(key); certificate != nil {
			// Backwards compatibility: <= 0.14 VendorClaimEvent was signed by organization certificate, now it's the
			// vendor's signing certificate. So either the organization ID in the certificate must match (<= 0.14) or
			// the vendor ID (>= 0.15)
			if valid, err := v.isVendorIDValid(certificate); err != nil {
				return errors2.Wrap(err, "vendor ID validation failed")
			} else if valid {
				continue
			}
			if valid, err := v.isOrganizationIDValid(certificate); err != nil {
				return errors2.Wrap(err, "organization ID validation failed")
			} else if !valid {
				return errors.New("event should either be signed by organization or vendor signing certificate")
			}
		}
	}
	return nil
}

func (v VendorClaimEvent) isVendorIDValid(certificate *x509.Certificate) (bool, error) {
	var vendorID core.PartyID
	var err error
	if vendorID, err = cert2.NewNutsCertificate(certificate).GetVendorID(); err != nil {
		return false, errors2.Wrap(err, "unable to unmarshal vendor ID from certificate")
	}
	if !vendorID.IsZero() && v.VendorID != vendorID {
		return false, fmt.Errorf("vendor ID in certificate (%s) doesn't match event (%s)", vendorID, v.VendorID.Value())
	}
	return !vendorID.IsZero(), nil
}

func (v VendorClaimEvent) isOrganizationIDValid(certificate *x509.Certificate) (bool, error) {
	var organizationID core.PartyID
	var err error
	if organizationID, err = cert2.NewNutsCertificate(certificate).GetOrganizationID(); err != nil {
		return false, errors2.Wrap(err, "unable to unmarshal organization ID from certificate")
	}
	if !organizationID.IsZero() && v.OrganizationID != organizationID {
		return false, fmt.Errorf("organization ID in certificate (%s) doesn't match event (%s)", organizationID, v.OrganizationID)
	}
	return !organizationID.IsZero(), nil
}

// OrganizationEventMatcher returns an EventMatcher which matches the VendorClaimEvent which registered the organization
// with the specified ID for the specified vendor (also by ID).
func OrganizationEventMatcher(vendorID core.PartyID, organizationID core.PartyID) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != VendorClaim {
			return false
		}
		var payload = VendorClaimEvent{}
		_ = event.Unmarshal(&payload)
		return vendorID == payload.VendorID && organizationID == payload.OrganizationID
	}
}
