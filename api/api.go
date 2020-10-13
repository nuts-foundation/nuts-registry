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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"

	"github.com/nuts-foundation/nuts-registry/pkg/events"

	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
)

// String converts an identifier to string
func (i Identifier) String() string {
	return string(i)
}

// ApiWrapper is needed to connect the implementation to the echo ServiceWrapper
type ApiWrapper struct {
	R pkg.RegistryClient
}

// To unmarshal an event EventFromJSON must be used (since we expose the interface rather than the internal struct).
// However, since we can't instruct Go to use a particular function to unmarshal list entries, we create a type alias
// for the []events.Event and implement json.Unmarshaler to allow Go to unmarshal the list.
type listOfEvents []events.Event

func (l *listOfEvents) UnmarshalJSON(data []byte) error {
	evts, err := events.EventsFromJSON(data)
	if err != nil {
		return err
	}
	*l = evts
	return nil
}

// altVerifyResponse is alternative, unmarshallable version of VerifyResponse in generated.go for client-side usage.
// Our OpenAPI code generator generates a struct for Event completely separate from our (thoroughly tested) implementation
// events.Event. Using the generator's struct would require elaborate conversion code, which would only be a potential source of bugs.
type altVerifyResponse struct {
	Events listOfEvents
	Fix    bool
}

func (apiResource ApiWrapper) Verify(ctx echo.Context, params VerifyParams) error {
	var fix = false
	if params.Fix != nil {
		fix = *params.Fix
	}
	resultingEvents, needsFixing, err := apiResource.R.Verify(fix)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, altVerifyResponse{Events: resultingEvents, Fix: needsFixing})
}

// DeprecatedVendorClaim is deprecated, use VendorClaim.
func (apiResource ApiWrapper) DeprecatedVendorClaim(ctx echo.Context, _ string) error {
	return apiResource.VendorClaim(ctx)
}

func (apiResource ApiWrapper) RefreshOrganizationCertificate(ctx echo.Context, id string) error {
	partyID := tryParsePartyID(id, ctx)
	if partyID.IsZero() {
		return nil
	}
	event, err := apiResource.R.RefreshOrganizationCertificate(partyID)
	if errors.Is(err, ErrOrganizationNotFound) {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// RegisterEndpoint is the Api implementation for registering an endpoint.
func (apiResource ApiWrapper) RegisterEndpoint(ctx echo.Context, id string) error {
	organizationID := tryParsePartyID(id, ctx)
	if organizationID.IsZero() {
		return nil
	}
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	ep := Endpoint{}
	err = json.Unmarshal(bytes, &ep)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err = ep.validate(); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	event, err := apiResource.R.RegisterEndpoint(organizationID, ep.Identifier.String(), ep.URL, ep.EndpointType, ep.Status, fromEndpointProperties(ep.Properties))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// VendorClaim is the Api implementation for registering a vendor claim.
func (apiResource ApiWrapper) VendorClaim(ctx echo.Context) error {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	org := Organization{}
	err = json.Unmarshal(bytes, &org)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	if err = org.validate(); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	var keys []interface{}
	if org.Keys != nil {
		keys = jwkToMap(*org.Keys)
	}
	organizationID := tryParsePartyID(org.Identifier.String(), ctx)
	if organizationID.IsZero() {
		return nil
	}
	event, err := apiResource.R.VendorClaim(organizationID, org.Name, keys)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, event)
}

// RegisterVendor is the Api implementation for registering a vendor.
func (apiResource ApiWrapper) RegisterVendor(ctx echo.Context) error {
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	if certificate, err := cert.PemToX509(bytes); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	} else {
		event, err := apiResource.R.RegisterVendor(certificate)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, event)
	}
}

// OrganizationById is the Api implementation for getting an organization based on its Id.
func (apiResource ApiWrapper) OrganizationById(ctx echo.Context, id string) error {
	organizationID := tryParsePartyID(id, ctx)
	if organizationID.IsZero() {
		return nil
	}
	result, err := apiResource.R.OrganizationById(organizationID)
	if err != nil {
		logrus.Errorf("Error getting organization %s: %v", organizationID, err)
	}
	if result == nil {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Could not find organization with id %s", organizationID))
	}
	return ctx.JSON(http.StatusOK, Organization{}.fromDb(*result))
}

// VendorById is the Api implementation for getting a vendor based on its Id.
func (apiResource ApiWrapper) VendorById(ctx echo.Context, id string) error {
	vendorID := tryParsePartyID(id, ctx)
	if vendorID.IsZero() {
		return nil
	}
	result, err := apiResource.R.VendorById(vendorID)
	if err != nil {
		logrus.Errorf("Error getting vendor %s: %v", vendorID, err)
	}
	if result == nil {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Could not find vendor with id %s", vendorID))
	}
	return ctx.JSON(http.StatusOK, Vendor{}.fromDB(*result))
}

