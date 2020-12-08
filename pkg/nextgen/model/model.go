package model

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/cert"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"time"
)

// Proof is a container for a cryptographic proof (TODO)
type Proof interface {
}

// Vendor is a JSON marshalable implementation of a vendor as specified by RFC006 Distributed Registry.
type Vendor struct {
	// Certificates contains the CA certificates of the organization
	Certificates []*cert.NutsCertificate `json:"certs"`
	// Proofs contains the pieces of cryptographically verifiable information that authenticate the vendor. (TODO)
	Proofs []Proof `json:"prfs,omitempty"`
}

// Organization is a JSON marshalable implementation of an organization as specified by RFC006 Distributed Registry.
type Organization struct {
	// ID of the organization
	ID core.PartyID `json:"id"`
	// Name contains the name of this organization.
	Name string `json:"name"`
	// VendorID contains the ID of the vendor that registered this organization.
	VendorID core.PartyID `json:"vid"`
	// Proofs contains the pieces of cryptographically verifiable information that authenticate the organization. (TODO)
	Proofs []Proof `json:"prfs,omitempty"`
}

// Endpoint is a JSON marshalable implementation of an endpoint as specified by RFC006 Distributed Registry.
type Endpoint struct {
	// ID contains the ID of this endpoint. It must be unique for the vendor.
	ID types.EndpointID `json:"id"`
	// VendorID contains the ID of the vendor which owns this endpoint.
	VendorID core.PartyID `json:"vid"`
	// NotBefore contains the date/time indicating from when the endpoint could be used.
	NotBefore time.Time `json:"nbf"`
	// Expiry contains the date/time on or after which the service should not be used anymore. Optional.
	Expiry *time.Time `json:"exp,omitempty"`
	// Location contains the URL exposed by this endpoint.
	Location URL `json:"loc"`
	// Type contains the type of this endpoint
	Type OIDURNValue `json:"type"`
}

// Service is a JSON marshalable implementation of an service as specified by RFC006 Distributed Registry.
type Service struct {
	// VendorID contains the ID of the vendor which defined this service.
	VendorID core.PartyID `json:"vid"`
	// Organization contains the ID of the organization which offers this service.
	OrganizationID core.PartyID `json:"oid"`
	// Name contains the name of this service. It must be unique for the vendor/organization combination.
	Name string `json:"name"`
	// Endpoints contains an array of ID of endpoints (of the vendor) that are associated with this service.
	Endpoints []types.EndpointID `json:"eps"`
	// NotBefore contains the date/time indicating from when the service could be used.
	NotBefore time.Time `json:"nbf"`
	// Expiry contains the date/time on or after which the service should not be used anymore. Optional.
	Expiry *time.Time `json:"exp,omitempty"`
}
