// Package api provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// DIDDocument defines model for DIDDocument.
type DIDDocument map[string]interface{}

// DIDDocumentMetadata defines model for DIDDocumentMetadata.
type DIDDocumentMetadata struct {

	// Date/time at which the document was originally created.
	Created *time.Time `json:"created,omitempty"`

	// Hash (SHA-256, hex-encoded) of DID document bytes. Is equal to payloadHash in network layer.
	Hash *string `json:"hash,omitempty"`

	// Hash (SHA-256, hex-encoded) of the JWS envelope of the first version of the DID document.
	OriginJwsHash *string `json:"originJwsHash,omitempty"`

	// Date/time at which the document (or this version) was updated.
	Updated *time.Time `json:"updated,omitempty"`

	// Semantic version of the DID document.
	Version *int `json:"version,omitempty"`
}

// DIDResolutionResult defines model for DIDResolutionResult.
type DIDResolutionResult struct {

	// The actual DID Document in JSON representation.
	Document         *DIDDocument         `json:"document,omitempty"`
	DocumentMetadata *DIDDocumentMetadata `json:"documentMetadata,omitempty"`

	// Metadata collected during DID Document (a.k.a. DID Resolution Metadata).
	ResolutionMetadata *map[string]interface{} `json:"resolutionMetadata,omitempty"`
}

// SearchDIDParams defines parameters for SearchDID.
type SearchDIDParams struct {

	// URL encoded DID or tag. When given a tag it must resolve to exactly one DID.
	Tags string `json:"tags"`
}

// UpdateDIDJSONBody defines parameters for UpdateDID.
type UpdateDIDJSONBody struct {

	// SHA-256 hash of the last version of the DID Document
	CurrentHash *string `json:"currentHash,omitempty"`

	// The actual DID Document in JSON representation.
	Document *DIDDocument `json:"document,omitempty"`
}

// UpdateDIDTagsJSONBody defines parameters for UpdateDIDTags.
type UpdateDIDTagsJSONBody []string

// UpdateDIDRequestBody defines body for UpdateDID for application/json ContentType.
type UpdateDIDJSONRequestBody UpdateDIDJSONBody

// UpdateDIDTagsRequestBody defines body for UpdateDIDTags for application/json ContentType.
type UpdateDIDTagsJSONRequestBody UpdateDIDTagsJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A callback for modifying requests which are generated before sending over
	// the network.
	RequestEditor RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = http.DefaultClient
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditor = fn
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// SearchDID request
	SearchDID(ctx context.Context, params *SearchDIDParams) (*http.Response, error)

	// CreateDID request
	CreateDID(ctx context.Context) (*http.Response, error)

	// GetDID request
	GetDID(ctx context.Context, didOrTag string) (*http.Response, error)

	// UpdateDID request  with any body
	UpdateDIDWithBody(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*http.Response, error)

	UpdateDID(ctx context.Context, didOrTag string, body UpdateDIDJSONRequestBody) (*http.Response, error)

	// UpdateDIDTags request  with any body
	UpdateDIDTagsWithBody(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*http.Response, error)

	UpdateDIDTags(ctx context.Context, didOrTag string, body UpdateDIDTagsJSONRequestBody) (*http.Response, error)
}

