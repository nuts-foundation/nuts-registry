package pkg

import (
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"time"
)

// RegisterVendor registers a vendor
func (r *Registry) RegisterVendor(id string, name string) error {
	logrus.Infof("Registering vendor, id=%s, name=%s", id, name)
	event := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
		Identifier: events.Identifier(id),
		Name:       name,
	})
	return r.EventSystem.PublishEvent(event)
}

// VendorClaim registers an organization under a vendor. orgKeys are the organization's keys in JWK format
func (r *Registry) VendorClaim(vendorID string, orgID string, orgName string, orgKeys []interface{}) error {
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))
	event := events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{
		VendorIdentifier: events.Identifier(vendorID),
		OrgIdentifier:    events.Identifier(orgID),
		OrgName:          orgName,
		OrgKeys:          orgKeys,
		Start:            time.Now(),
	})
	return r.EventSystem.PublishEvent(event)
}

// RegisterEndpoint registers an endpoint for an organization
func (r *Registry) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, version string) error {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s, version=%s",
		organizationID, id, endpointType, url, status, version)
	event := events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
		Organization: events.Identifier(organizationID),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   events.Identifier(id),
		Status:       status,
		Version:      version,
	})
	return r.EventSystem.PublishEvent(event)
}
