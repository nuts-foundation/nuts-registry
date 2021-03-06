/*
 * Nuts registry
 * Copyright (C) 2020. Nuts community
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
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nuts-foundation/nuts-registry/logging"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nuts-foundation/nuts-crypto/pkg/cert"

	"github.com/nuts-foundation/nuts-registry/pkg/events"

	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
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

func (hb HttpClient) Verify(fix bool) ([]events.Event, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	response, err := hb.client().Verify(ctx, &VerifyParams{Fix: &fix})
	if err != nil {
		logging.Log().Error("Error while running verify: ", err)
		return nil, false, err
	}
	if err := testResponseCode(http.StatusOK, response); err != nil {
		return nil, false, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logging.Log().Error("Error while parsing verify response: ", err)
		return nil, false, err
	}
	verifyResponse := altVerifyResponse{}
	if err := json.Unmarshal(responseData, &verifyResponse); err != nil {
		logging.Log().Error("Error while unmarshalling verify response: ", err)
		return nil, false, err
	}
	return verifyResponse.Events, verifyResponse.Fix, nil
}

func (hb HttpClient) RefreshOrganizationCertificate(organizationID core.PartyID) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	response, err := hb.client().RefreshOrganizationCertificate(ctx, organizationID.String())
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(response)
}

// EndpointsByOrganization is the client Api implementation for getting all or certain types of endpoints for an organization
func (hb HttpClient) EndpointsByOrganizationAndType(organizationID core.PartyID, endpointType *string) ([]db.Endpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &EndpointsByOrganisationIdParams{
		OrgIds: []string{organizationID.String()},
		Type:   endpointType,
	}
	res, err := hb.client().EndpointsByOrganisationId(ctx, params)
	if err != nil {
		logging.Log().Error("error while getting endpoints by organization", err)
		return nil, core.Wrap(err)
	}

	parsed, err := ParseEndpointsByOrganisationIdResponse(res)
	if err != nil {
		logging.Log().Error("error while reading response body", err)
		return nil, err
	}

	var endpoints []Endpoint
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(parsed.Body, &endpoints); err != nil {
		logging.Log().Error("could not unmarshal response body")
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
		logging.Log().Error("Reverse lookup returned more than 1 match")
		return nil, ErrOrganizationNotFound
	}

	return &orgs[0], nil
}

func (hb HttpClient) searchOrganization(params SearchOrganizationsParams) ([]db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().SearchOrganizations(ctx, &params)
	if err != nil {
		logging.Log().Error("error while searching for organizations", err)
		return nil, core.Wrap(err)
	}

	parsed, err := ParseSearchOrganizationsResponse(res)
	if err != nil {
		logging.Log().Error("error while reading response body", err)
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
		logging.Log().Error("could not unmarshal response body")
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
func (hb HttpClient) RegisterEndpoint(organizationID core.PartyID, id string, url string, endpointType string, status string, properties map[string]string) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().RegisterEndpoint(ctx, organizationID.String(), RegisterEndpointJSONRequestBody{
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
func (hb HttpClient) VendorClaim(orgID core.PartyID, orgName string, orgKeys []interface{}) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	var keys = make([]JWK, 0)
	if orgKeys != nil {
		for _, key := range orgKeys {
			keys = append(keys, key.(map[string]interface{}))
		}
	}
	res, err := hb.client().VendorClaim(ctx, VendorClaimJSONRequestBody{
		Identifier: Identifier(orgID.String()),
		Keys:       &keys,
		Name:       orgName,
	})
	if err != nil {
		return nil, err
	}
	return testAndParseEventResponse(res)
}

// RegisterVendor is the client Api implementation for registering a vendor.
func (hb HttpClient) RegisterVendor(certificate *x509.Certificate) (events.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()
	res, err := hb.client().RegisterVendorWithBody(ctx, "application/x-pem-file", strings.NewReader(cert.CertificateToPEM(certificate)))
	if err != nil {
		logging.Log().Error("error while registering vendor", err)
		return nil, core.Wrap(err)
	}
	return testAndParseEventResponse(res)
}

// OrganizationById is the client Api implementation for getting an organization based on its Id.
func (hb HttpClient) OrganizationById(id core.PartyID) (*db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().OrganizationById(ctx, id.String())
	if err != nil {
		logging.Log().Error("error while getting organization by id", err)
		return nil, core.Wrap(err)
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	parsed, err := ParseOrganizationByIdResponse(res)
	if err != nil {
		logging.Log().Error("error while reading response body", err)
		return nil, err
	}

	var organization Organization
	if err := json.Unmarshal(parsed.Body, &organization); err != nil {
		logging.Log().Errorf("could not unmarshal response body: %v", err)
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

// VendorById is the client Api implementation for getting a vendor based on its Id.
func (hb HttpClient) VendorById(id core.PartyID) (*db.Vendor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().VendorById(ctx, id.String())
	if err != nil {
		logging.Log().Error("error while getting vendor by id", err)
		return nil, core.Wrap(err)
	}
	if err := testResponseCode(http.StatusOK, res); err != nil {
		return nil, err
	}

	parsed, err := ParseVendorByIdResponse(res)
	if err != nil {
		logging.Log().Error("error while reading response body", err)
		return nil, err
	}

	var vendor Vendor
	if err := json.Unmarshal(parsed.Body, &vendor); err != nil {
		logging.Log().Errorf("could not unmarshal response body: %v", err)
		return nil, err
	}

	o := vendor.toDb()
	return &o, nil
}

// VendorCAs on the client is not implemented
func (hb HttpClient) VendorCAs() [][]*x509.Certificate {
	return [][]*x509.Certificate{}
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
