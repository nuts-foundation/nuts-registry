package types

type EndpointID string

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

