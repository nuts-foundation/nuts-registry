package registry

import (
	types "github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/generated"
)

type Client interface {
	EndpointsByOrganization(legalEntity types.LegalEntity) ([]generated.Endpoint, error)
	SearchOrganizations(query string) ([]generated.Organization, error)
	OrganizationById(legalEntity types.LegalEntity) (*generated.Organization, error)
}