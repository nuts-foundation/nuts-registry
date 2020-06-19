package domain

import "github.com/nuts-foundation/nuts-registry/pkg/events"

// RegisterEndpoint event type
const RegisterEndpoint events.EventType = "RegisterEndpointEvent"


// RegisterEndpointEvent event
type RegisterEndpointEvent struct {
	Organization Identifier        `json:"organization"`
	URL          string            `json:"URL"`
	EndpointType string            `json:"endpointType"`
	Identifier   Identifier        `json:"identifier"`
	Status       string            `json:"status"`
	Properties   map[string]string `json:"properties,omitempty"`
}
