package pkg

import (
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"time"
)

func (r *Registry) RegisterVendor(id string, name string) error {
	logrus.Infof("Registering vendor, id=%s, name=%s", id, name)
	event, err := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{
		Identifier: events.Identifier(id),
		Name:       name,
	})
	if err != nil {
		return err
	}
	return r.EventSystem.PublishEvent(event)
}

func (r *Registry) VendorClaim(vendorId string, orgId string, orgName string, orgKeys []interface{}) error {
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorId, orgId, orgName, len(orgKeys))
	event, err := events.CreateEvent(events.VendorClaim, events.VendorClaimEvent{
		VendorIdentifier: events.Identifier(vendorId),
		OrgIdentifier:    events.Identifier(orgId),
		OrgName:          orgName,
		OrgKeys:          orgKeys,
		Start:            time.Now(),
	})
	if err != nil {
		return err
	}
	return r.EventSystem.PublishEvent(event)
}

func (r *Registry) RegisterEndpoint(organizationId string, id string, url string, endpointType string, status string, version string) error {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s, version=%s",
		organizationId, id, endpointType, url, status, version)
	event, err := events.CreateEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
		Organization: events.Identifier(organizationId),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   events.Identifier(id),
		Status:       status,
		Version:      version,
	})
	if err != nil {
		return err
	}
	return r.EventSystem.PublishEvent(event)
}
