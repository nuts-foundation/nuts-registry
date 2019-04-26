/*
 * Nuts registry
 * Copyright (C) 2019 Nuts community
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
 */

// This is an autogenerated file, any edits which you make here will be lost!
package NutsRegistry

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo"
	"net/http"
	"strings"
)

// Type definition for component schema "Endpoint"
type Endpoint struct {
	URL          string     `json:"URL"`
	EndpointType string     `json:"endpointType"`
	Identifier   Identifier `json:"identifier"`
	Status       string     `json:"status"`
	Version      string     `json:"version"`
}

// Type definition for component schema "Identifier"
type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

// Parameters object for EndpointsByOrganisationId
type EndpointsByOrganisationIdParams struct {
	OrgIds []string `json:"orgIds"`
	Type   *string  `json:"type,omitempty"`
}

type ServerInterface interface {
	// Find endpoints based on organisation identifiers and type of endpoint (optional) (GET /api/endpoints)
	EndpointsByOrganisationId(ctx echo.Context, params EndpointsByOrganisationIdParams) error
}

type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// Wrapper for EndpointsByOrganisationId
func (w *ServerInterfaceWrapper) EndpointsByOrganisationId(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the
	// context.
	var params EndpointsByOrganisationIdParams
	// ------------- Required query parameter "orgIds" -------------

	{
		err = runtime.BindQueryParameter("form", true, true, "orgIds", ctx.QueryParams(), &params.OrgIds)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter orgIds: %s", err))
		}
	}

	// ------------- Optional query parameter "type" -------------

	{
		err = runtime.BindQueryParameter("form", true, false, "type", ctx.QueryParams(), &params.Type)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter type: %s", err))
		}
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.EndpointsByOrganisationId(ctx, params)
	return err
}

func RegisterHandlers(router runtime.EchoRouter, si ServerInterface) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/api/endpoints", wrapper.EndpointsByOrganisationId)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/4RV72sjNxD9VwZdoS3ovHZykHbBH9rQy5kGJ/iST0cIsjReC3YlnTTrO3P4fy+j9fp3",
	"ep9W7Egzb948Pf0Q2jfBO3SURPlDJL3ERuXlP84Ebx3xOkQfMJLFHHme3fPHYNLRBrLeiVLUXitegl8A",
	"LRGUplbVgA5wmwi8yxHrCKNDElLgd9WEGkUpSIeyKEZXN4PhYDgYlaOr6w8SlkQhlUXhWkoDVxd9KiEF",
	"rQOfSxStq8RGij72lAOn8J5nkx4Zn+T1QbI9jtOKt97g53UibHbF33OCd9q7hJeRWIOO7MJiZBy/RFyI",
	"Urwr9kwXW5qLyX7nRopEitp0jr3738M/hO3aRpRfhNJkVyikMDapeY1GvFyAtcKYcsLT/NsAuLaZYzyt",
	"I6FNaIA8GJvIuqq1aQlzpG+IDtpQRWUwnRB6UnwjRcSvrY1oGO/RqI742pGwhyuz3l42UkyOeD3WZMoz",
	"Ou+NO9HeIHQbul4WPgItbYKjynsRtNGVFmlRxoUuRzc3f16a8krVLV4umEOSP9bAN0tL2yn/AMhRvYfx",
	"FI0KEh6ex59Q1bTUKqKE2/H0XsLnp/Ed1gZjrZyRcD++ix5d7SXcTses01fto1GvBldY+9Cgo1e30td/",
	"jIY/ncQOTNfNy4Y3WLfw54399TiBFFDbhd3edKZx9ngLCePKakygVsrWLEBQlBtmdO8jVjZRXAspaqvR",
	"pcyaUw3junu8X10znWQpczFtKUF/BPqiB3ooxXAwGgz5jA/oVLCiFNdsG0KKoGiZ9VCoYHc3Nv+pMFsZ",
	"iybDnxhR7pSY/l4/xEo5m/oY54qqQcKYRPnljA2obSJWvT84dyCoxDfma4txzTwNjiIqIjzPJr+mLA7o",
	"lAyYtAo8JJ6AKEU+LGTPlI/VxPDF2M+PYoty69ncmyVs0mXvy6Z3gOGSorc/VIxqzUo5zfN0wTyB0WAi",
	"NBKwgjw9H+Hjp8nsjT6ou/T7LhaqTkdtnEr2hXenwIabu7saDvmjvSPs3ifC71SEWlm3f8SOCPk/E949",
	"dOcUbOQJBQ//Qg+lm12vgu6u7+Qku0Cj1jBHwCbQmtN/+AnyvSVM/dvCquwK2RffZuwMt3Xax4iaIKGK",
	"etkpM9tBaptGxbUoxUfrDlqAuWKv5Bf9LSDKmXNB/OZzUVX/zhpTVTp0/JQtZvNfAAAA///+TnIZdwgA",
	"AA==",
}

// Returns the Swagger specification corresponding to the generated code
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

