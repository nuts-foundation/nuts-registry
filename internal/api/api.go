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
	"encoding/json"
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"net/http"
	"strings"

	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/pkg"
)

const jsonDIDMimeType = "application/json+did-document"
const tagPrefix = "tag:"

// ApiWrapper is needed to connect the implementation to the echo ServiceWrapper
type ApiWrapper struct {
	R pkg.RegistryClient
}

func (apiResource ApiWrapper) SearchDID(ctx echo.Context, params SearchDIDParams) error {
	// TODO: support OnlyOwn
	// TODO: Support multiple tags passed in query string
	// TODO: Allow 'expand' parameter to be passed, to return the complete DID Document instead of just the DID?
	if searchResults, err := apiResource.R.Search(false, []string{params.Tags}); err != nil {
		return err
	} else {
		response := make([]string, len(searchResults))
		for i, searchResult := range searchResults {
			response[i] = searchResult.ID.String()
		}
		return ctx.JSON(http.StatusOK, response)
	}
}

func (apiResource ApiWrapper) CreateDID(ctx echo.Context) error {
	if createdDID, err := apiResource.R.Create(); err != nil {
		return err
	} else {
		ctx.Set(echo.HeaderContentType, jsonDIDMimeType)
		data, _ := json.Marshal(*createdDID)
		return ctx.String(http.StatusCreated, string(data))
	}
}

func (apiResource ApiWrapper) GetDID(ctx echo.Context, didOrTag string) error {
	if strings.HasPrefix(didOrTag, tagPrefix) {
		if document, metadata, err := apiResource.R.GetByTag(strings.TrimPrefix(didOrTag, tagPrefix)); err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, (&DIDResolutionResult{}).FromModel(*document, *metadata))
		}
	} else {
		if actualDID, err := did.ParseDID(didOrTag); err != nil {
			return ctx.String(http.StatusBadRequest, err.Error())
		} else if document, metadata, err := apiResource.R.Get(*actualDID); err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, (&DIDResolutionResult{}).FromModel(*document, *metadata))
		}
	}
}

func (apiResource ApiWrapper) UpdateDID(ctx echo.Context, didToUpdate string) error {
	request := UpdateDIDJSONBody{}
	if err := unmarshalRequestBody(ctx, &request); err != nil {
		return err
	}
	if document, err := request.Document.ToModel(); err != nil {
		return err
	} else if currentHash, err := model.ParseHash(request.CurrentHash); err != nil {
		return err
	} else if actualDID, err := did.ParseDID(didToUpdate); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	} else if updatedDocument, err := apiResource.R.Update(*actualDID, currentHash, *document); err != nil {
		return err
	} else {
		ctx.Set(echo.HeaderContentType, jsonDIDMimeType)
		data, _ := json.Marshal(*updatedDocument)
		return ctx.String(http.StatusCreated, string(data))
	}
}

func (apiResource ApiWrapper) UpdateDIDTags(ctx echo.Context, didToTag string) error {
	newTags := make([]string, 2) // Guessed average number of tags
	if err := unmarshalRequestBody(ctx, &newTags); err != nil {
		return err
	}
	if actualDID, err := did.ParseDID(didToTag); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	} else if err := apiResource.R.Tag(*actualDID, newTags); err != nil {
		ctx.Error(err)
		return nil
	} else {
		return ctx.NoContent(http.StatusNoContent)
	}
}

func unmarshalRequestBody(ctx echo.Context, target interface{}) error {
	if bodyAsBytes, err := ioutil.ReadAll(ctx.Request().Body); err != nil {
		return err
	} else if err := json.Unmarshal(bodyAsBytes, target); err != nil {
		return err
	}
	return nil
}
