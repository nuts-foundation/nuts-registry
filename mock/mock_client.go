// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/registry.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	db "github.com/nuts-foundation/nuts-registry/pkg/db"
	events "github.com/nuts-foundation/nuts-registry/pkg/events"
	reflect "reflect"
)

// MockRegistryClient is a mock of RegistryClient interface
type MockRegistryClient struct {
	ctrl     *gomock.Controller
	recorder *MockRegistryClientMockRecorder
}

// MockRegistryClientMockRecorder is the mock recorder for MockRegistryClient
type MockRegistryClientMockRecorder struct {
	mock *MockRegistryClient
}

// NewMockRegistryClient creates a new mock instance
func NewMockRegistryClient(ctrl *gomock.Controller) *MockRegistryClient {
	mock := &MockRegistryClient{ctrl: ctrl}
	mock.recorder = &MockRegistryClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRegistryClient) EXPECT() *MockRegistryClientMockRecorder {
	return m.recorder
}

// EndpointsByOrganizationAndType mocks base method
func (m *MockRegistryClient) EndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]db.Endpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EndpointsByOrganizationAndType", organizationIdentifier, endpointType)
	ret0, _ := ret[0].([]db.Endpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EndpointsByOrganizationAndType indicates an expected call of EndpointsByOrganizationAndType
func (mr *MockRegistryClientMockRecorder) EndpointsByOrganizationAndType(organizationIdentifier, endpointType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndpointsByOrganizationAndType", reflect.TypeOf((*MockRegistryClient)(nil).EndpointsByOrganizationAndType), organizationIdentifier, endpointType)
}

// SearchOrganizations mocks base method
func (m *MockRegistryClient) SearchOrganizations(query string) ([]db.Organization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchOrganizations", query)
	ret0, _ := ret[0].([]db.Organization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchOrganizations indicates an expected call of SearchOrganizations
func (mr *MockRegistryClientMockRecorder) SearchOrganizations(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchOrganizations", reflect.TypeOf((*MockRegistryClient)(nil).SearchOrganizations), query)
}

// OrganizationById mocks base method
func (m *MockRegistryClient) OrganizationById(id string) (*db.Organization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrganizationById", id)
	ret0, _ := ret[0].(*db.Organization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrganizationById indicates an expected call of OrganizationById
func (mr *MockRegistryClientMockRecorder) OrganizationById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrganizationById", reflect.TypeOf((*MockRegistryClient)(nil).OrganizationById), id)
}

// ReverseLookup mocks base method
func (m *MockRegistryClient) ReverseLookup(name string) (*db.Organization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReverseLookup", name)
	ret0, _ := ret[0].(*db.Organization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReverseLookup indicates an expected call of ReverseLookup
func (mr *MockRegistryClientMockRecorder) ReverseLookup(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReverseLookup", reflect.TypeOf((*MockRegistryClient)(nil).ReverseLookup), name)
}

// RegisterEndpoint mocks base method
func (m *MockRegistryClient) RegisterEndpoint(organizationID, id, url, endpointType, status string, properties map[string]string) (events.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterEndpoint", organizationID, id, url, endpointType, status, properties)
	ret0, _ := ret[0].(events.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterEndpoint indicates an expected call of RegisterEndpoint
func (mr *MockRegistryClientMockRecorder) RegisterEndpoint(organizationID, id, url, endpointType, status, properties interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterEndpoint", reflect.TypeOf((*MockRegistryClient)(nil).RegisterEndpoint), organizationID, id, url, endpointType, status, properties)
}

// VendorClaim mocks base method
func (m *MockRegistryClient) VendorClaim(orgID, orgName string, orgKeys []interface{}) (events.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VendorClaim", orgID, orgName, orgKeys)
	ret0, _ := ret[0].(events.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VendorClaim indicates an expected call of VendorClaim
func (mr *MockRegistryClientMockRecorder) VendorClaim(orgID, orgName, orgKeys interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VendorClaim", reflect.TypeOf((*MockRegistryClient)(nil).VendorClaim), orgID, orgName, orgKeys)
}

// RegisterVendor mocks base method
func (m *MockRegistryClient) RegisterVendor(name, domain string) (events.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterVendor", name, domain)
	ret0, _ := ret[0].(events.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterVendor indicates an expected call of RegisterVendor
func (mr *MockRegistryClientMockRecorder) RegisterVendor(name, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterVendor", reflect.TypeOf((*MockRegistryClient)(nil).RegisterVendor), name, domain)
}
