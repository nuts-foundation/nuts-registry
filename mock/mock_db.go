// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/db/db.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	db "github.com/nuts-foundation/nuts-registry/pkg/db"
	events "github.com/nuts-foundation/nuts-registry/pkg/events"
	reflect "reflect"
)

// MockDb is a mock of Db interface
type MockDb struct {
	ctrl     *gomock.Controller
	recorder *MockDbMockRecorder
}

// MockDbMockRecorder is the mock recorder for MockDb
type MockDbMockRecorder struct {
	mock *MockDb
}

// NewMockDb creates a new mock instance
func NewMockDb(ctrl *gomock.Controller) *MockDb {
	mock := &MockDb{ctrl: ctrl}
	mock.recorder = &MockDbMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDb) EXPECT() *MockDbMockRecorder {
	return m.recorder
}

// RegisterEventHandlers mocks base method
func (m *MockDb) RegisterEventHandlers(system events.EventSystem) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterEventHandlers", system)
}

// RegisterEventHandlers indicates an expected call of RegisterEventHandlers
func (mr *MockDbMockRecorder) RegisterEventHandlers(system interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterEventHandlers", reflect.TypeOf((*MockDb)(nil).RegisterEventHandlers), system)
}

// FindEndpointsByOrganizationAndType mocks base method
func (m *MockDb) FindEndpointsByOrganizationAndType(organizationIdentifier string, endpointType *string) ([]db.Endpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindEndpointsByOrganizationAndType", organizationIdentifier, endpointType)
	ret0, _ := ret[0].([]db.Endpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindEndpointsByOrganizationAndType indicates an expected call of FindEndpointsByOrganizationAndType
func (mr *MockDbMockRecorder) FindEndpointsByOrganizationAndType(organizationIdentifier, endpointType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindEndpointsByOrganizationAndType", reflect.TypeOf((*MockDb)(nil).FindEndpointsByOrganizationAndType), organizationIdentifier, endpointType)
}

// SearchOrganizations mocks base method
func (m *MockDb) SearchOrganizations(query string) []db.Organization {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchOrganizations", query)
	ret0, _ := ret[0].([]db.Organization)
	return ret0
}

// SearchOrganizations indicates an expected call of SearchOrganizations
func (mr *MockDbMockRecorder) SearchOrganizations(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchOrganizations", reflect.TypeOf((*MockDb)(nil).SearchOrganizations), query)
}

// OrganizationById mocks base method
func (m *MockDb) OrganizationById(id string) (*db.Organization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OrganizationById", id)
	ret0, _ := ret[0].(*db.Organization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrganizationById indicates an expected call of OrganizationById
func (mr *MockDbMockRecorder) OrganizationById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrganizationById", reflect.TypeOf((*MockDb)(nil).OrganizationById), id)
}

// VendorByID mocks base method
func (m *MockDb) VendorByID(id string) *db.Vendor {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VendorByID", id)
	ret0, _ := ret[0].(*db.Vendor)
	return ret0
}

// VendorByID indicates an expected call of VendorByID
func (mr *MockDbMockRecorder) VendorByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VendorByID", reflect.TypeOf((*MockDb)(nil).VendorByID), id)
}

// ReverseLookup mocks base method
func (m *MockDb) ReverseLookup(name string) (*db.Organization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReverseLookup", name)
	ret0, _ := ret[0].(*db.Organization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReverseLookup indicates an expected call of ReverseLookup
func (mr *MockDbMockRecorder) ReverseLookup(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReverseLookup", reflect.TypeOf((*MockDb)(nil).ReverseLookup), name)
}
