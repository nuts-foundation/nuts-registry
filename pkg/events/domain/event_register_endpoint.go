package domain

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
)

// RegisterEndpoint event type
const RegisterEndpoint events.EventType = "RegisterEndpointEvent"

// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Organization core.PartyID      `json:"organization"`
	URL          string            `json:"URL"`
	EndpointType string            `json:"endpointType"`
	Identifier   types.EndpointID  `json:"identifier"`
	Status       string            `json:"status"`
	Properties   map[string]string `json:"properties,omitempty"`
}
