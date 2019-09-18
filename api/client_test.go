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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// RoundTripFunc
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn RoundTripFunc) HttpClient {
	return HttpClient{
		ServerAddress: "http://localhost:1323",
		customClient: &http.Client{
			Transport: RoundTripFunc(fn),
		},
	}
}

func TestHttpClient_OrganizationById(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString("error reason")),
				Header:     make(http.Header),
			}
		})

		_, err := client.OrganizationById("id")

		expected := "-: Registry returned 404, reason: error reason"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("200", func(t *testing.T) {
		org, _ := json.Marshal(organizations[0])
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(org)),
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			}
		})

		res, err := client.OrganizationById("id")

		if err != nil {
			t.Errorf("Expected no error, got [%s]", err.Error())
		}

		if res.Identifier != organizations[0].Identifier {
			t.Errorf("Expected return organization identifier to be [%s], got [%s]", organizations[0].Identifier, res.Identifier)
		}
	})
}

func TestHttpClient_RegisterOrganization(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewBufferString("error reason")),
				Header:     make(http.Header),
			}
		})

		err := client.RegisterOrganization(organizations[0])

		expected := "-: Registry returned 400, reason: error reason"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("201", func(t *testing.T) {
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 201,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				Header:     make(http.Header),
			}
		})

		err := client.RegisterOrganization(organizations[0])

		if err != nil {
			t.Errorf("Expected no error, got [%s]", err.Error())
		}
	})
}

func TestHttpClient_RemoveOrganization(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString("error reason")),
				Header:     make(http.Header),
			}
		})

		err := client.RemoveOrganization("id")

		expected := "-: Registry returned 404, reason: error reason"
		if err.Error() != expected {
			t.Errorf("Expected error [%s], got [%s]", expected, err.Error())
		}
	})

	t.Run("202", func(t *testing.T) {
		client := newTestClient(func(req *http.Request) *http.Response {
			// Test request parameters
			return &http.Response{
				StatusCode: 202,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				Header:     make(http.Header),
			}
		})

		err := client.RemoveOrganization("id")

		if err != nil {
			t.Errorf("Expected no error, got [%s]", err.Error())
		}
	})
}
