package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"time"
)

// HttpClient holds the server address and other basic settings for the http client
type HttpClient struct {
	ServerAddress string
	Timeout       time.Duration
}

func (hb HttpClient) client() *Client {
	return &Client{
		Server: fmt.Sprintf("http://%v", hb.ServerAddress),
	}
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