func (c *Client) SearchDID(ctx context.Context, params *SearchDIDParams) (*http.Response, error) {
	req, err := NewSearchDIDRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDID(ctx context.Context) (*http.Response, error) {
	req, err := NewCreateDIDRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetDID(ctx context.Context, didOrTag string) (*http.Response, error) {
	req, err := NewGetDIDRequest(c.Server, didOrTag)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDIDWithBody(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewUpdateDIDRequestWithBody(c.Server, didOrTag, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDID(ctx context.Context, didOrTag string, body UpdateDIDJSONRequestBody) (*http.Response, error) {
	req, err := NewUpdateDIDRequest(c.Server, didOrTag, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDIDTagsWithBody(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewUpdateDIDTagsRequestWithBody(c.Server, didOrTag, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDIDTags(ctx context.Context, didOrTag string, body UpdateDIDTagsJSONRequestBody) (*http.Response, error) {
	req, err := NewUpdateDIDTagsRequest(c.Server, didOrTag, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

// NewSearchDIDRequest generates requests for SearchDID
func NewSearchDIDRequest(server string, params *SearchDIDParams) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/internal/registry/v1/did")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	queryValues := queryUrl.Query()

	if queryFrag, err := runtime.StyleParam("form", true, "tags", params.Tags); err != nil {
		return nil, err
	} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
		return nil, err
	} else {
		for k, v := range parsed {
			for _, v2 := range v {
				queryValues.Add(k, v2)
			}
		}
	}

	queryUrl.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDIDRequest generates requests for CreateDID
func NewCreateDIDRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/internal/registry/v1/did")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDIDRequest generates requests for GetDID
func NewGetDIDRequest(server string, didOrTag string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "didOrTag", didOrTag)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/internal/registry/v1/did/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateDIDRequest calls the generic UpdateDID builder with application/json body
func NewUpdateDIDRequest(server string, didOrTag string, body UpdateDIDJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDIDRequestWithBody(server, didOrTag, "application/json", bodyReader)
}

// NewUpdateDIDRequestWithBody generates requests for UpdateDID with any type of body
func NewUpdateDIDRequestWithBody(server string, didOrTag string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "didOrTag", didOrTag)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/internal/registry/v1/did/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewUpdateDIDTagsRequest calls the generic UpdateDIDTags builder with application/json body
func NewUpdateDIDTagsRequest(server string, didOrTag string, body UpdateDIDTagsJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDIDTagsRequestWithBody(server, didOrTag, "application/json", bodyReader)
}

// NewUpdateDIDTagsRequestWithBody generates requests for UpdateDIDTags with any type of body
func NewUpdateDIDTagsRequestWithBody(server string, didOrTag string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "didOrTag", didOrTag)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/internal/registry/v1/did/%s/tag", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
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

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// SearchDID request
	SearchDIDWithResponse(ctx context.Context, params *SearchDIDParams) (*SearchDIDResponse, error)

	// CreateDID request
	CreateDIDWithResponse(ctx context.Context) (*CreateDIDResponse, error)

	// GetDID request
	GetDIDWithResponse(ctx context.Context, didOrTag string) (*GetDIDResponse, error)

	// UpdateDID request  with any body
	UpdateDIDWithBodyWithResponse(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*UpdateDIDResponse, error)

	UpdateDIDWithResponse(ctx context.Context, didOrTag string, body UpdateDIDJSONRequestBody) (*UpdateDIDResponse, error)

	// UpdateDIDTags request  with any body
	UpdateDIDTagsWithBodyWithResponse(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*UpdateDIDTagsResponse, error)

	UpdateDIDTagsWithResponse(ctx context.Context, didOrTag string, body UpdateDIDTagsJSONRequestBody) (*UpdateDIDTagsResponse, error)
}

type SearchDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]string
}

// Status returns HTTPResponse.Status
func (r SearchDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r SearchDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r CreateDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DIDResolutionResult
}

// Status returns HTTPResponse.Status
func (r GetDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r UpdateDIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDIDTagsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r UpdateDIDTagsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDIDTagsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// SearchDIDWithResponse request returning *SearchDIDResponse
func (c *ClientWithResponses) SearchDIDWithResponse(ctx context.Context, params *SearchDIDParams) (*SearchDIDResponse, error) {
	rsp, err := c.SearchDID(ctx, params)
	if err != nil {
		return nil, err
	}
	return ParseSearchDIDResponse(rsp)
}

// CreateDIDWithResponse request returning *CreateDIDResponse
func (c *ClientWithResponses) CreateDIDWithResponse(ctx context.Context) (*CreateDIDResponse, error) {
	rsp, err := c.CreateDID(ctx)
	if err != nil {
		return nil, err
	}
	return ParseCreateDIDResponse(rsp)
}

// GetDIDWithResponse request returning *GetDIDResponse
func (c *ClientWithResponses) GetDIDWithResponse(ctx context.Context, didOrTag string) (*GetDIDResponse, error) {
	rsp, err := c.GetDID(ctx, didOrTag)
	if err != nil {
		return nil, err
	}
	return ParseGetDIDResponse(rsp)
}

// UpdateDIDWithBodyWithResponse request with arbitrary body returning *UpdateDIDResponse
func (c *ClientWithResponses) UpdateDIDWithBodyWithResponse(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*UpdateDIDResponse, error) {
	rsp, err := c.UpdateDIDWithBody(ctx, didOrTag, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDResponse(rsp)
}

func (c *ClientWithResponses) UpdateDIDWithResponse(ctx context.Context, didOrTag string, body UpdateDIDJSONRequestBody) (*UpdateDIDResponse, error) {
	rsp, err := c.UpdateDID(ctx, didOrTag, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDResponse(rsp)
}

// UpdateDIDTagsWithBodyWithResponse request with arbitrary body returning *UpdateDIDTagsResponse
func (c *ClientWithResponses) UpdateDIDTagsWithBodyWithResponse(ctx context.Context, didOrTag string, contentType string, body io.Reader) (*UpdateDIDTagsResponse, error) {
	rsp, err := c.UpdateDIDTagsWithBody(ctx, didOrTag, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDTagsResponse(rsp)
}

func (c *ClientWithResponses) UpdateDIDTagsWithResponse(ctx context.Context, didOrTag string, body UpdateDIDTagsJSONRequestBody) (*UpdateDIDTagsResponse, error) {
	rsp, err := c.UpdateDIDTags(ctx, didOrTag, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDIDTagsResponse(rsp)
}

// ParseSearchDIDResponse parses an HTTP response from a SearchDIDWithResponse call
func ParseSearchDIDResponse(rsp *http.Response) (*SearchDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &SearchDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDIDResponse parses an HTTP response from a CreateDIDWithResponse call
func ParseCreateDIDResponse(rsp *http.Response) (*CreateDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParseGetDIDResponse parses an HTTP response from a GetDIDWithResponse call
func ParseGetDIDResponse(rsp *http.Response) (*GetDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DIDResolutionResult
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDIDResponse parses an HTTP response from a UpdateDIDWithResponse call
func ParseUpdateDIDResponse(rsp *http.Response) (*UpdateDIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &UpdateDIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ParseUpdateDIDTagsResponse parses an HTTP response from a UpdateDIDTagsWithResponse call
func ParseUpdateDIDTagsResponse(rsp *http.Response) (*UpdateDIDTagsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &UpdateDIDTagsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Searches for Nuts DIDs
	// (GET /internal/registry/v1/did)
	SearchDID(ctx echo.Context, params SearchDIDParams) error
	// Creates a new Nuts DID
	// (POST /internal/registry/v1/did)
	CreateDID(ctx echo.Context) error
	// Resolves a Nuts DID Document
	// (GET /internal/registry/v1/did/{didOrTag})
	GetDID(ctx echo.Context, didOrTag string) error
	// Updates a Nuts DID Document
	// (PUT /internal/registry/v1/did/{didOrTag})
	UpdateDID(ctx echo.Context, didOrTag string) error
	// Replaces the tags of the DID Document.
	// (POST /internal/registry/v1/did/{didOrTag}/tag)
	UpdateDIDTags(ctx echo.Context, didOrTag string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// SearchDID converts echo context to params.
func (w *ServerInterfaceWrapper) SearchDID(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params SearchDIDParams
	// ------------- Required query parameter "tags" -------------

	err = runtime.BindQueryParameter("form", true, true, "tags", ctx.QueryParams(), &params.Tags)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tags: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.SearchDID(ctx, params)
	return err
}

// CreateDID converts echo context to params.
func (w *ServerInterfaceWrapper) CreateDID(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateDID(ctx)
	return err
}

// GetDID converts echo context to params.
func (w *ServerInterfaceWrapper) GetDID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "didOrTag" -------------
	var didOrTag string

	err = runtime.BindStyledParameter("simple", false, "didOrTag", ctx.Param("didOrTag"), &didOrTag)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter didOrTag: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetDID(ctx, didOrTag)
	return err
}

// UpdateDID converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateDID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "didOrTag" -------------
	var didOrTag string

	err = runtime.BindStyledParameter("simple", false, "didOrTag", ctx.Param("didOrTag"), &didOrTag)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter didOrTag: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateDID(ctx, didOrTag)
	return err
}

// UpdateDIDTags converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateDIDTags(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "didOrTag" -------------
	var didOrTag string

	err = runtime.BindStyledParameter("simple", false, "didOrTag", ctx.Param("didOrTag"), &didOrTag)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter didOrTag: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateDIDTags(ctx, didOrTag)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/internal/registry/v1/did", wrapper.SearchDID)
	router.POST(baseURL+"/internal/registry/v1/did", wrapper.CreateDID)
	router.GET(baseURL+"/internal/registry/v1/did/:didOrTag", wrapper.GetDID)
	router.PUT(baseURL+"/internal/registry/v1/did/:didOrTag", wrapper.UpdateDID)
	router.POST(baseURL+"/internal/registry/v1/did/:didOrTag/tag", wrapper.UpdateDIDTags)

}

