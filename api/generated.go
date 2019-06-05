// Package api provides primitives to interact the openapi HTTP API.
//
// This is an autogenerated file, any edits which you make here will be lost!
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// Actor defines component schema for Actor.
type Actor struct {
	Identifier Identifier `json:"identifier"`
}

// Endpoint defines component schema for Endpoint.
type Endpoint struct {
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// EndpointOrganization defines component schema for EndpointOrganization.
type EndpointOrganization struct {
	Endpoint     Identifier `json:"endpoint"`
	Organization Identifier `json:"organization"`
	Status       string     `json:"status"`
}

// Identifier defines component schema for Identifier.
type Identifier string

// Organization defines component schema for Organization.
type Organization struct {
	Actors     []Actor    `json:"actors,omitempty"`
	Identifier Identifier `json:"identifier"`
	Name       string     `json:"name"`
	PublicKey  *string    `json:"publicKey,omitempty"`
}

// Client which conforms to the OpenAPI3 specification for this service. The
// server should be fully qualified with shema and server, ie,
// https://deepmap.com.
type Client struct {
	Server string
	Client http.Client
}

// EndpointsByOrganisationId request
func (c *Client) EndpointsByOrganisationId(ctx context.Context, params *EndpointsByOrganisationIdParams) (*http.Response, error) {
	req, err := NewEndpointsByOrganisationIdRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// OrganizationById request
func (c *Client) OrganizationById(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewOrganizationByIdRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// OrganizationActors request
func (c *Client) OrganizationActors(ctx context.Context, id string, params *OrganizationActorsParams) (*http.Response, error) {
	req, err := NewOrganizationActorsRequest(c.Server, id, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// SearchOrganizations request
func (c *Client) SearchOrganizations(ctx context.Context, params *SearchOrganizationsParams) (*http.Response, error) {
	req, err := NewSearchOrganizationsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// NewEndpointsByOrganisationIdRequest generates requests for EndpointsByOrganisationId
func NewEndpointsByOrganisationIdRequest(server string, params *EndpointsByOrganisationIdParams) (*http.Request, error) {
	var err error

	queryURL := fmt.Sprintf("%s/api/endpoints", server)

	var queryStrings []string

	var queryParam0 string

	queryParam0, err = runtime.StyleParam("form", true, "orgIds", params.OrgIds)
	if err != nil {
		return nil, err
	}

	queryStrings = append(queryStrings, queryParam0)

	var queryParam1 string
	if params.Type != nil {

		queryParam1, err = runtime.StyleParam("form", true, "type", *params.Type)
		if err != nil {
			return nil, err
		}

		queryStrings = append(queryStrings, queryParam1)
	}

	if len(queryStrings) != 0 {
		queryURL += "?" + strings.Join(queryStrings, "&")
	}

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewOrganizationByIdRequest generates requests for OrganizationById
func NewOrganizationByIdRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryURL := fmt.Sprintf("%s/api/organization/%s", server, pathParam0)

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewOrganizationActorsRequest generates requests for OrganizationActors
func NewOrganizationActorsRequest(server string, id string, params *OrganizationActorsParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryURL := fmt.Sprintf("%s/api/organization/%s/actors", server, pathParam0)

	var queryStrings []string

	var queryParam0 string

	queryParam0, err = runtime.StyleParam("form", true, "actorId", params.ActorId)
	if err != nil {
		return nil, err
	}

	queryStrings = append(queryStrings, queryParam0)

	if len(queryStrings) != 0 {
		queryURL += "?" + strings.Join(queryStrings, "&")
	}

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewSearchOrganizationsRequest generates requests for SearchOrganizations
func NewSearchOrganizationsRequest(server string, params *SearchOrganizationsParams) (*http.Request, error) {
	var err error

	queryURL := fmt.Sprintf("%s/api/organizations", server)

	var queryStrings []string

	var queryParam0 string

	queryParam0, err = runtime.StyleParam("form", true, "query", params.Query)
	if err != nil {
		return nil, err
	}

	queryStrings = append(queryStrings, queryParam0)

	if len(queryStrings) != 0 {
		queryURL += "?" + strings.Join(queryStrings, "&")
	}

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// EndpointsByOrganisationIdParams defines parameters for EndpointsByOrganisationId.
type EndpointsByOrganisationIdParams struct {
	OrgIds []string `json:"orgIds"`
	Type   *string  `json:"type,omitempty"`
}

// OrganizationActorsParams defines parameters for OrganizationActors.
type OrganizationActorsParams struct {
	ActorId string `json:"actorId"`
}

// SearchOrganizationsParams defines parameters for SearchOrganizations.
type SearchOrganizationsParams struct {
	Query string `json:"query"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Find endpoints based on organisation identifiers and type of endpoint (optional) (GET /api/endpoints)
	EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error
	// Get organization bij id (GET /api/organization/{id})
	OrganizationById(ctx echo.Context, id string) error
	// get actors for given organization, the main question that is answered by this api: may the professional represent the organization? (GET /api/organization/{id}/actors)
	OrganizationActors(ctx echo.Context, id string, params OrganizationActorsParams) error
	// Search for organizations (GET /api/organizations)
	SearchOrganizations(ctx echo.Context, params SearchOrganizationsParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// EndpointsByOrganisationId converts echo context to params.
func (w *ServerInterfaceWrapper) EndpointsByOrganisationId(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the
	// context.
	var params EndpointsByOrganisationIdParams
	// ------------- Required query parameter "orgIds" -------------
	if paramValue := ctx.QueryParam("orgIds"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument orgIds is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "orgIds", ctx.QueryParams(), &params.OrgIds)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgIds: %s", err))
	}

	// ------------- Optional query parameter "type" -------------
	if paramValue := ctx.QueryParam("type"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "type", ctx.QueryParams(), &params.Type)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.EndpointsByOrganisationId(ctx, params)
	return err
}

// OrganizationById converts echo context to params.
func (w *ServerInterfaceWrapper) OrganizationById(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameter("simple", false, "id", ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OrganizationById(ctx, id)
	return err
}

// OrganizationActors converts echo context to params.
func (w *ServerInterfaceWrapper) OrganizationActors(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameter("simple", false, "id", ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the
	// context.
	var params OrganizationActorsParams
	// ------------- Required query parameter "actorId" -------------
	if paramValue := ctx.QueryParam("actorId"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument actorId is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "actorId", ctx.QueryParams(), &params.ActorId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter actorId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OrganizationActors(ctx, id, params)
	return err
}

// SearchOrganizations converts echo context to params.
func (w *ServerInterfaceWrapper) SearchOrganizations(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the
	// context.
	var params SearchOrganizationsParams
	// ------------- Required query parameter "query" -------------
	if paramValue := ctx.QueryParam("query"); paramValue != "" {

	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query argument query is required, but not found"))
	}

	err = runtime.BindQueryParameter("form", true, true, "query", ctx.QueryParams(), &params.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter query: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SearchOrganizations(ctx, params)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router runtime.EchoRouter, si ServerInterface) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)
	router.GET("/api/organization/:id", wrapper.OrganizationById)
	router.GET("/api/organization/:id/actors", wrapper.OrganizationActors)
	router.GET("/api/organizations", wrapper.SearchOrganizations)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"d7Jpufw7AAD//95tWrT+FAAA",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
