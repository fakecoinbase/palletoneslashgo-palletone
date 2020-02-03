// Code generated by MockGen. DO NOT EDIT.
// Source: ./dag/rwset/dependency.go

// Package rwset is a generated GoMock package.
package rwset

import (
	gomock "github.com/golang/mock/gomock"
	common "github.com/palletone/go-palletone/common"
	modules "github.com/palletone/go-palletone/dag/modules"
	reflect "reflect"
)

// MockIDataQuery is a mock of IDataQuery interface
type MockIDataQuery struct {
	ctrl     *gomock.Controller
	recorder *MockIDataQueryMockRecorder
}

// MockIDataQueryMockRecorder is the mock recorder for MockIDataQuery
type MockIDataQueryMockRecorder struct {
	mock *MockIDataQuery
}

// NewMockIDataQuery creates a new mock instance
func NewMockIDataQuery(ctrl *gomock.Controller) *MockIDataQuery {
	mock := &MockIDataQuery{ctrl: ctrl}
	mock.recorder = &MockIDataQueryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIDataQuery) EXPECT() *MockIDataQueryMockRecorder {
	return m.recorder
}

// GetContractStatesById mocks base method
func (m *MockIDataQuery) GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractStatesById", contractid)
	ret0, _ := ret[0].(map[string]*modules.ContractStateValue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractStatesById indicates an expected call of GetContractStatesById
func (mr *MockIDataQueryMockRecorder) GetContractStatesById(contractid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractStatesById", reflect.TypeOf((*MockIDataQuery)(nil).GetContractStatesById), contractid)
}

// GetContractState mocks base method
func (m *MockIDataQuery) GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractState", contractid, field)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(*modules.StateVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetContractState indicates an expected call of GetContractState
func (mr *MockIDataQueryMockRecorder) GetContractState(contractid, field interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractState", reflect.TypeOf((*MockIDataQuery)(nil).GetContractState), contractid, field)
}

// GetContractStatesByPrefix mocks base method
func (m *MockIDataQuery) GetContractStatesByPrefix(contractid []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractStatesByPrefix", contractid, prefix)
	ret0, _ := ret[0].(map[string]*modules.ContractStateValue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractStatesByPrefix indicates an expected call of GetContractStatesByPrefix
func (mr *MockIDataQueryMockRecorder) GetContractStatesByPrefix(contractid, prefix interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractStatesByPrefix", reflect.TypeOf((*MockIDataQuery)(nil).GetContractStatesByPrefix), contractid, prefix)
}

// UnstableHeadUnitProperty mocks base method
func (m *MockIDataQuery) UnstableHeadUnitProperty(asset modules.AssetId) (*modules.UnitProperty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnstableHeadUnitProperty", asset)
	ret0, _ := ret[0].(*modules.UnitProperty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnstableHeadUnitProperty indicates an expected call of UnstableHeadUnitProperty
func (mr *MockIDataQueryMockRecorder) UnstableHeadUnitProperty(asset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnstableHeadUnitProperty", reflect.TypeOf((*MockIDataQuery)(nil).UnstableHeadUnitProperty), asset)
}

// GetGlobalProp mocks base method
func (m *MockIDataQuery) GetGlobalProp() *modules.GlobalProperty {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGlobalProp")
	ret0, _ := ret[0].(*modules.GlobalProperty)
	return ret0
}

// GetGlobalProp indicates an expected call of GetGlobalProp
func (mr *MockIDataQueryMockRecorder) GetGlobalProp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGlobalProp", reflect.TypeOf((*MockIDataQuery)(nil).GetGlobalProp))
}

// CurrentHeader mocks base method
func (m *MockIDataQuery) CurrentHeader(token modules.AssetId) *modules.Header {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentHeader", token)
	ret0, _ := ret[0].(*modules.Header)
	return ret0
}

// CurrentHeader indicates an expected call of CurrentHeader
func (mr *MockIDataQueryMockRecorder) CurrentHeader(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentHeader", reflect.TypeOf((*MockIDataQuery)(nil).CurrentHeader), token)
}

// GetHeaderByNumber mocks base method
func (m *MockIDataQuery) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeaderByNumber", number)
	ret0, _ := ret[0].(*modules.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeaderByNumber indicates an expected call of GetHeaderByNumber
func (mr *MockIDataQueryMockRecorder) GetHeaderByNumber(number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeaderByNumber", reflect.TypeOf((*MockIDataQuery)(nil).GetHeaderByNumber), number)
}

// GetAddrUtxos mocks base method
func (m *MockIDataQuery) GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddrUtxos", addr)
	ret0, _ := ret[0].(map[modules.OutPoint]*modules.Utxo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddrUtxos indicates an expected call of GetAddrUtxos
func (mr *MockIDataQueryMockRecorder) GetAddrUtxos(addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddrUtxos", reflect.TypeOf((*MockIDataQuery)(nil).GetAddrUtxos), addr)
}

// GetAddr1TokenUtxos mocks base method
func (m *MockIDataQuery) GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (map[modules.OutPoint]*modules.Utxo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddr1TokenUtxos", addr, asset)
	ret0, _ := ret[0].(map[modules.OutPoint]*modules.Utxo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddr1TokenUtxos indicates an expected call of GetAddr1TokenUtxos
func (mr *MockIDataQueryMockRecorder) GetAddr1TokenUtxos(addr, asset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddr1TokenUtxos", reflect.TypeOf((*MockIDataQuery)(nil).GetAddr1TokenUtxos), addr, asset)
}

// GetStableTransactionOnly mocks base method
func (m *MockIDataQuery) GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStableTransactionOnly", hash)
	ret0, _ := ret[0].(*modules.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStableTransactionOnly indicates an expected call of GetStableTransactionOnly
func (mr *MockIDataQueryMockRecorder) GetStableTransactionOnly(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStableTransactionOnly", reflect.TypeOf((*MockIDataQuery)(nil).GetStableTransactionOnly), hash)
}

// GetStableUnit mocks base method
func (m *MockIDataQuery) GetStableUnit(hash common.Hash) (*modules.Unit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStableUnit", hash)
	ret0, _ := ret[0].(*modules.Unit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStableUnit indicates an expected call of GetStableUnit
func (mr *MockIDataQueryMockRecorder) GetStableUnit(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStableUnit", reflect.TypeOf((*MockIDataQuery)(nil).GetStableUnit), hash)
}

// GetStableUnitByNumber mocks base method
func (m *MockIDataQuery) GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStableUnitByNumber", number)
	ret0, _ := ret[0].(*modules.Unit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStableUnitByNumber indicates an expected call of GetStableUnitByNumber
func (mr *MockIDataQueryMockRecorder) GetStableUnitByNumber(number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStableUnitByNumber", reflect.TypeOf((*MockIDataQuery)(nil).GetStableUnitByNumber), number)
}

// MockIStateQuery is a mock of IStateQuery interface
type MockIStateQuery struct {
	ctrl     *gomock.Controller
	recorder *MockIStateQueryMockRecorder
}

// MockIStateQueryMockRecorder is the mock recorder for MockIStateQuery
type MockIStateQueryMockRecorder struct {
	mock *MockIStateQuery
}

// NewMockIStateQuery creates a new mock instance
func NewMockIStateQuery(ctrl *gomock.Controller) *MockIStateQuery {
	mock := &MockIStateQuery{ctrl: ctrl}
	mock.recorder = &MockIStateQueryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIStateQuery) EXPECT() *MockIStateQueryMockRecorder {
	return m.recorder
}

// GetContractStatesById mocks base method
func (m *MockIStateQuery) GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractStatesById", contractid)
	ret0, _ := ret[0].(map[string]*modules.ContractStateValue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractStatesById indicates an expected call of GetContractStatesById
func (mr *MockIStateQueryMockRecorder) GetContractStatesById(contractid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractStatesById", reflect.TypeOf((*MockIStateQuery)(nil).GetContractStatesById), contractid)
}

// GetContractState mocks base method
func (m *MockIStateQuery) GetContractState(contractid []byte, field string) ([]byte, *modules.StateVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractState", contractid, field)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(*modules.StateVersion)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetContractState indicates an expected call of GetContractState
func (mr *MockIStateQueryMockRecorder) GetContractState(contractid, field interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractState", reflect.TypeOf((*MockIStateQuery)(nil).GetContractState), contractid, field)
}

// GetContractStatesByPrefix mocks base method
func (m *MockIStateQuery) GetContractStatesByPrefix(contractid []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractStatesByPrefix", contractid, prefix)
	ret0, _ := ret[0].(map[string]*modules.ContractStateValue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractStatesByPrefix indicates an expected call of GetContractStatesByPrefix
func (mr *MockIStateQueryMockRecorder) GetContractStatesByPrefix(contractid, prefix interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractStatesByPrefix", reflect.TypeOf((*MockIStateQuery)(nil).GetContractStatesByPrefix), contractid, prefix)
}
