package domain

import (
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	errors2 "github.com/pkg/errors"
	"time"
)

// VendorClaim event type
const VendorClaim events.EventType = "VendorClaimEvent"

// VendorClaimEvent event
type VendorClaimEvent struct {
	VendorIdentifier Identifier `json:"vendorIdentifier"`
	OrgIdentifier    Identifier `json:"orgIdentifier"`
	OrgName          string     `json:"orgName"`
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
			var err error
			var organizationId string
			if organizationId, err = cert2.GetOrganizationSubjectAltName(certificate); err != nil {
				return errors2.Wrap(err, "unable to unmarshal organization ID from certificate")
			}
			if string(v.OrgIdentifier) != organizationId {
				return fmt.Errorf("organization ID in certificate (%s) doesn't match event (%s)", organizationId, v.OrgIdentifier)
			}
		}
	}
	return nil
}

// OrganizationEventMatcher returns an EventMatcher which matches the VendorClaimEvent which registered the organization
// with the specified ID for the specified vendor (also by ID).
func OrganizationEventMatcher(vendorID string, organizationID string) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != VendorClaim {
			return false
		}
		var payload = VendorClaimEvent{}
		_ = event.Unmarshal(&payload)
		return Identifier(vendorID) == payload.VendorIdentifier && Identifier(organizationID) == payload.OrgIdentifier
	}
}
