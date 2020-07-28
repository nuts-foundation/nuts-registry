package domain

import (
	"fmt"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	cert2 "github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	errors2 "github.com/pkg/errors"
)

// RegisterVendor event type
const RegisterVendor events.EventType = "RegisterVendorEvent"

// RegisterVendorEvent event
type RegisterVendorEvent struct {
	Identifier core.PartyID  `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

func (r *RegisterVendorEvent) PostProcessUnmarshal(event events.Event) error {
	// Default fallback to 'healthcare' domain when none is set, for handling legacy data when 'domain' didn't exist.
	if r.Domain == "" {
		r.Domain = types.FallbackDomain
	}
	if err := cert.ValidateJWK(r.Keys...); err != nil {
		return err
	}
	for _, key := range r.Keys {
		var vendorID core.PartyID
		if certificate := cert.GetCertificate(key); certificate != nil {
			var err error
			if vendorID, err = cert2.NewNutsCertificate(certificate).GetVendorID(); err != nil {
				return errors2.Wrap(err, "unable to unmarshal vendor ID from certificate")
			}
		}
		if r.Identifier != vendorID {
			return fmt.Errorf("vendor ID in certificate (%s) doesn't match event (%s)", vendorID, r.Identifier)
		}
	}
	return nil
}

// VendorEventMatcher returns an EventMatcher which matches the RegisterVendorEvent for the vendor with the specified ID.
func VendorEventMatcher(vendorID core.PartyID) events.EventMatcher {
	return func(event events.Event) bool {
		if event.Type() != RegisterVendor {
			return false
		}
		var p = RegisterVendorEvent{}
		_ = event.Unmarshal(&p)
		return vendorID == p.Identifier
	}
}
