// Code generated by MockGen. DO NOT EDIT.
// Source: ./ICryptoCurrency.go

// Package adaptor is a generated GoMock package.
package adaptor

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockICryptoCurrency is a mock of ICryptoCurrency interface
type MockICryptoCurrency struct {
	ctrl     *gomock.Controller
	recorder *MockICryptoCurrencyMockRecorder
}

// MockICryptoCurrencyMockRecorder is the mock recorder for MockICryptoCurrency
type MockICryptoCurrencyMockRecorder struct {
	mock *MockICryptoCurrency
}

// NewMockICryptoCurrency creates a new mock instance
func NewMockICryptoCurrency(ctrl *gomock.Controller) *MockICryptoCurrency {
	mock := &MockICryptoCurrency{ctrl: ctrl}
	mock.recorder = &MockICryptoCurrencyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockICryptoCurrency) EXPECT() *MockICryptoCurrencyMockRecorder {
	return m.recorder
}

// NewPrivateKey mocks base method
func (m *MockICryptoCurrency) NewPrivateKey(input *NewPrivateKeyInput) (*NewPrivateKeyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewPrivateKey", input)
	ret0, _ := ret[0].(*NewPrivateKeyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewPrivateKey indicates an expected call of NewPrivateKey
func (mr *MockICryptoCurrencyMockRecorder) NewPrivateKey(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewPrivateKey", reflect.TypeOf((*MockICryptoCurrency)(nil).NewPrivateKey), input)
}

// GetPublicKey mocks base method
func (m *MockICryptoCurrency) GetPublicKey(input *GetPublicKeyInput) (*GetPublicKeyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublicKey", input)
	ret0, _ := ret[0].(*GetPublicKeyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublicKey indicates an expected call of GetPublicKey
func (mr *MockICryptoCurrencyMockRecorder) GetPublicKey(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublicKey", reflect.TypeOf((*MockICryptoCurrency)(nil).GetPublicKey), input)
}

// GetAddress mocks base method
func (m *MockICryptoCurrency) GetAddress(key *GetAddressInput) (*GetAddressOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddress", key)
	ret0, _ := ret[0].(*GetAddressOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddress indicates an expected call of GetAddress
func (mr *MockICryptoCurrencyMockRecorder) GetAddress(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddress", reflect.TypeOf((*MockICryptoCurrency)(nil).GetAddress), key)
}

// GetPalletOneMappingAddress mocks base method
func (m *MockICryptoCurrency) GetPalletOneMappingAddress(addr *GetPalletOneMappingAddressInput) (*GetPalletOneMappingAddressOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPalletOneMappingAddress", addr)
	ret0, _ := ret[0].(*GetPalletOneMappingAddressOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPalletOneMappingAddress indicates an expected call of GetPalletOneMappingAddress
func (mr *MockICryptoCurrencyMockRecorder) GetPalletOneMappingAddress(addr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPalletOneMappingAddress", reflect.TypeOf((*MockICryptoCurrency)(nil).GetPalletOneMappingAddress), addr)
}

// SignMessage mocks base method
func (m *MockICryptoCurrency) SignMessage(input *SignMessageInput) (*SignMessageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignMessage", input)
	ret0, _ := ret[0].(*SignMessageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignMessage indicates an expected call of SignMessage
func (mr *MockICryptoCurrencyMockRecorder) SignMessage(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignMessage", reflect.TypeOf((*MockICryptoCurrency)(nil).SignMessage), input)
}

// VerifySignature mocks base method
func (m *MockICryptoCurrency) VerifySignature(input *VerifySignatureInput) (*VerifySignatureOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifySignature", input)
	ret0, _ := ret[0].(*VerifySignatureOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifySignature indicates an expected call of VerifySignature
func (mr *MockICryptoCurrencyMockRecorder) VerifySignature(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifySignature", reflect.TypeOf((*MockICryptoCurrency)(nil).VerifySignature), input)
}

// SignTransaction mocks base method
func (m *MockICryptoCurrency) SignTransaction(input *SignTransactionInput) (*SignTransactionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignTransaction", input)
	ret0, _ := ret[0].(*SignTransactionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignTransaction indicates an expected call of SignTransaction
func (mr *MockICryptoCurrencyMockRecorder) SignTransaction(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignTransaction", reflect.TypeOf((*MockICryptoCurrency)(nil).SignTransaction), input)
}

// BindTxAndSignature mocks base method
func (m *MockICryptoCurrency) BindTxAndSignature(input *BindTxAndSignatureInput) (*BindTxAndSignatureOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BindTxAndSignature", input)
	ret0, _ := ret[0].(*BindTxAndSignatureOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BindTxAndSignature indicates an expected call of BindTxAndSignature
func (mr *MockICryptoCurrencyMockRecorder) BindTxAndSignature(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BindTxAndSignature", reflect.TypeOf((*MockICryptoCurrency)(nil).BindTxAndSignature), input)
}

// CalcTxHash mocks base method
func (m *MockICryptoCurrency) CalcTxHash(input *CalcTxHashInput) (*CalcTxHashOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalcTxHash", input)
	ret0, _ := ret[0].(*CalcTxHashOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CalcTxHash indicates an expected call of CalcTxHash
func (mr *MockICryptoCurrencyMockRecorder) CalcTxHash(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalcTxHash", reflect.TypeOf((*MockICryptoCurrency)(nil).CalcTxHash), input)
}

// SendTransaction mocks base method
func (m *MockICryptoCurrency) SendTransaction(input *SendTransactionInput) (*SendTransactionOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendTransaction", input)
	ret0, _ := ret[0].(*SendTransactionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendTransaction indicates an expected call of SendTransaction
func (mr *MockICryptoCurrencyMockRecorder) SendTransaction(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendTransaction", reflect.TypeOf((*MockICryptoCurrency)(nil).SendTransaction), input)
}

// GetTxBasicInfo mocks base method
func (m *MockICryptoCurrency) GetTxBasicInfo(input *GetTxBasicInfoInput) (*GetTxBasicInfoOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTxBasicInfo", input)
	ret0, _ := ret[0].(*GetTxBasicInfoOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTxBasicInfo indicates an expected call of GetTxBasicInfo
func (mr *MockICryptoCurrencyMockRecorder) GetTxBasicInfo(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTxBasicInfo", reflect.TypeOf((*MockICryptoCurrency)(nil).GetTxBasicInfo), input)
}

// GetBlockInfo mocks base method
func (m *MockICryptoCurrency) GetBlockInfo(input *GetBlockInfoInput) (*GetBlockInfoOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockInfo", input)
	ret0, _ := ret[0].(*GetBlockInfoOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockInfo indicates an expected call of GetBlockInfo
func (mr *MockICryptoCurrencyMockRecorder) GetBlockInfo(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockInfo", reflect.TypeOf((*MockICryptoCurrency)(nil).GetBlockInfo), input)
}

// GetBalance mocks base method
func (m *MockICryptoCurrency) GetBalance(input *GetBalanceInput) (*GetBalanceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", input)
	ret0, _ := ret[0].(*GetBalanceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance
func (mr *MockICryptoCurrencyMockRecorder) GetBalance(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockICryptoCurrency)(nil).GetBalance), input)
}

// GetAssetDecimal mocks base method
func (m *MockICryptoCurrency) GetAssetDecimal(asset *GetAssetDecimalInput) (*GetAssetDecimalOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAssetDecimal", asset)
	ret0, _ := ret[0].(*GetAssetDecimalOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAssetDecimal indicates an expected call of GetAssetDecimal
func (mr *MockICryptoCurrencyMockRecorder) GetAssetDecimal(asset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssetDecimal", reflect.TypeOf((*MockICryptoCurrency)(nil).GetAssetDecimal), asset)
}

// CreateTransferTokenTx mocks base method
func (m *MockICryptoCurrency) CreateTransferTokenTx(input *CreateTransferTokenTxInput) (*CreateTransferTokenTxOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTransferTokenTx", input)
	ret0, _ := ret[0].(*CreateTransferTokenTxOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTransferTokenTx indicates an expected call of CreateTransferTokenTx
func (mr *MockICryptoCurrencyMockRecorder) CreateTransferTokenTx(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTransferTokenTx", reflect.TypeOf((*MockICryptoCurrency)(nil).CreateTransferTokenTx), input)
}

// GetAddrTxHistory mocks base method
func (m *MockICryptoCurrency) GetAddrTxHistory(input *GetAddrTxHistoryInput) (*GetAddrTxHistoryOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddrTxHistory", input)
	ret0, _ := ret[0].(*GetAddrTxHistoryOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddrTxHistory indicates an expected call of GetAddrTxHistory
func (mr *MockICryptoCurrencyMockRecorder) GetAddrTxHistory(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddrTxHistory", reflect.TypeOf((*MockICryptoCurrency)(nil).GetAddrTxHistory), input)
}

// GetTransferTx mocks base method
func (m *MockICryptoCurrency) GetTransferTx(input *GetTransferTxInput) (*GetTransferTxOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTransferTx", input)
	ret0, _ := ret[0].(*GetTransferTxOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTransferTx indicates an expected call of GetTransferTx
func (mr *MockICryptoCurrencyMockRecorder) GetTransferTx(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransferTx", reflect.TypeOf((*MockICryptoCurrency)(nil).GetTransferTx), input)
}

// CreateMultiSigAddress mocks base method
func (m *MockICryptoCurrency) CreateMultiSigAddress(input *CreateMultiSigAddressInput) (*CreateMultiSigAddressOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMultiSigAddress", input)
	ret0, _ := ret[0].(*CreateMultiSigAddressOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMultiSigAddress indicates an expected call of CreateMultiSigAddress
func (mr *MockICryptoCurrencyMockRecorder) CreateMultiSigAddress(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMultiSigAddress", reflect.TypeOf((*MockICryptoCurrency)(nil).CreateMultiSigAddress), input)
}
