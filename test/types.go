package test

import (
	"encoding/asn1"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	"github.com/nuts-foundation/nuts-go-core"
)

// OrganizationID is a helper function which creates PartyIDs for vendors.
func VendorID(value string) core.PartyID {
	partyID, _ := core.NewPartyID(cert.OIDNutsVendor.String(), value)
	return partyID
}

// OrganizationID is a helper function which creates PartyIDs for organizations.
func OrganizationID(value string) core.PartyID {
	partyID, _ := core.NewPartyID(asn1.ObjectIdentifier{2, 16, 840, 1, 113883, 2, 4, 6, 1}.String(), value)
	return partyID
}