// EndpointsByOrganisationId is the Api implementation for getting all or certain types of endpoints for an organization
func (apiResource ApiWrapper) EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error {
	foundEPs := []Endpoint{}
	strict := params.Strict
	for _, id := range params.OrgIds {
		organizationID := tryParsePartyID(id, ctx)
		if organizationID.IsZero() {
			return nil
		}
		dbEndpoints, err := apiResource.R.EndpointsByOrganizationAndType(organizationID, params.Type)

		if err != nil {
			logrus.Warning(err.Error())
		} else {
			foundEPs = append(endpointsFromDb(dbEndpoints), foundEPs...)
		}

		if strict != nil && *strict && len(dbEndpoints) == 0 {
			var t = ""
			if params.Type != nil {
				t = *params.Type
			}
			return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("organization with id %s does not have an endpoint of type %s", id, t))
		}
	}

	// filter on type
	filtered := []Endpoint{}
	if params.Type == nil {
		filtered = foundEPs
	} else {
		for _, u := range foundEPs {
			if u.EndpointType == *params.Type {
				filtered = append(filtered, u)
			}
		}
	}

	// generate output
	return ctx.JSON(http.StatusOK, filtered)
}

// SearchOrganizations is the Api implementation for finding organizations by (partial) query
func (apiResource ApiWrapper) SearchOrganizations(ctx echo.Context, params SearchOrganizationsParams) error {

	var (
		searchResult []db.Organization
		org          *db.Organization
		err          error
	)

	if params.Exact != nil && *params.Exact {
		org, err = apiResource.R.ReverseLookup(params.Query)

		if org != nil {
			searchResult = append(searchResult, *org)
		}
	} else {
		searchResult, err = apiResource.R.SearchOrganizations(params.Query)
	}

	if errors.Is(err, db.ErrOrganizationNotFound) {
		return ctx.NoContent(http.StatusNotFound)
	}

	if err != nil {
		return err
	}

	result := make([]Organization, len(searchResult))
	for i, o := range searchResult {
		result[i] = Organization{}.fromDb(o)
	}

	return ctx.JSON(http.StatusOK, result)
}

func (apiResource ApiWrapper) MTLSCAs(ctx echo.Context) error {
	CAs := apiResource.R.VendorCAs()

	acceptHeader := ctx.Request().Header.Get("Accept")
	if "application/json" == acceptHeader {
		result := toCAListWithChain(CAs)
		ctx.JSON(http.StatusOK, &result)
	} else { // otherwise application/x-pem-file
		ctx.Response().Header().Set("Content-Type", "application/x-pem-file")
		ctx.String(http.StatusOK, toSinglePEM(CAs))
	}

	return nil
}

// toSinglePEM transforms the list of chains to a single PEM file removing all duplicate entries
func toSinglePEM(CAs [][]*x509.Certificate) string {
	var depthSet []map[string]bool
	size := 0 // memory optimization
	for _, chain := range CAs {
		for i, cert := range chain {
			if len(depthSet) <= i {
				depthSet = append(depthSet, map[string]bool{})
			}
			currentDepthSet := depthSet[i]
			pem := certificateToPEM(cert)
			if !currentDepthSet[pem] {
				currentDepthSet[pem] = true
				size += len(pem) + 1 // memory optimization (newline)
			}
		}
	}
	var result strings.Builder
	result.Grow(size) // memory optimization
	// now iterate per depth
	for i, _ := range depthSet {
		// reverse order for correct pem ordering
		set := depthSet[len(depthSet)-i-1]
		for pem, _ := range set {
			result.WriteString(pem)
			result.WriteString("\n")
		}
	}
	return result.String()
}

// toCAListWithChain transforms a list of certificate chains to a list of certificates and removes all double root and intermediate entries.
func toCAListWithChain(CAs [][]*x509.Certificate) CAListWithChain {
	var depthSet []map[string]bool // set
	var result CAListWithChain

	for _, chain := range CAs {
		for j, cert := range chain {
			pem := certificateToPEM(cert)
			if j == 0 { // leaf, thus vendor CA
				result.CAList = append(result.CAList, pem)
			} else {
				if len(depthSet) <= j {
					depthSet = append(depthSet, map[string]bool{})
				}
				currentDepthSet := depthSet[j-1]
				if !currentDepthSet[pem] {
					currentDepthSet[pem] = true
				}
			}
		}
	}
	// now iterate per depth for root/intermediates and reverse order
	for i, _ := range depthSet {
		set := depthSet[len(depthSet)-i-1]
		for pem, _ := range set {
			result.Chain = append(result.Chain, pem)
		}
	}

	return result
}

// todo will be available in nuts-crypto after merging nuts-crypto#69
func certificateToPEM(certificate *x509.Certificate) string {
	bytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	})
	return string(bytes)
}

func (apiResource ApiWrapper) MTLSCertificates(ctx echo.Context) error {
	var err error

	return err
}

func tryParsePartyID(id string, ctx echo.Context) core.PartyID {
	unescapedID, err := url.PathUnescape(id)
	if err != nil {
		_ = ctx.String(http.StatusBadRequest, err.Error())
		return core.PartyID{}
	}
	if partyID, err := core.ParsePartyID(unescapedID); err != nil {
		_ = ctx.String(http.StatusBadRequest, err.Error())
		return core.PartyID{}
	} else {
		return partyID
	}
}
