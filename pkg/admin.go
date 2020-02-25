package pkg

import (
	"github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/types"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
	"time"
)

// RegisterVendor registers a vendor
func (r *Registry) RegisterVendor(id string, name string) (events.Event, error) {
	logrus.Infof("Registering vendor, id=%s, name=%s", id, name)
	return r.publishEvent(events.RegisterVendor, events.RegisterVendorEvent{
		Identifier: events.Identifier(id),
		Name:       name,
	})
}

// VendorClaim registers an organization under a vendor. orgKeys are the organization's keys in JWK format
func (r *Registry) VendorClaim(vendorID string, orgID string, orgName string, orgKeys []interface{}) (events.Event, error) {
	logrus.Infof("Vendor claiming organization, vendor=%s, organization=%s, name=%s, keys=%d",
		vendorID, orgID, orgName, len(orgKeys))

	if orgKeys == nil || len(orgKeys) == 0 {
		logrus.Infof("No keys specified for organisation (id = %s). Keys will be loaded from crypto module.", orgID)
		orgKey, err := r.loadOrGenerateOrganisationKey(orgID)
		if err != nil {
			return nil, err
		}
		orgKeys = append(orgKeys, orgKey)
	}

	return r.publishEvent(events.VendorClaim, events.VendorClaimEvent{
		VendorIdentifier: events.Identifier(vendorID),
		OrgIdentifier:    events.Identifier(orgID),
		OrgName:          orgName,
		OrgKeys:          orgKeys,
		Start:            time.Now(),
	})
}

// RegisterEndpoint registers an endpoint for an organization
func (r *Registry) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, version string) (events.Event, error) {
	logrus.Infof("Registering endpoint, organization=%s, id=%s, type=%s, url=%s, status=%s, version=%s",
		organizationID, id, endpointType, url, status, version)
	return r.publishEvent(events.RegisterEndpoint, events.RegisterEndpointEvent{
		Organization: events.Identifier(organizationID),
		URL:          url,
		EndpointType: endpointType,
		Identifier:   events.Identifier(id),
		Status:       status,
		Version:      version,
	})
}

func (r *Registry) loadOrGenerateOrganisationKey(orgID string) (map[string]interface{}, error) {
	entity := types.LegalEntity{URI: orgID}
	if !r.crypto.KeyExistsFor(entity) {
		logrus.Infof("No keys found for organisation (id = %s), will generate a new key pair.", orgID)
		if err := r.crypto.GenerateKeyPairFor(entity); err != nil {
			return nil, err
		}
	}
	orgKeyAsJwk, err := r.crypto.PublicKeyInJWK(entity)
	if err != nil {
		return nil, err
	}
	return pkg.JwkToMap(orgKeyAsJwk)
}

func (r *Registry) publishEvent(eventType events.EventType, payload interface{}) (events.Event, error) {
	event := events.CreateEvent(eventType, payload)
	if err := r.EventSystem.PublishEvent(event); err != nil {
		return nil, err
	}
	return event, nil
}
