// Package api provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Endpoint defines model for Endpoint.
type Endpoint struct {
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// EndpointOrganization defines model for EndpointOrganization.
type EndpointOrganization struct {
	Endpoint     Identifier `json:"endpoint"`
	Organization Identifier `json:"organization"`
	Status       string     `json:"status"`
}

// Identifier defines model for Identifier.
type Identifier string

// JWK defines model for JWK.
type JWK struct {
	AdditionalProperties map[string]interface{} `json:"-"`
}

// Organization defines model for Organization.
type Organization struct {
	Endpoints  *[]Endpoint `json:"endpoints,omitempty"`
	Identifier Identifier  `json:"identifier"`
	Keys       *[]JWK      `json:"keys,omitempty"`
	Name       string      `json:"name"`
	PublicKey  *string     `json:"publicKey,omitempty"`
}

// EndpointsByOrganisationIdParams defines parameters for EndpointsByOrganisationId.
type EndpointsByOrganisationIdParams struct {
	OrgIds []string `json:"orgIds"`
	Type   *string  `json:"type,omitempty"`
	Strict *bool    `json:"strict,omitempty"`
}

// SearchOrganizationsParams defines parameters for SearchOrganizations.
type SearchOrganizationsParams struct {
	Query string `json:"query"`
	Exact *bool  `json:"exact,omitempty"`
}

// registerOrganizationJSONBody defines parameters for RegisterOrganization.
type registerOrganizationJSONBody Organization

// RegisterOrganizationRequestBody defines body for RegisterOrganization for application/json ContentType.
type RegisterOrganizationJSONRequestBody registerOrganizationJSONBody

// Getter for additional properties for JWK. Returns the specified
// element and whether it was found
func (a JWK) Get(fieldName string) (value interface{}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for JWK
func (a *JWK) Set(fieldName string, value interface{}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]interface{})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for JWK to handle AdditionalProperties
func (a *JWK) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]interface{})
		for fieldName, fieldBuf := range object {
			var fieldVal interface{}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for JWK to handle AdditionalProperties
func (a JWK) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(req *http.Request, ctx context.Context) error

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example.
	Server string

	// HTTP client with any customized settings, such as certificate chains.
	Client http.Client

	// A callback for modifying requests which are generated before sending over
	// the network.
	RequestEditor RequestEditorFn
}

// The interface specification for the client above.
type ClientInterface interface {
	// EndpointsByOrganisationId request
	EndpointsByOrganisationId(ctx context.Context, params *EndpointsByOrganisationIdParams) (*http.Response, error)

	// DeregisterOrganization request
	DeregisterOrganization(ctx context.Context, id string) (*http.Response, error)

	// OrganizationById request
	OrganizationById(ctx context.Context, id string) (*http.Response, error)

	// SearchOrganizations request
	SearchOrganizations(ctx context.Context, params *SearchOrganizationsParams) (*http.Response, error)

	// RegisterOrganization request  with any body
	RegisterOrganizationWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error)

	RegisterOrganization(ctx context.Context, body RegisterOrganizationJSONRequestBody) (*http.Response, error)
}

