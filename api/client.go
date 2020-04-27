/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nuts-foundation/nuts-registry/pkg/events"

	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
)

// HttpClient holds the server address and other basic settings for the http client
type HttpClient struct {
	ServerAddress string
	Timeout       time.Duration
}

func (hb HttpClient) client() ClientInterface {
	url := hb.ServerAddress
	if !strings.Contains(url, "http") {
		url = fmt.Sprintf("http://%v", hb.ServerAddress)
	}

	response, err := NewClientWithResponses(url)
	if err != nil {
		panic(err)
	}
	return response
}

func (hb HttpClient) RefreshVendorCertificate() (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	response, err := hb.client().RefreshVendorCertificate(ctx)
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(response)
}

func (hb HttpClient) RefreshOrganizationCertificate(organizationID string) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	response, err := hb.client().RefreshOrganizationCertificate(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(response)
}

// EndpointsByOrganization is the client Api implementation for getting all or certain types of endpoints for an organization
func (hb HttpClient) EndpointsByOrganizationAndType(legalEntity string, endpointType *string) ([]db.Endpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &EndpointsByOrganisationIdParams{
		OrgIds: []string{legalEntity},
		Type:   endpointType,
	}
	res, err := hb.client().EndpointsByOrganisationId(ctx, params)
	if err != nil {
		logrus.Error("error while getting endpoints by organization", err)
		return nil, core.Wrap(err)
	}

	parsed, err := ParseEndpointsByOrganisationIdResponse(res)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	var endpoints []Endpoint
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(parsed.Body, &endpoints); err != nil {
		logrus.Error("could not unmarshal response body")
		return nil, err
	}

	return endpointsToDb(endpoints), nil
}

// SearchOrganizations is the client Api implementation for finding organizations by (partial) query
func (hb HttpClient) SearchOrganizations(query string) ([]db.Organization, error) {
	params := SearchOrganizationsParams{Query: query}

	return hb.searchOrganization(params)
}

// ErrOrganizationNotFound is returned by the reverseLookup when the organization is not found
var ErrOrganizationNotFound = errors.New("organization not found")

// ReverseLookup returns an exact match or an error
func (hb HttpClient) ReverseLookup(name string) (*db.Organization, error) {
	t := true
	params := SearchOrganizationsParams{Query: name, Exact: &t}

	orgs, err := hb.searchOrganization(params)
	if err != nil {
		return nil, err
	}

	// should not be reachable
	if len(orgs) != 1 {
		logrus.Error("Reverse lookup returned more than 1 match")
		return nil, ErrOrganizationNotFound
	}

	return &orgs[0], nil
}

func (hb HttpClient) searchOrganization(params SearchOrganizationsParams) ([]db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().SearchOrganizations(ctx, &params)
	if err != nil {
		logrus.Error("error while searching for organizations", err)
		return nil, core.Wrap(err)
	}

	parsed, err := ParseSearchOrganizationsResponse(res)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	if parsed.StatusCode() == http.StatusNotFound {
		return nil, ErrOrganizationNotFound
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	var organizations []Organization

	if err := json.Unmarshal(parsed.Body, &organizations); err != nil {
		logrus.Error("could not unmarshal response body")
		return nil, err
	}

	for _, org := range organizations {
		// parse the newlines in the public key
		if org.PublicKey != nil {
			publicKey, _ := strconv.Unquote(`"` + *org.PublicKey + `"`)
			org.PublicKey = &publicKey
		}
	}

	return organizationsToDb(organizations), nil
}

// RegisterEndpoint is the client Api implementation for registering an endpoint for an organisation.
func (hb HttpClient) RegisterEndpoint(organizationID string, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().RegisterEndpoint(ctx, organizationID, RegisterEndpointJSONRequestBody{
		URL:          url,
		EndpointType: endpointType,
		Identifier:   Identifier(id),
		Status:       status,
		Properties:   toEndpointProperties(properties),
	})
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(res)
}

// VendorClaim is the client Api implementation for registering an organisation.
func (hb HttpClient) VendorClaim(orgID string, orgName string, orgKeys []interface{}) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	var keys = make([]JWK, 0)
	if orgKeys != nil {
		for _, key := range orgKeys {
			keys = append(keys, JWK{AdditionalProperties: key.(map[string]interface{})})
		}
	}
	res, err := hb.client().VendorClaim(ctx, VendorClaimJSONRequestBody{
		Identifier: Identifier(orgID),
		Keys:       &keys,
		Name:       orgName,
	})
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(res)
}

// RegisterVendor is the client Api implementation for registering a vendor.
func (hb HttpClient) RegisterVendor(name string, domain string) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().RegisterVendor(ctx, RegisterVendorJSONRequestBody{
		Name:       name,
		Domain:     Domain(domain),
	})
	if err != nil {
		logrus.Error("error while registering vendor", err)
		return nil, core.Wrap(err)
	}
	return testAndParseEventResponse(res)
}

// OrganizationById is the client Api implementation for getting an organization based on its Id.
func (hb HttpClient) OrganizationById(legalEntity string) (*db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().OrganizationById(ctx, legalEntity)
	if err != nil {
		logrus.Error("error while getting endpoints by organization", err)
		return nil, core.Wrap(err)
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	parsed, err := ParseOrganizationByIdResponse(res)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	var organization Organization
	if err := json.Unmarshal(parsed.Body, &organization); err != nil {
		logrus.Errorf("could not unmarshal response body: %v", err)
		return nil, err
	}
	// parse the newlines in the public key
	if organization.PublicKey != nil {
		publicKey, _ := strconv.Unquote(`"` + *organization.PublicKey + `"`)
		organization.PublicKey = &publicKey
	}

	o := organization.toDb()
	return &o, nil
}

func testResponseCode(expectedStatusCode int, response *http.Response) error {
	if response.StatusCode != expectedStatusCode {
		responseData, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("registry returned HTTP %d (expected: %d), response: %s",
			response.StatusCode, expectedStatusCode, string(responseData))
	}
	return nil
}

func testAndParseEventResponse(response *http.Response) (events.Event, error) {
	if err := testResponseCode(http.StatusOK, response); err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return events.EventFromJSON(responseData)
}
