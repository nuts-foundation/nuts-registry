package domain

import (
	"github.com/nuts-foundation/nuts-registry/pkg/events"
)

func GetEventTypes() []events.EventType {
	return []events.EventType{
		RegisterEndpoint,
		RegisterVendor,
		VendorClaim,
	}
}