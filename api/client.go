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
	"fmt"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
	"go/types"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// HttpClient holds the server address and other basic settings for the http client
type HttpClient struct {
	ServerAddress string
	Timeout       time.Duration
	customClient  *http.Client
}

func (hb HttpClient) client() *Client {
	if hb.customClient != nil {
		return &Client{
			Server: fmt.Sprintf("http://%v", hb.ServerAddress),
			Client: *hb.customClient,
		}
	}

	return &Client{
		Server: fmt.Sprintf("http://%v", hb.ServerAddress),
	}
}

// RemoveOrganization removes an organization and its endpoints from the registry
func (hb HttpClient) RemoveOrganization(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	result, err := hb.client().DeregisterOrganization(ctx, id)
	if err != nil {
		logrus.Error("error while removing organization from registry", err)
		return err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return err
	}

	if result.StatusCode != http.StatusAccepted {
		err = types.Error{Msg: fmt.Sprintf("Registry returned %d, reason: %s", result.StatusCode, body)}
		logrus.Error(err.Error())
		return err
	}

	return nil
}

// RegisterOrganization adds an organization to the registry
func (hb HttpClient) RegisterOrganization(org db.Organization) error {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	e := endpointsArrayFromDb(org.Endpoints)
	req := RegisterOrganizationJSONRequestBody{
		Endpoints:  &e,
		Identifier: Identifier(string(org.Identifier)),
		Name:       org.Name,
		PublicKey:  org.PublicKey,
	}
	result, err := hb.client().RegisterOrganization(ctx, req)
	if err != nil {
		logrus.Error("error while registering organization in registry", err)
		return err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return err
	}

	if result.StatusCode != http.StatusCreated {
		err = types.Error{Msg: fmt.Sprintf("Registry returned %d, reason: %s", result.StatusCode, body)}
		logrus.Error(err.Error())
		return err
	}

	return nil
}

// EndpointsByOrganization is the client Api implementation for getting all or certain types of endpoints for an organization
func (hb HttpClient) EndpointsByOrganization(legalEntity string) ([]db.Endpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &EndpointsByOrganisationIdParams{
		OrgIds: []string{legalEntity},
	}
	res, err := hb.client().EndpointsByOrganisationId(ctx, params)
	if err != nil {
		logrus.Error("error while getting endpoints by organization", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	var endpoints []Endpoint

	if err := json.Unmarshal(body, &endpoints); err != nil {
		logrus.Error("could not unmarshal response body")
		return nil, err
	}

	return endpointsArrayToDb(endpoints), nil
}

// SearchOrganizations is the client Api implementation for finding organizations by (partial) query
func (hb HttpClient) SearchOrganizations(query string) ([]db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &SearchOrganizationsParams{Query: query}
	res, err := hb.client().SearchOrganizations(ctx, params)
	if err != nil {
		logrus.Error("error while searching for organizations", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	var organizations []Organization

	if err := json.Unmarshal(body, &organizations); err != nil {
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

	return organizationsToFromDb(organizations), nil

}

// OrganizationById is the client Api implementation for getting an organization based on its Id.
func (hb HttpClient) OrganizationById(legalEntity string) (*db.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := hb.client().OrganizationById(ctx, legalEntity)
	if err != nil {
		logrus.Error("error while getting endpoints by organization", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error("error while reading response body", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		err = types.Error{Msg: fmt.Sprintf("Registry returned %d, reason: %s", res.StatusCode, body)}
		logrus.Error(err.Error())
		return nil, err
	}

	var organization Organization
	if err := json.Unmarshal(body, &organization); err != nil {
		logrus.Error("could not unmarshal response body")
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
