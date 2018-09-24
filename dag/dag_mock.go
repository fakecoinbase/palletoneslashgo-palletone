// Code generated by MockGen. DO NOT EDIT.
// Source: ./dag/interface.go

// Package dag is a generated GoMock package.
package dag

import (
	kyber "github.com/dedis/kyber"
	gomock "github.com/golang/mock/gomock"
	common "github.com/palletone/go-palletone/common"
	event "github.com/palletone/go-palletone/common/event"
	discover "github.com/palletone/go-palletone/common/p2p/discover"
	keystore "github.com/palletone/go-palletone/core/accounts/keystore"
	modules "github.com/palletone/go-palletone/dag/modules"
	txspool "github.com/palletone/go-palletone/dag/txspool"
	reflect "reflect"
	time "time"
)

// MockIDag is a mock of IDag interface
type MockIDag struct {
	ctrl     *gomock.Controller
	recorder *MockIDagMockRecorder
}

// MockIDagMockRecorder is the mock recorder for MockIDag
type MockIDagMockRecorder struct {
	mock *MockIDag
}

// NewMockIDag creates a new mock instance
func NewMockIDag(ctrl *gomock.Controller) *MockIDag {
	mock := &MockIDag{ctrl: ctrl}
	mock.recorder = &MockIDagMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIDag) EXPECT() *MockIDagMockRecorder {
	return m.recorder
}

// CurrentUnit mocks base method
func (m *MockIDag) CurrentUnit() *modules.Unit {
	ret := m.ctrl.Call(m, "CurrentUnit")
	ret0, _ := ret[0].(*modules.Unit)
	return ret0
}

// CurrentUnit indicates an expected call of CurrentUnit
func (mr *MockIDagMockRecorder) CurrentUnit() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentUnit", reflect.TypeOf((*MockIDag)(nil).CurrentUnit))
}

// SaveDag mocks base method
func (m *MockIDag) SaveDag(unit modules.Unit, isGenesis bool) (int, error) {
	ret := m.ctrl.Call(m, "SaveDag", unit, isGenesis)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveDag indicates an expected call of SaveDag
func (mr *MockIDagMockRecorder) SaveDag(unit, isGenesis interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveDag", reflect.TypeOf((*MockIDag)(nil).SaveDag), unit, isGenesis)
}

// GetActiveMediatorNodes mocks base method
func (m *MockIDag) GetActiveMediatorNodes() []*discover.Node {
	ret := m.ctrl.Call(m, "GetActiveMediatorNodes")
	ret0, _ := ret[0].([]*discover.Node)
	return ret0
}

// GetActiveMediatorNodes indicates an expected call of GetActiveMediatorNodes
func (mr *MockIDagMockRecorder) GetActiveMediatorNodes() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveMediatorNodes", reflect.TypeOf((*MockIDag)(nil).GetActiveMediatorNodes))
}

