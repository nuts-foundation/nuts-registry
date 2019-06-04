package registry

import (
	"context"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/generated"
	"io/ioutil"
	"strconv"
	"time"
)

type HttpClient struct {
	ServerAddress string
	Timeout       time.Duration
}

func (hb HttpClient) EndpointsByOrganization(legalEntity types.LegalEntity) ([]generated.Endpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &generated.EndpointsByOrganisationIdParams{
		OrgIds: []string{legalEntity.URI},
	}
	res, err := (&generated.Client{Server: hb.ServerAddress}).EndpointsByOrganisationId(ctx, params)
	if err != nil {
		log.Error("error while getting endpoints by organization", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("error while reading response body", err)
		return nil, err
	}

	var endpoints []generated.Endpoint

	if err := json.Unmarshal(body, &endpoints); err != nil {
		log.Error("could not unmarshal response body")
		return nil, err
	}

	return endpoints, nil
}

func (hb HttpClient) SearchOrganizations(query string) ([]generated.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	params := &generated.SearchOrganizationsParams{Query: query}
	res, err := (&generated.Client{Server: hb.ServerAddress}).SearchOrganizations(ctx, params)
	if err != nil {
		log.Error("error while searching for organizations", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("error while reading response body", err)
		return nil, err
	}

	var organizations []generated.Organization

	if err := json.Unmarshal(body, &organizations); err != nil {
		log.Error("could not unmarshal response body")
		return nil, err
	}

	for _, org := range organizations {
		// parse the newlines in the public key
		publicKey, _ := strconv.Unquote(`"` + *org.PublicKey + `"`)
		org.PublicKey = &publicKey
	}

	return organizations, nil

}

func (hb HttpClient) OrganizationById(legalEntity types.LegalEntity) (*generated.Organization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hb.Timeout)
	defer cancel()

	res, err := (&generated.Client{Server: hb.ServerAddress}).OrganizationById(ctx, legalEntity.URI)
	if err != nil {
		log.Error("error while getting endpoints by organization", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("error while reading response body", err)
		return nil, err
	}

	var organization generated.Organization
	if err := json.Unmarshal(body, &organization); err != nil {
		log.Error("could not unmarshal response body")
		return nil, err
	}
	// parse the newlines in the public key
	publicKey, _ := strconv.Unquote(`"` + *organization.PublicKey + `"`)
	organization.PublicKey = &publicKey

	return &organization, nil
}