func (c *Client) EndpointsByOrganisationId(ctx context.Context, params *EndpointsByOrganisationIdParams) (*http.Response, error) {
	req, err := NewEndpointsByOrganisationIdRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeregisterOrganization(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewDeregisterOrganizationRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) OrganizationById(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewOrganizationByIdRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) SearchOrganizations(ctx context.Context, params *SearchOrganizationsParams) (*http.Response, error) {
	req, err := NewSearchOrganizationsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterOrganizationWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewRegisterOrganizationRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterOrganization(ctx context.Context, body RegisterOrganizationJSONRequestBody) (*http.Response, error) {
	req, err := NewRegisterOrganizationRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(req, ctx)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

// NewEndpointsByOrganisationIdRequest generates requests for EndpointsByOrganisationId
func NewEndpointsByOrganisationIdRequest(server string, params *EndpointsByOrganisationIdParams) (*http.Request, error) {
	var err error

	queryUrl := fmt.Sprintf("%s/api/endpoints", server)

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

	var queryParam2 string
	if params.Strict != nil {

		queryParam2, err = runtime.StyleParam("form", true, "strict", *params.Strict)
		if err != nil {
			return nil, err
		}

		queryStrings = append(queryStrings, queryParam2)
	}

	if len(queryStrings) != 0 {
		queryUrl += "?" + strings.Join(queryStrings, "&")
	}

	req, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeregisterOrganizationRequest generates requests for DeregisterOrganization
func NewDeregisterOrganizationRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl := fmt.Sprintf("%s/api/organization/%s", server, pathParam0)

	req, err := http.NewRequest("DELETE", queryUrl, nil)
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

	queryUrl := fmt.Sprintf("%s/api/organization/%s", server, pathParam0)

	req, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewSearchOrganizationsRequest generates requests for SearchOrganizations
func NewSearchOrganizationsRequest(server string, params *SearchOrganizationsParams) (*http.Request, error) {
	var err error

	queryUrl := fmt.Sprintf("%s/api/organizations", server)

	var queryStrings []string

	var queryParam0 string

	queryParam0, err = runtime.StyleParam("form", true, "query", params.Query)
	if err != nil {
		return nil, err
	}

	queryStrings = append(queryStrings, queryParam0)

	var queryParam1 string
	if params.Exact != nil {

		queryParam1, err = runtime.StyleParam("form", true, "exact", *params.Exact)
		if err != nil {
			return nil, err
		}

		queryStrings = append(queryStrings, queryParam1)
	}

	if len(queryStrings) != 0 {
		queryUrl += "?" + strings.Join(queryStrings, "&")
	}

	req, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRegisterOrganizationRequest calls the generic RegisterOrganization builder with application/json body
func NewRegisterOrganizationRequest(server string, body RegisterOrganizationJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewRegisterOrganizationRequestWithBody(server, "application/json", bodyReader)
}

// NewRegisterOrganizationRequestWithBody generates requests for RegisterOrganization with any type of body
func NewRegisterOrganizationRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	queryUrl := fmt.Sprintf("%s/api/organizations", server)

	req, err := http.NewRequest("POST", queryUrl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses returns a ClientWithResponses with a default Client:
func NewClientWithResponses(server string) *ClientWithResponses {
	return &ClientWithResponses{
		ClientInterface: &Client{
			Client: http.Client{},
			Server: server,
		},
	}
}

// NewClientWithResponsesAndRequestEditorFunc takes in a RequestEditorFn callback function and returns a ClientWithResponses with a default Client:
func NewClientWithResponsesAndRequestEditorFunc(server string, reqEditorFn RequestEditorFn) *ClientWithResponses {
	return &ClientWithResponses{
		ClientInterface: &Client{
			Client:        http.Client{},
			Server:        server,
			RequestEditor: reqEditorFn,
		},
	}
}

type endpointsByOrganisationIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r endpointsByOrganisationIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r endpointsByOrganisationIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type deregisterOrganizationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r deregisterOrganizationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r deregisterOrganizationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type organizationByIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r organizationByIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r organizationByIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type searchOrganizationsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r searchOrganizationsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r searchOrganizationsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type registerOrganizationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r registerOrganizationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r registerOrganizationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// EndpointsByOrganisationIdWithResponse request returning *EndpointsByOrganisationIdResponse
func (c *ClientWithResponses) EndpointsByOrganisationIdWithResponse(ctx context.Context, params *EndpointsByOrganisationIdParams) (*endpointsByOrganisationIdResponse, error) {
	rsp, err := c.EndpointsByOrganisationId(ctx, params)
	if err != nil {
		return nil, err
	}
	return ParseendpointsByOrganisationIdResponse(rsp)
}

// DeregisterOrganizationWithResponse request returning *DeregisterOrganizationResponse
func (c *ClientWithResponses) DeregisterOrganizationWithResponse(ctx context.Context, id string) (*deregisterOrganizationResponse, error) {
	rsp, err := c.DeregisterOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParsederegisterOrganizationResponse(rsp)
}

// OrganizationByIdWithResponse request returning *OrganizationByIdResponse
func (c *ClientWithResponses) OrganizationByIdWithResponse(ctx context.Context, id string) (*organizationByIdResponse, error) {
	rsp, err := c.OrganizationById(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseorganizationByIdResponse(rsp)
}

// SearchOrganizationsWithResponse request returning *SearchOrganizationsResponse
func (c *ClientWithResponses) SearchOrganizationsWithResponse(ctx context.Context, params *SearchOrganizationsParams) (*searchOrganizationsResponse, error) {
	rsp, err := c.SearchOrganizations(ctx, params)
	if err != nil {
		return nil, err
	}
	return ParsesearchOrganizationsResponse(rsp)
}

// RegisterOrganizationWithBodyWithResponse request with arbitrary body returning *RegisterOrganizationResponse
func (c *ClientWithResponses) RegisterOrganizationWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*registerOrganizationResponse, error) {
	rsp, err := c.RegisterOrganizationWithBody(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseregisterOrganizationResponse(rsp)
}

func (c *ClientWithResponses) RegisterOrganizationWithResponse(ctx context.Context, body RegisterOrganizationJSONRequestBody) (*registerOrganizationResponse, error) {
	rsp, err := c.RegisterOrganization(ctx, body)
	if err != nil {
		return nil, err
	}
	return ParseregisterOrganizationResponse(rsp)
}

// ParseendpointsByOrganisationIdResponse parses an HTTP response from a EndpointsByOrganisationIdWithResponse call
func ParseendpointsByOrganisationIdResponse(rsp *http.Response) (*endpointsByOrganisationIdResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &endpointsByOrganisationIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParsederegisterOrganizationResponse parses an HTTP response from a DeregisterOrganizationWithResponse call
func ParsederegisterOrganizationResponse(rsp *http.Response) (*deregisterOrganizationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &deregisterOrganizationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParseorganizationByIdResponse parses an HTTP response from a OrganizationByIdWithResponse call
func ParseorganizationByIdResponse(rsp *http.Response) (*organizationByIdResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &organizationByIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParsesearchOrganizationsResponse parses an HTTP response from a SearchOrganizationsWithResponse call
func ParsesearchOrganizationsResponse(rsp *http.Response) (*searchOrganizationsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &searchOrganizationsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParseregisterOrganizationResponse parses an HTTP response from a RegisterOrganizationWithResponse call
func ParseregisterOrganizationResponse(rsp *http.Response) (*registerOrganizationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &registerOrganizationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Find endpoints based on organisation identifiers and type of endpoint (optional)// (GET /api/endpoints)
	EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error
	// Remove organization by id// (DELETE /api/organization/{id})
	DeregisterOrganization(ctx echo.Context, id string) error
	// Get organization by id// (GET /api/organization/{id})
	OrganizationById(ctx echo.Context, id string) error
	// Search for organizations// (GET /api/organizations)
	SearchOrganizations(ctx echo.Context, params SearchOrganizationsParams) error
	// Add an organization to the registry// (POST /api/organizations)
	RegisterOrganization(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// EndpointsByOrganisationId converts echo context to params.
func (w *ServerInterfaceWrapper) EndpointsByOrganisationId(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
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

	// ------------- Optional query parameter "strict" -------------
	if paramValue := ctx.QueryParam("strict"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "strict", ctx.QueryParams(), &params.Strict)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter strict: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.EndpointsByOrganisationId(ctx, params)
	return err
}

// DeregisterOrganization converts echo context to params.
func (w *ServerInterfaceWrapper) DeregisterOrganization(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id string

	err = runtime.BindStyledParameter("simple", false, "id", ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeregisterOrganization(ctx, id)
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

// SearchOrganizations converts echo context to params.
func (w *ServerInterfaceWrapper) SearchOrganizations(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
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

	// ------------- Optional query parameter "exact" -------------
	if paramValue := ctx.QueryParam("exact"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "exact", ctx.QueryParams(), &params.Exact)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter exact: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SearchOrganizations(ctx, params)
	return err
}

// RegisterOrganization converts echo context to params.
func (w *ServerInterfaceWrapper) RegisterOrganization(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RegisterOrganization(ctx)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router runtime.EchoRouter, si ServerInterface) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)
	router.DELETE("/api/organization/:id", wrapper.DeregisterOrganization)
	router.GET("/api/organization/:id", wrapper.OrganizationById)
	router.GET("/api/organizations", wrapper.SearchOrganizations)
	router.POST("/api/organizations", wrapper.RegisterOrganization)

}