// VerifyHeader mocks base method
func (m *MockIDag) VerifyHeader(header *modules.Header, seal bool) error {
	ret := m.ctrl.Call(m, "VerifyHeader", header, seal)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyHeader indicates an expected call of VerifyHeader
func (mr *MockIDagMockRecorder) VerifyHeader(header, seal interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyHeader", reflect.TypeOf((*MockIDag)(nil).VerifyHeader), header, seal)
}

// GetCurrentUnit mocks base method
func (m *MockIDag) GetCurrentUnit(assetId modules.IDType16) *modules.Unit {
	ret := m.ctrl.Call(m, "GetCurrentUnit", assetId)
	ret0, _ := ret[0].(*modules.Unit)
	return ret0
}

// GetCurrentUnit indicates an expected call of GetCurrentUnit
func (mr *MockIDagMockRecorder) GetCurrentUnit(assetId interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentUnit", reflect.TypeOf((*MockIDag)(nil).GetCurrentUnit), assetId)
}

// InsertDag mocks base method
func (m *MockIDag) InsertDag(units modules.Units) (int, error) {
	ret := m.ctrl.Call(m, "InsertDag", units)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertDag indicates an expected call of InsertDag
func (mr *MockIDagMockRecorder) InsertDag(units interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertDag", reflect.TypeOf((*MockIDag)(nil).InsertDag), units)
}

// GetUnitByHash mocks base method
func (m *MockIDag) GetUnitByHash(hash common.Hash) *modules.Unit {
	ret := m.ctrl.Call(m, "GetUnitByHash", hash)
	ret0, _ := ret[0].(*modules.Unit)
	return ret0
}

// GetUnitByHash indicates an expected call of GetUnitByHash
func (mr *MockIDagMockRecorder) GetUnitByHash(hash interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnitByHash", reflect.TypeOf((*MockIDag)(nil).GetUnitByHash), hash)
}

// HasHeader mocks base method
func (m *MockIDag) HasHeader(arg0 common.Hash, arg1 uint64) bool {
	ret := m.ctrl.Call(m, "HasHeader", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasHeader indicates an expected call of HasHeader
func (mr *MockIDagMockRecorder) HasHeader(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasHeader", reflect.TypeOf((*MockIDag)(nil).HasHeader), arg0, arg1)
}

// GetHeaderByNumber mocks base method
func (m *MockIDag) GetHeaderByNumber(number modules.ChainIndex) *modules.Header {
	ret := m.ctrl.Call(m, "GetHeaderByNumber", number)
	ret0, _ := ret[0].(*modules.Header)
	return ret0
}

// GetHeaderByNumber indicates an expected call of GetHeaderByNumber
func (mr *MockIDagMockRecorder) GetHeaderByNumber(number interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeaderByNumber", reflect.TypeOf((*MockIDag)(nil).GetHeaderByNumber), number)
}

// GetHeaderByHash mocks base method
func (m *MockIDag) GetHeaderByHash(arg0 common.Hash) *modules.Header {
	ret := m.ctrl.Call(m, "GetHeaderByHash", arg0)
	ret0, _ := ret[0].(*modules.Header)
	return ret0
}

// GetHeaderByHash indicates an expected call of GetHeaderByHash
func (mr *MockIDagMockRecorder) GetHeaderByHash(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeaderByHash", reflect.TypeOf((*MockIDag)(nil).GetHeaderByHash), arg0)
}

// GetHeader mocks base method
func (m *MockIDag) GetHeader(hash common.Hash, number uint64) (*modules.Header, error) {
	ret := m.ctrl.Call(m, "GetHeader", hash, number)
	ret0, _ := ret[0].(*modules.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeader indicates an expected call of GetHeader
func (mr *MockIDagMockRecorder) GetHeader(hash, number interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeader", reflect.TypeOf((*MockIDag)(nil).GetHeader), hash, number)
}

// CurrentHeader mocks base method
func (m *MockIDag) CurrentHeader() *modules.Header {
	ret := m.ctrl.Call(m, "CurrentHeader")
	ret0, _ := ret[0].(*modules.Header)
	return ret0
}

// CurrentHeader indicates an expected call of CurrentHeader
func (mr *MockIDagMockRecorder) CurrentHeader() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentHeader", reflect.TypeOf((*MockIDag)(nil).CurrentHeader))
}

// GetTransactionByHash mocks base method
func (m *MockIDag) GetTransactionsByHash(hash common.Hash) (modules.Transactions, error) {
	ret := m.ctrl.Call(m, "GetTransactionByHash", hash)
	ret0, _ := ret[0].(modules.Transactions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTransactionByHash indicates an expected call of GetTransactionByHash
func (mr *MockIDagMockRecorder) GetTransactionsByHash(hash interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransactionByHash", reflect.TypeOf((*MockIDag)(nil).GetTransactionsByHash), hash)
}

// InsertHeaderDag mocks base method
func (m *MockIDag) InsertHeaderDag(arg0 []*modules.Header, arg1 int) (int, error) {
	ret := m.ctrl.Call(m, "InsertHeaderDag", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertHeaderDag indicates an expected call of InsertHeaderDag
func (mr *MockIDagMockRecorder) InsertHeaderDag(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertHeaderDag", reflect.TypeOf((*MockIDag)(nil).InsertHeaderDag), arg0, arg1)
}

// HasUnit mocks base method
func (m *MockIDag) HasUnit(hash common.Hash) bool {
	ret := m.ctrl.Call(m, "HasUnit", hash)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasUnit indicates an expected call of HasUnit
func (mr *MockIDagMockRecorder) HasUnit(hash interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasUnit", reflect.TypeOf((*MockIDag)(nil).HasUnit), hash)
}

// SaveUnit mocks base method
func (m *MockIDag) SaveUnit(unit modules.Unit, isGenesis bool) error {
	ret := m.ctrl.Call(m, "SaveUnit", unit, isGenesis)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveUnit indicates an expected call of SaveUnit
func (mr *MockIDagMockRecorder) SaveUnit(unit, isGenesis interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUnit", reflect.TypeOf((*MockIDag)(nil).SaveUnit), unit, isGenesis)
}

// GetAllLeafNodes mocks base method
func (m *MockIDag) GetAllLeafNodes() ([]*modules.Header, error) {
	ret := m.ctrl.Call(m, "GetAllLeafNodes")
	ret0, _ := ret[0].([]*modules.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllLeafNodes indicates an expected call of GetAllLeafNodes
func (mr *MockIDagMockRecorder) GetAllLeafNodes() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllLeafNodes", reflect.TypeOf((*MockIDag)(nil).GetAllLeafNodes))
}

// GetUnit mocks base method
func (m *MockIDag) GetUnit(arg0 common.Hash) *modules.Unit {
	ret := m.ctrl.Call(m, "GetUnit", arg0)
	ret0, _ := ret[0].(*modules.Unit)
	return ret0
}

// GetUnit indicates an expected call of GetUnit
func (mr *MockIDagMockRecorder) GetUnit(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnit", reflect.TypeOf((*MockIDag)(nil).GetUnit), arg0)
}

// CreateUnit mocks base method
func (m *MockIDag) CreateUnit(mAddr *common.Address, txpool *txspool.TxPool, ks *keystore.KeyStore, t time.Time) ([]modules.Unit, error) {
	ret := m.ctrl.Call(m, "CreateUnit", mAddr, txpool, ks, t)
	ret0, _ := ret[0].([]modules.Unit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUnit indicates an expected call of CreateUnit
func (mr *MockIDagMockRecorder) CreateUnit(mAddr, txpool, ks, t interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUnit", reflect.TypeOf((*MockIDag)(nil).CreateUnit), mAddr, txpool, ks, t)
}

// GetActiveMediatorNode mocks base method
func (m *MockIDag) GetActiveMediatorNode(index int) *discover.Node {
	ret := m.ctrl.Call(m, "GetActiveMediatorNode", index)
	ret0, _ := ret[0].(*discover.Node)
	return ret0
}

// GetActiveMediatorNode indicates an expected call of GetActiveMediatorNode
func (mr *MockIDagMockRecorder) GetActiveMediatorNode(index interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveMediatorNode", reflect.TypeOf((*MockIDag)(nil).GetActiveMediatorNode), index)
}

// FastSyncCommitHead mocks base method
func (m *MockIDag) FastSyncCommitHead(arg0 common.Hash) error {
	ret := m.ctrl.Call(m, "FastSyncCommitHead", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// FastSyncCommitHead indicates an expected call of FastSyncCommitHead
func (mr *MockIDagMockRecorder) FastSyncCommitHead(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FastSyncCommitHead", reflect.TypeOf((*MockIDag)(nil).FastSyncCommitHead), arg0)
}

// GetGenesisUnit mocks base method
func (m *MockIDag) GetGenesisUnit(index uint64) (*modules.Unit, error) {
	ret := m.ctrl.Call(m, "GetGenesisUnit", index)
	ret0, _ := ret[0].(*modules.Unit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGenesisUnit indicates an expected call of GetGenesisUnit
func (mr *MockIDagMockRecorder) GetGenesisUnit(index interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGenesisUnit", reflect.TypeOf((*MockIDag)(nil).GetGenesisUnit), index)
}

// GetContractState mocks base method
func (m *MockIDag) GetContractState(id, field string) (*modules.StateVersion, []byte) {
	ret := m.ctrl.Call(m, "GetContractState", id, field)
	ret0, _ := ret[0].(*modules.StateVersion)
	ret1, _ := ret[1].([]byte)
	return ret0, ret1
}

// GetContractState indicates an expected call of GetContractState
func (mr *MockIDagMockRecorder) GetContractState(id, field interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractState", reflect.TypeOf((*MockIDag)(nil).GetContractState), id, field)
}

// GetUnitNumber mocks base method
func (m *MockIDag) GetUnitNumber(hash common.Hash) (modules.ChainIndex, error) {
	ret := m.ctrl.Call(m, "GetUnitNumber", hash)
	ret0, _ := ret[0].(modules.ChainIndex)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUnitNumber indicates an expected call of GetUnitNumber
func (mr *MockIDagMockRecorder) GetUnitNumber(hash interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnitNumber", reflect.TypeOf((*MockIDag)(nil).GetUnitNumber), hash)
}

// GetCanonicalHash mocks base method
func (m *MockIDag) GetCanonicalHash(number uint64) (common.Hash, error) {
	ret := m.ctrl.Call(m, "GetCanonicalHash", number)
	ret0, _ := ret[0].(common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCanonicalHash indicates an expected call of GetCanonicalHash
func (mr *MockIDagMockRecorder) GetCanonicalHash(number interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCanonicalHash", reflect.TypeOf((*MockIDag)(nil).GetCanonicalHash), number)
}

// GetHeadHeaderHash mocks base method
func (m *MockIDag) GetHeadHeaderHash() (common.Hash, error) {
	ret := m.ctrl.Call(m, "GetHeadHeaderHash")
	ret0, _ := ret[0].(common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeadHeaderHash indicates an expected call of GetHeadHeaderHash
func (mr *MockIDagMockRecorder) GetHeadHeaderHash() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeadHeaderHash", reflect.TypeOf((*MockIDag)(nil).GetHeadHeaderHash))
}

// GetHeadUnitHash mocks base method
func (m *MockIDag) GetHeadUnitHash() (common.Hash, error) {
	ret := m.ctrl.Call(m, "GetHeadUnitHash")
	ret0, _ := ret[0].(common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeadUnitHash indicates an expected call of GetHeadUnitHash
func (mr *MockIDagMockRecorder) GetHeadUnitHash() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeadUnitHash", reflect.TypeOf((*MockIDag)(nil).GetHeadUnitHash))
}

// GetHeadFastUnitHash mocks base method
func (m *MockIDag) GetHeadFastUnitHash() (common.Hash, error) {
	ret := m.ctrl.Call(m, "GetHeadFastUnitHash")
	ret0, _ := ret[0].(common.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeadFastUnitHash indicates an expected call of GetHeadFastUnitHash
func (mr *MockIDagMockRecorder) GetHeadFastUnitHash() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeadFastUnitHash", reflect.TypeOf((*MockIDag)(nil).GetHeadFastUnitHash))
}

// GetUtxoView mocks base method
func (m *MockIDag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	ret := m.ctrl.Call(m, "GetUtxoView", tx)
	ret0, _ := ret[0].(*txspool.UtxoViewpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUtxoView indicates an expected call of GetUtxoView
func (mr *MockIDagMockRecorder) GetUtxoView(tx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUtxoView", reflect.TypeOf((*MockIDag)(nil).GetUtxoView), tx)
}

// SubscribeChainHeadEvent mocks base method
func (m *MockIDag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	ret := m.ctrl.Call(m, "SubscribeChainHeadEvent", ch)
	ret0, _ := ret[0].(event.Subscription)
	return ret0
}

// SubscribeChainHeadEvent indicates an expected call of SubscribeChainHeadEvent
func (mr *MockIDagMockRecorder) SubscribeChainHeadEvent(ch interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeChainHeadEvent", reflect.TypeOf((*MockIDag)(nil).SubscribeChainHeadEvent), ch)
}

// GetTrieSyncProgress mocks base method
func (m *MockIDag) GetTrieSyncProgress() (uint64, error) {
	ret := m.ctrl.Call(m, "GetTrieSyncProgress")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTrieSyncProgress indicates an expected call of GetTrieSyncProgress
func (mr *MockIDagMockRecorder) GetTrieSyncProgress() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrieSyncProgress", reflect.TypeOf((*MockIDag)(nil).GetTrieSyncProgress))
}

// GetUtxoEntry mocks base method
func (m *MockIDag) GetUtxoEntry(key []byte) (*modules.Utxo, error) {
	ret := m.ctrl.Call(m, "GetUtxoEntry", key)
	ret0, _ := ret[0].(*modules.Utxo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUtxoEntry indicates an expected call of GetUtxoEntry
func (mr *MockIDagMockRecorder) GetUtxoEntry(key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUtxoEntry", reflect.TypeOf((*MockIDag)(nil).GetUtxoEntry), key)
}

// GetAddrOutput mocks base method
func (m *MockIDag) GetAddrOutput(addr string) ([]modules.Output, error) {
	ret := m.ctrl.Call(m, "GetAddrOutput", addr)
	ret0, _ := ret[0].([]modules.Output)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddrOutput indicates an expected call of GetAddrOutput
func (mr *MockIDagMockRecorder) GetAddrOutput(addr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddrOutput", reflect.TypeOf((*MockIDag)(nil).GetAddrOutput), addr)
}

// GetAddrTransactions mocks base method
func (m *MockIDag) GetAddrTransactions(addr string) (modules.Transactions, error) {
	ret := m.ctrl.Call(m, "GetAddrTransactions", addr)
	ret0, _ := ret[0].(modules.Transactions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddrTransactions indicates an expected call of GetAddrTransactions
func (mr *MockIDagMockRecorder) GetAddrTransactions(addr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddrTransactions", reflect.TypeOf((*MockIDag)(nil).GetAddrTransactions), addr)
}

// GetContractTpl mocks base method
func (m *MockIDag) GetContractTpl(templateID []byte) (*modules.StateVersion, []byte, string, string) {
	ret := m.ctrl.Call(m, "GetContractTpl", templateID)
	ret0, _ := ret[0].(*modules.StateVersion)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(string)
	return ret0, ret1, ret2, ret3
}

// GetContractTpl indicates an expected call of GetContractTpl
func (mr *MockIDagMockRecorder) GetContractTpl(templateID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractTpl", reflect.TypeOf((*MockIDag)(nil).GetContractTpl), templateID)
}

// WalletTokens mocks base method
func (m *MockIDag) WalletTokens(addr common.Address) (map[string]*modules.AccountToken, error) {
	ret := m.ctrl.Call(m, "WalletTokens", addr)
	ret0, _ := ret[0].(map[string]*modules.AccountToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WalletTokens indicates an expected call of WalletTokens
func (mr *MockIDagMockRecorder) WalletTokens(addr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WalletTokens", reflect.TypeOf((*MockIDag)(nil).WalletTokens), addr)
}

// WalletBalance mocks base method
func (m *MockIDag) WalletBalance(address common.Address, assetid, uniqueid []byte, chainid uint64) (uint64, error) {
	ret := m.ctrl.Call(m, "WalletBalance", address, assetid, uniqueid, chainid)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WalletBalance indicates an expected call of WalletBalance
func (mr *MockIDagMockRecorder) WalletBalance(address, assetid, uniqueid, chainid interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WalletBalance", reflect.TypeOf((*MockIDag)(nil).WalletBalance), address, assetid, uniqueid, chainid)
}

// GetContract mocks base method
func (m *MockIDag) GetContract(id common.Hash) (*modules.Contract, error) {
	ret := m.ctrl.Call(m, "GetContract", id)
	ret0, _ := ret[0].(*modules.Contract)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContract indicates an expected call of GetContract
func (mr *MockIDagMockRecorder) GetContract(id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContract", reflect.TypeOf((*MockIDag)(nil).GetContract), id)
}

// GetActiveMediatorAddr mocks base method
func (m *MockIDag) GetActiveMediatorAddr(index int) common.Address {
	ret := m.ctrl.Call(m, "GetActiveMediatorAddr", index)
	ret0, _ := ret[0].(common.Address)
	return ret0
}

// GetActiveMediatorAddr indicates an expected call of GetActiveMediatorAddr
func (mr *MockIDagMockRecorder) GetActiveMediatorAddr(index interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveMediatorAddr", reflect.TypeOf((*MockIDag)(nil).GetActiveMediatorAddr), index)
}

// GetActiveMediatorInitPubs mocks base method
func (m *MockIDag) GetActiveMediatorInitPubs() []kyber.Point {
	ret := m.ctrl.Call(m, "GetActiveMediatorInitPubs")
	ret0, _ := ret[0].([]kyber.Point)
	return ret0
}

// GetActiveMediatorInitPubs indicates an expected call of GetActiveMediatorInitPubs
func (mr *MockIDagMockRecorder) GetActiveMediatorInitPubs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveMediatorInitPubs", reflect.TypeOf((*MockIDag)(nil).GetActiveMediatorInitPubs))
}

// GetCurThreshold mocks base method
func (m *MockIDag) GetCurThreshold() int {
	ret := m.ctrl.Call(m, "GetCurThreshold")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetCurThreshold indicates an expected call of GetCurThreshold
func (mr *MockIDagMockRecorder) GetCurThreshold() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurThreshold", reflect.TypeOf((*MockIDag)(nil).GetCurThreshold))
}

// IsActiveMediator mocks base method
func (m *MockIDag) IsActiveMediator(add common.Address) bool {
	ret := m.ctrl.Call(m, "IsActiveMediator", add)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsActiveMediator indicates an expected call of IsActiveMediator
func (mr *MockIDagMockRecorder) IsActiveMediator(add interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsActiveMediator", reflect.TypeOf((*MockIDag)(nil).IsActiveMediator), add)
}

// GetGlobalProp mocks base method
func (m *MockIDag) GetGlobalProp() *modules.GlobalProperty {
	ret := m.ctrl.Call(m, "GetGlobalProp")
	ret0, _ := ret[0].(*modules.GlobalProperty)
	return ret0
}

// GetGlobalProp indicates an expected call of GetGlobalProp
func (mr *MockIDagMockRecorder) GetGlobalProp() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGlobalProp", reflect.TypeOf((*MockIDag)(nil).GetGlobalProp))
}

// GetDynGlobalProp mocks base method
func (m *MockIDag) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	ret := m.ctrl.Call(m, "GetDynGlobalProp")
	ret0, _ := ret[0].(*modules.DynamicGlobalProperty)
	return ret0
}

// GetDynGlobalProp indicates an expected call of GetDynGlobalProp
func (mr *MockIDagMockRecorder) GetDynGlobalProp() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDynGlobalProp", reflect.TypeOf((*MockIDag)(nil).GetDynGlobalProp))
}

// GetMediatorSchl mocks base method
func (m *MockIDag) GetMediatorSchl() *modules.MediatorSchedule {
	ret := m.ctrl.Call(m, "GetMediatorSchl")
	ret0, _ := ret[0].(*modules.MediatorSchedule)
	return ret0
}

// GetMediatorSchl indicates an expected call of GetMediatorSchl
func (mr *MockIDagMockRecorder) GetMediatorSchl() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMediatorSchl", reflect.TypeOf((*MockIDag)(nil).GetMediatorSchl))
}