package domain

import "github.com/nuts-foundation/nuts-registry/pkg/events"

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

func GetEventTypes() []events.EventType {
	return []events.EventType{
		RegisterEndpoint,
		RegisterVendor,
		VendorClaim,
	}
}

// Identifier defines component schema for Identifier.
type Identifier string

func (i Identifier) String() string {
	return string(i)
}