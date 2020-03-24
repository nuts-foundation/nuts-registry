package events

import (
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"time"
)

const (
	// RegisterEndpoint event type
	RegisterEndpoint EventType = "RegisterEndpointEvent"
	// RegisterVendor event type
	RegisterVendor EventType = "RegisterVendorEvent"
	// VendorClaim event type
	VendorClaim EventType = "VendorClaimEvent"
)

const (
	// HealthcareDomain is a const for domain 'healthcare'
	HealthcareDomain string = "healthcare"
	// PersonalDomain is a const for domain 'personal' (which are "PGO's")
	PersonalDomain = "personal"
	// InsuranceDomain is a const for domain 'insurance'
	InsuranceDomain = "insurance"
	// FallbackDomain is a const for the fallback domain in case there's no domain set, which can be the case for legacy data.
	FallbackDomain = HealthcareDomain
)

func init() {
	eventTypes = []EventType{
		RegisterEndpoint,
		RegisterVendor,
		VendorClaim,
	}
}

// Identifier defines component schema for Identifier.
type Identifier string

// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Organization Identifier        `json:"organization"`
	URL          string            `json:"URL"`
	EndpointType string            `json:"endpointType"`
	Identifier   Identifier        `json:"identifier"`
	Status       string            `json:"status"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// RegisterVendorEvent event
type RegisterVendorEvent struct {
	Identifier Identifier    `json:"identifier"`
	Name       string        `json:"name"`
	Domain     string        `json:"domain,omitempty"`
	Keys       []interface{} `json:"keys,omitempty"`
}

func (r *RegisterVendorEvent) unmarshalPostProcess() error {
	// Default fallback to 'healthcare' domain when none is set, for handling legacy data when 'domain' didn't exist.
	if r.Domain == "" {
		r.Domain = FallbackDomain
	}
	if err := crypto.ValidateJWK(r.Keys...); err != nil {
		return err
	}
	return nil
}

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

func (v VendorClaimEvent) unmarshalPostProcess() error {
	if err := crypto.ValidateJWK(v.OrgKeys...); err != nil {
		return err
	}
	return nil
}