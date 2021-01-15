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
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-registry/mock"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var genericError = errors.New("generic error occurred")

func newRequest(didStore pkg.DIDService, method string, path string, body []byte) (*ServerInterfaceWrapper, echo.Context, *httptest.ResponseRecorder) {
	server := echo.New()
	stub := ApiWrapper{R: &pkg.Registry{DIDStore: didStore}}
	wrapper := &ServerInterfaceWrapper{
		Handler: stub,
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, path, bodyReader)
	recorder := httptest.NewRecorder()
	callContext := server.NewContext(request, recorder)
	callContext.SetPath(path)

	return wrapper, callContext, recorder
}

func TestApiResource_CreateDID(t *testing.T) {
	const path = "/internal/registry/v1/did/"
	t.Run("200", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		didStore := mock.NewMockDIDStore(ctrl)
		didStore.EXPECT().Create().Return(&did.Document{}, nil)
		wrapper, request, recorder := newRequest(didStore, echo.POST, path, nil)
		err := wrapper.CreateDID(request)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, `{"context":null,"id":""}`, recorder.Body.String())
	})
	t.Run("500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		didStore := mock.NewMockDIDStore(ctrl)
		didStore.EXPECT().Create().Return(nil, genericError)
		wrapper, request, _ := newRequest(didStore, echo.POST, path, nil)
		err := wrapper.CreateDID(request)
		assert.Error(t, err)
	})
}

func TestApiResource_GetDID(t *testing.T) {
	const path = "/internal/registry/v1/did/{didOrTag}"
	t.Run("by DID", func(t *testing.T) {
		t.Run("200", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			didStore := mock.NewMockDIDStore(ctrl)
			expectedDID, _ := did.ParseDID("did:nuts:1234")
			didStore.EXPECT().Get(*expectedDID).Return(&did.Document{ID: *expectedDID}, &pkg.DIDDocumentMetadata{}, nil)
			wrapper, request, recorder := newRequest(didStore, echo.POST, path, nil)
			request.SetParamNames("didOrTag")
			request.SetParamValues(expectedDID.String())
			err := wrapper.GetDID(request)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, http.StatusOK, recorder.Code)
		})
		t.Run("500", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			didStore := mock.NewMockDIDStore(ctrl)
			didStore.EXPECT().Create().Return(nil, genericError)
			wrapper, request, _ := newRequest(didStore, echo.POST, path, nil)
			err := wrapper.CreateDID(request)
			assert.Error(t, err)
		})
	})
}