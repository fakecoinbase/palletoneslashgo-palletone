/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package modules

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/obj"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/parameter"
)

var (
	//TXFEE       = big.NewInt(100000000) // transaction fee =1ptn
	TX_MAXSIZE  = 256 * 1024 //256kb
	TX_BASESIZE = 100 * 1024 //100kb
)

//一个交易的状态
type TxStatus byte

const (
	TxStatus_NotFound TxStatus = iota //找不到该交易
	TxStatus_InPool                   //未打包
	TxStatus_Unstable                 //已打包未稳定
	TxStatus_Stable                   //已打包，已稳定
)

func (s TxStatus) String() string {
	switch s {
	case TxStatus_NotFound:
		return "NotFound"
	case TxStatus_InPool:
		return "InPool"
	case TxStatus_Unstable:
		return "Unstable"
	case TxStatus_Stable:
		return "Stable"
	}
	return "Unknown"
}

//var DepositContractLockScript = common.Hex2Bytes("140000000000000000000000000000000000000001c8")

// TxOut defines a bitcoin transaction output.
type TxOut struct {
	Value    int64  `json:"value"`
	PkScript []byte `json:"pk_script"`
	Asset    *Asset `json:"asset_info"`
}

// TxIn defines a bitcoin transaction input.
type TxIn struct {
	PreviousOutPoint *OutPoint `json:"pre_outpoint"`
	SignatureScript  []byte    `json:"signature_script"`
	Sequence         uint32    `json:"sequence"`
}

func NewTransaction(msg []*Message) *Transaction {
	return newTransaction(msg)
}

//func NewContractCreation(msg []*Message) *Transaction {
//	return newTransaction(msg)
//}

func newTransaction(msg []*Message) *Transaction {
	tx := transaction_sdw{}
	if len(msg) > 0 {
		tx.TxMessages = make([]*Message, len(msg))
		copy(tx.TxMessages, msg)
	}
	return &Transaction{txdata: tx}
}

// AddTxIn adds a transaction input to the message.
func (tx *Transaction) AddMessage(msg *Message) {
	msgs := tx.Messages()
	if msg != nil {
		msgs = append(msgs, CopyMessage(msg))
	}
	tx.SetMessages(msgs)
}

type TransactionWithUnitInfo struct {
	*Transaction
	UnitHash  common.Hash
	UnitIndex uint64
	Timestamp uint64
	TxIndex   uint64
}

type TxPackInfo struct {
	TxHash      common.Hash
	RequestHash common.Hash
	UnitHash    common.Hash
	UnitIndex   uint64
	Timestamp   uint64
	TxIndex     uint64
	Error       string
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		cHash := hash.(common.Hash)
		if !common.EmptyHash(cHash) {
			return cHash
		}
	}
	oldFlag := tx.Illegal()
	if oldFlag {
		tx.txdata.Illegal = false
		v := util.RlpHash(tx)
		tx.hash.Store(v)
		tx.txdata.Illegal = true
		return v
	}
	v := util.RlpHash(tx)
	tx.hash.Store(v)
	return v
}

func (tx *Transaction) RequestHash() common.Hash {

	if hash := tx.reqHash.Load(); hash != nil {
		cHash := hash.(common.Hash)
		if !common.EmptyHash(cHash) {
			return cHash
		}
	}

	v := tx.GetRequestTx().Hash()
	tx.reqHash.Store(v)
	return v

}
func (tx *Transaction) ErrorResult() string {
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE {
			invoke := msg.Payload.(*ContractInvokePayload)
			return invoke.ErrMsg.Message
		}
		if msg.App == APP_CONTRACT_DEPLOY {
			dep := msg.Payload.(*ContractDeployPayload)
			return dep.ErrMsg.Message
		}
		if msg.App == APP_CONTRACT_STOP {
			stop := msg.Payload.(*ContractStopPayload)
			return stop.ErrMsg.Message
		}
		if msg.App == APP_CONTRACT_TPL {
			tpl := msg.Payload.(*ContractTplPayload)
			return tpl.ErrMsg.Message
		}
	}
	return ""
}

func (tx *Transaction) GetContractId() []byte {
	for _, msg := range tx.txdata.TxMessages {
		switch msg.App {
		case APP_CONTRACT_DEPLOY_REQUEST:
			addr := crypto.RequestIdToContractAddress(tx.RequestHash())
			return addr.Bytes()
		case APP_CONTRACT_INVOKE_REQUEST:
			payload := msg.Payload.(*ContractInvokeRequestPayload)
			return common.CopyBytes(payload.ContractId)
		case APP_CONTRACT_STOP_REQUEST:
			payload := msg.Payload.(*ContractStopRequestPayload)
			return common.CopyBytes(payload.ContractId)
		}
	}
	return nil
}

//浅拷贝
func (tx *Transaction) Messages() []*Message {
	msgs := make([]*Message, len(tx.txdata.TxMessages))
	copy(msgs, tx.txdata.TxMessages)
	return msgs
}

// 深拷贝
func (tx *Transaction) TxMessages() []*Message {
	temp_msgs := make([]*Message, 0)
	for _, msg := range tx.txdata.TxMessages {
		temp_msgs = append(temp_msgs, CopyMessage(msg))
	}

	return temp_msgs
}

// Size returns the true RLP encoded storage UnitSize of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	size := CalcDateSize(tx)
	tx.size.Store(size)
	return size
}

func (tx *Transaction) Asset() *Asset {
	if tx == nil {
		return nil
	}
	asset := new(Asset)
	msg := tx.txdata.TxMessages[0]
	if msg.App == APP_PAYMENT {
		pay := msg.Payload.(*PaymentPayload)
		for _, out := range pay.Outputs {
			if out.Asset != nil {
				asset.AssetId = out.Asset.AssetId
				asset.UniqueId = out.Asset.UniqueId
				break
			}
		}
	}
	return asset
}
func (tx *Transaction) CopyFrTransaction(cpy *Transaction) {
	obj.DeepCopy(&tx, cpy)
}

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

type WriteCounter common.StorageSize

func (c *WriteCounter) Write(b []byte) (int, error) {
	*c += WriteCounter(len(b))
	return len(b), nil
}

func CalcDateSize(data interface{}) common.StorageSize {
	c := WriteCounter(0)
	rlp.Encode(&c, data)
	return common.StorageSize(c)
}

var (
	EmptyRootHash = core.DeriveSha(Transactions{})
)

type TxLookupEntry struct {
	UnitHash  common.Hash `json:"unit_hash"`
	UnitIndex uint64      `json:"unit_index"`
	Index     uint64      `json:"index"`
	Timestamp uint64      `json:"timestamp"`
}
type Transactions []*Transaction

func (txs Transactions) GetTxIds() []common.Hash {
	ids := make([]common.Hash, len(txs))
	for i, tx := range txs {
		ids[i] = tx.Hash()
	}
	return ids
}

type Transaction struct {
	txdata  transaction_sdw
	hash    atomic.Value
	reqHash atomic.Value
	size    atomic.Value
}
type transaction_sdw struct {
	Version      uint32     `json:"version"`
	AccountNonce uint64     `json:"nonce"`
	TxMessages   []*Message `json:"messages"`
	CertId       []byte     `json:"cert_id"` // should be big.Int byte
	Illegal      bool       `json:"Illegal"` // not hash, 1:no valid, 0:ok
}
type QueryUtxoFunc func(outpoint *OutPoint) (*Utxo, error)
type GetAddressFromScriptFunc func(lockScript []byte) (common.Address, error)
type GetScriptSignersFunc func(tx *Transaction, msgIdx, inputIndex int) ([]common.Address, error)
type QueryStateByVersionFunc func(id []byte, field string, version *StateVersion) ([]byte, error)
type GetJurorRewardAddFunc func(jurorAdd common.Address) common.Address

//计算该交易的手续费，基于UTXO，所以传入查询UTXO的函数指针
func (tx *Transaction) GetTxFee(queryUtxoFunc QueryUtxoFunc) (*AmountAsset, error) {
	msg0 := tx.txdata.TxMessages[0]
	if msg0.App != APP_PAYMENT { //no gas fee
		//return nil, errors.New("Tx message 0 must a payment payload")
		return NewAmountAsset(0, nil), nil
	}
	payload := msg0.Payload.(*PaymentPayload)

	if payload.IsCoinbase() {
		return NewAmountAsset(0, nil), nil
	}
	inAmount := uint64(0)
	outAmount := uint64(0)
	var feeAsset *Asset
	for _, txin := range payload.Inputs {
		utxo, err := queryUtxoFunc(txin.PreviousOutPoint)
		if err != nil {
			return nil, fmt.Errorf("Txin(txhash=%s, msgindex=%v, outindex=%v)'s utxo is empty:%s",
				txin.PreviousOutPoint.TxHash.String(),
				txin.PreviousOutPoint.MessageIndex,
				txin.PreviousOutPoint.OutIndex,
				err.Error())
		}
		feeAsset = utxo.Asset
		// check overflow
		if inAmount+utxo.Amount > (1<<64 - 1) {
			return nil, fmt.Errorf("Compute fees: txin total overflow")
		}
		inAmount += utxo.Amount

		//if unitTime > 0 {
		//	//计算币龄利息
		//	rate := parameter.CurrentSysParameters.TxCoinDayInterest
		//	if bytes.Equal(utxo.PkScript, DepositContractLockScript) {
		//		rate = parameter.CurrentSysParameters.DepositContractInterest
		//	}
		//
		//	interest := award.GetCoinDayInterest(utxo.GetTimestamp(), unitTime, utxo.Amount, rate)
		//	if interest > 0 {
		//		//	log.Infof("Calculate tx fee,Add interest value:%d to tx[%s] fee", interest, tx.Hash().String())
		//		inAmount += interest
		//	}
		//}

	}

	for _, txout := range payload.Outputs {
		// check overflow
		if outAmount+txout.Value > (1<<64 - 1) {
			return nil, fmt.Errorf("Compute fees: txout total overflow")
		}
		outAmount += txout.Value
	}
	if inAmount < outAmount {
		return nil, fmt.Errorf("Compute fees: tx %s txin amount less than txout amount. amount:%d ,outAmount:%d ",
			tx.Hash().String(), inAmount, outAmount)
	}
	log.Debugf("Compute fees: tx %s txin amount more than txout amount. amount:%d ,outAmount:%d ",
		tx.Hash().String(), inAmount, outAmount)
	fees := inAmount - outAmount

	return &AmountAsset{Amount: fees, Asset: feeAsset}, nil
}

func (tx *Transaction) CertId() []byte { return common.CopyBytes(tx.txdata.CertId) }
func (tx *Transaction) Illegal() bool  { return tx.txdata.Illegal }

func (tx *Transaction) SetMessages(msgs []*Message) {
	if len(msgs) > 0 {
		txMessages := make([]*Message, len(msgs))
		copy(txMessages, msgs)
		tx.txdata.TxMessages = txMessages
	}
	tx.resetCache()
}

func (tx *Transaction) resetCache() {
	hash := common.Hash{}
	tx.hash.Store(hash)
	tx.reqHash.Store(hash)
	size0 := common.StorageSize(0)
	tx.size.Store(size0)
}

func (tx *Transaction) SetCertId(certid []byte) {
	tx.txdata.CertId = certid
	tx.resetCache()
}
func (tx *Transaction) SetIllegal(illegal bool) {
	tx.txdata.Illegal = illegal
	// tx.resetCache()  //SetIllegal 不会导致TxHash改变
}

func (tx *Transaction) ModifiedMsg(index int, msg *Message) {
	if len(tx.Messages()) < index {
		return
	}
	tx.txdata.TxMessages[index] = msg
	tx.resetCache()
}

func (tx *Transaction) GetCoinbaseReward(versionFunc QueryStateByVersionFunc,
	scriptFunc GetAddressFromScriptFunc) (*AmountAsset, error) {
	writeMap := make(map[string][]AmountAsset)
	readMap := make(map[string][]AmountAsset)
	msgs := tx.TxMessages()
	if len(msgs) == 2 && msgs[0].App == APP_PAYMENT &&
		msgs[1].App == APP_CONTRACT_INVOKE { //进行了汇总付款
		invoke := msgs[1].Payload.(*ContractInvokePayload)
		for _, read := range invoke.ReadSet {
			readResult, err := versionFunc(read.ContractId, read.Key, read.Version)
			if err != nil {
				return nil, err
			}
			var aa []AmountAsset
			err = rlp.DecodeBytes(readResult, &aa)
			if err != nil {
				return nil, err
			}
			addr := read.Key[len(constants.RewardAddressPrefix):]
			readMap[addr] = aa
		}
		payment := msgs[0].Payload.(*PaymentPayload)
		for _, out := range payment.Outputs {
			aa := AmountAsset{
				Amount: out.Value,
				Asset:  out.Asset,
			}
			addr, _ := scriptFunc(out.PkScript)
			writeMap[addr.String()] = []AmountAsset{aa}
		}
	} else if msgs[0].App == APP_CONTRACT_INVOKE { //进行了记账
		invoke := msgs[0].Payload.(*ContractInvokePayload)
		for _, write := range invoke.WriteSet {
			var aa []AmountAsset
			err := rlp.DecodeBytes(write.Value, &aa)
			if err != nil {
				return nil, err
			}
			addr := write.Key[len(constants.RewardAddressPrefix):]
			writeMap[addr] = aa
		}

		for _, read := range invoke.ReadSet {
			readResult, err := versionFunc(read.ContractId, read.Key, read.Version)
			if err != nil {
				return nil, err
			}
			var aa []AmountAsset
			err = rlp.DecodeBytes(readResult, &aa)
			if err != nil {
				return nil, err
			}
			addr := read.Key[len(constants.RewardAddressPrefix):]
			readMap[addr] = aa
		}
	} else {
		return &AmountAsset{Amount: 0}, nil
	}

	//计算Write Map和Read Map的差，获得Reward值
	reward := &AmountAsset{}
	for writeAddr, writeAA := range writeMap {
		reward.Asset = writeAA[0].Asset
		if readAA, ok := readMap[writeAddr]; ok {
			readAmt := uint64(0)
			if len(readAA) != 0 { //上一次没有清空
				readAmt = readAA[0].Amount
			}
			reward.Amount += writeAA[0].Amount - readAmt
		} else {
			reward.Amount += writeAA[0].Amount
		}
	}
	return reward, nil
}

//该Tx如果保存后，会产生的新的Utxo
func (tx *Transaction) GetNewUtxos() map[OutPoint]*Utxo {
	result := map[OutPoint]*Utxo{}
	txHash := tx.Hash()
	for msgIndex, msg := range tx.txdata.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		txouts := pay.Outputs
		for outIndex, txout := range txouts {
			utxo := &Utxo{
				Amount:   txout.Value,
				Asset:    txout.Asset,
				PkScript: txout.PkScript,
				LockTime: pay.LockTime,
			}

			// write to database
			outpoint := OutPoint{
				TxHash:       txHash,
				MessageIndex: uint32(msgIndex),
				OutIndex:     uint32(outIndex),
			}
			result[outpoint] = utxo
		}
	}
	return result
}

//该Tx如果保存后，会产生的新的Utxo包括ReqUtxo
func (tx *Transaction) GetNewTxUtxoAndReqUtxos() map[OutPoint]*Utxo {
	result := map[OutPoint]*Utxo{}
	txHash := tx.Hash()
	reqHash := tx.RequestHash()
	for msgIndex, msg := range tx.txdata.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		txouts := pay.Outputs
		for outIndex, txout := range txouts {
			utxo := &Utxo{
				Amount:   txout.Value,
				Asset:    txout.Asset,
				PkScript: txout.PkScript,
				LockTime: pay.LockTime,
			}

			// write to database
			outpoint := OutPoint{
				TxHash:       txHash,
				MessageIndex: uint32(msgIndex),
				OutIndex:     uint32(outIndex),
			}
			result[outpoint] = utxo
			if reqHash != txHash {
				outpoint2 := outpoint
				outpoint2.TxHash = reqHash
				result[outpoint2] = utxo
			}
		}
	}
	return result
}

//获取一个交易中花费了哪些OutPoint
func (tx *Transaction) GetSpendOutpoints() []*OutPoint {
	result := []*OutPoint{}
	chongfu := false
	for _, msg := range tx.txdata.TxMessages {
		if msg.App != APP_PAYMENT {
			continue
		}
		pay := msg.Payload.(*PaymentPayload)
		inputs := pay.Inputs
		for _, input := range inputs {
			if input.PreviousOutPoint != nil {
				if input.PreviousOutPoint.TxHash.IsSelfHash() { //合约Payback的情形
					op := NewOutPoint(tx.Hash(), input.PreviousOutPoint.MessageIndex, input.PreviousOutPoint.OutIndex)
					result = append(result, op)
				} else {
					for _, v := range result {
						if v.TxHash.String() == input.PreviousOutPoint.TxHash.String() {
							if v.MessageIndex == input.PreviousOutPoint.MessageIndex && v.OutIndex == input.PreviousOutPoint.OutIndex {
								chongfu = true
							}
						}
					}
					if chongfu == false {
						result = append(result, input.PreviousOutPoint)
					}
					chongfu = false

				}
			}
		}
	}
	return result
}

//获得合约交易的签名对应的陪审员地址
//func (tx *Transaction) GetContractTxSignatureAddress() []common.Address {
//	if !tx.IsContractTx() {
//		return nil
//	}
//	for _, msg := range tx.txdata.TxMessages {
//		switch msg.App {
//		case APP_SIGNATURE:
//			payload := msg.Payload.(*SignaturePayload)
//			return payload.SignAddress()
//		}
//	}
//	return []common.Address{}
//}

//获得合约结果部分的Signature，如果不是合约Tx，则返回第一个Signature
//func (tx *Transaction) GetResultSignature() (int, *SignaturePayload, error) {
//	reqSign := tx.NeedRequestSignature()
//	for msgIdx, msg := range tx.txdata.TxMessages {
//		if msg.App == APP_SIGNATURE {
//			if reqSign {
//				reqSign = false
//			} else {
//				return msgIdx, msg.Payload.(*SignaturePayload), nil
//			}
//		}
//	}
//	return 0, nil, errors.ErrMessageNotFound
//}

//如果是合约调用交易，Copy其中的Msg0到ContractRequest的部分，如果不是请求，那么返回完整Tx
func (tx *Transaction) GetRequestTx() *Transaction {
	msgs := tx.TxMessages()
	request := transaction_sdw{}
	if msgs[0].App != APP_PAYMENT { //no gas fee
		for _, msg := range msgs {
			request.TxMessages = append(request.TxMessages, msg)
			if msg.App == APP_SIGNATURE {
				break
			}
		}
	} else { //有GasFee
		for _, msg := range msgs {
			request.TxMessages = append(request.TxMessages, msg)
			if msg.App.IsRequest() {
				break
			}
		}
	}
	request.CertId = tx.CertId()
	request.Version = tx.Version()
	request.AccountNonce = tx.Nonce()
	return &Transaction{txdata: request}

}

//判断一个交易的请求部分有多少个Message，如果不是一个请求，则返回全部Message数量
func (tx *Transaction) GetRequestMsgCount() int {
	if tx.txdata.TxMessages[0].App != APP_PAYMENT { //no gas fee
		for i, msg := range tx.txdata.TxMessages {
			if msg.App == APP_SIGNATURE {
				return i + 1
			}
		}
	}
	//有GasFee
	for i, msg := range tx.txdata.TxMessages {
		if msg.App.IsRequest() {
			return i + 1
		}
	}
	return len(tx.txdata.TxMessages)
}

//获得一个交易中执行结果的签名Payload和所在的MessageIndex,如果没有执行结果，则返回nil
func (tx *Transaction) GetResultSignaturePayload() (*SignaturePayload, int) {
	signCount := 1
	if tx.txdata.TxMessages[0].App != APP_PAYMENT { //no gas fee
		signCount = 2
	}
	for i, msg := range tx.txdata.TxMessages {
		if msg.App == APP_SIGNATURE {
			signCount--
			if signCount == 0 {
				return msg.Payload.(*SignaturePayload), i
			}
		}
	}
	return nil, -1
}

//获取一个被Jury执行完成后，但是还没有进行陪审员签名的交易
func (tx *Transaction) GetResultRawTx() *Transaction {
	sdw := transaction_sdw{}
	reqMsgCount := tx.GetRequestMsgCount()
	for i, msg := range tx.TxMessages() {
		if i < reqMsgCount {
			sdw.TxMessages = append(sdw.TxMessages, msg)
			continue
		}
		if msg.App == APP_SIGNATURE {
			continue //移除SignaturePayload
		}
		if msg.App == APP_PAYMENT { //移除ContractPayout中的解锁脚本
			pay := msg.Payload.(*PaymentPayload)
			for _, in := range pay.Inputs {
				in.SignatureScript = nil
			}
		}
		sdw.TxMessages = append(sdw.TxMessages, msg)
	}
	sdw.CertId = tx.CertId()
	sdw.Version = tx.Version()
	sdw.AccountNonce = tx.Nonce()
	sdw.Illegal = tx.Illegal()
	return &Transaction{txdata: sdw}
}

//复制一个只包含前面几条Message的Tx，包括msgIdx本身这条
func (tx *Transaction) CopyPartTx(msgIdx int) *Transaction {
	sdw := transaction_sdw{}
	for i, msg := range tx.TxMessages() {
		if i > msgIdx {
			break
		}
		sdw.TxMessages = append(sdw.TxMessages, msg)
	}
	sdw.CertId = tx.CertId()
	sdw.Version = tx.Version()
	sdw.AccountNonce = tx.Nonce()
	sdw.Illegal = tx.Illegal()
	return &Transaction{txdata: sdw}
}

//
//func (tx *Transaction) GetResultTx() *Transaction {
//	sdw := transaction_sdw{}
//	for _, msg := range tx.TxMessages() {
//		if msg.App == APP_SIGNATURE {
//			continue //移除SignaturePayload
//		}
//		sdw.TxMessages = append(sdw.TxMessages, msg)
//	}
//	sdw.CertId = tx.CertId()
//	return &Transaction{txdata: sdw}
//}

//Request 这条Message的Index是多少
func (tx *Transaction) GetRequestMsgIndex() int {
	for idx, msg := range tx.txdata.TxMessages {
		if msg.App.IsRequest() {
			return idx
		}
	}
	return -1
}

//这个交易是否包含了从合约付款出去的结果,有则返回该Payment
func (tx *Transaction) HasContractPayoutMsg() (bool, int, *Message) {
	reqMsgCount := tx.GetRequestMsgCount()
	for i, msg := range tx.txdata.TxMessages {
		if i < reqMsgCount {
			continue
		}
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			if !pay.IsCoinbase() {
				return true, i, msg
			}
		}
	}
	return false, 0, nil
}

//获取该交易的所有From地址
func (tx *Transaction) GetFromAddrs(queryUtxoFunc QueryUtxoFunc, getAddrFunc GetAddressFromScriptFunc) ([]common.Address, error) {
	if tx.txdata.TxMessages[0].App != APP_PAYMENT { //没有GasFee
		for _, msg := range tx.txdata.TxMessages {
			if msg.App == APP_SIGNATURE {
				sign := msg.Payload.(*SignaturePayload)
				return sign.SignAddress(), nil
			}
		}
	}
	//有GasFee
	resultMap := map[common.Address]int{}
	msgs := tx.TxMessages()
	for _, msg := range msgs {
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			for _, input := range pay.Inputs {
				if input.PreviousOutPoint != nil {
					var lockScript []byte
					txo, err := queryUtxoFunc(input.PreviousOutPoint)
					if err != nil {
						if input.PreviousOutPoint.TxHash.IsSelfHash() {
							out := msgs[input.PreviousOutPoint.MessageIndex].Payload.(*PaymentPayload).Outputs[input.PreviousOutPoint.OutIndex]
							lockScript = out.PkScript
						} else {
							log.Errorf("[%s]Cannot find txo by:%s", tx.RequestHash().ShortStr(), input.PreviousOutPoint.String())
							return nil, err
						}
					} else {
						lockScript = txo.PkScript
					}

					addr, _ := getAddrFunc(lockScript)
					if _, ok := resultMap[addr]; !ok {
						resultMap[addr] = 1
					}
				}

			}
		}
	}
	keys := make([]common.Address, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys, nil
}
func (tx *Transaction) GetToAddrs(getAddrFunc GetAddressFromScriptFunc) ([]common.Address, error) {
	resultMap := map[common.Address]int{}
	msgs := tx.TxMessages()
	for _, msg := range msgs {
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			for _, output := range pay.Outputs {
				lockScript := output.PkScript
				addr, _ := getAddrFunc(lockScript)
				if _, ok := resultMap[addr]; !ok {
					resultMap[addr] = 1
				}
			}
		}
	}

	keys := make([]common.Address, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}
	return keys, nil
}

//判断一个Tx是否有请求的SignaturePayload
func (tx *Transaction) NeedRequestSignature() bool {
	return tx.txdata.TxMessages[0].App != APP_PAYMENT
}

//根据Tx中的第一个SignaturePayload，计算签名者地址
func (tx *Transaction) GetSignatureAddr() (common.Address, error) {
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_SIGNATURE {
			sign := msg.Payload.(*SignaturePayload)
			if len(sign.Signatures) == 0 {
				return common.Address{}, errors.New("Invalid SignaturePayload")
			}
			return crypto.PubkeyBytesToAddress(sign.Signatures[0].PubKey), nil
		}
	}
	return common.Address{}, errors.New("SignaturePayload not found")
}

//获取该交易的发起人地址
func (tx *Transaction) GetRequesterAddr(queryUtxoFunc QueryUtxoFunc, getAddrFunc GetAddressFromScriptFunc) (
	common.Address, error) {
	msg0 := tx.txdata.TxMessages[0]
	if msg0.App != APP_PAYMENT {
		return common.Address{}, fmt.Errorf("[%s]Coinbase or Invalid Tx, first message must be a payment",
			tx.RequestHash().ShortStr())
	}
	pay := msg0.Payload.(*PaymentPayload)

	utxo, err := queryUtxoFunc(pay.Inputs[0].PreviousOutPoint)
	if err != nil {
		return common.Address{}, err
	}
	return getAddrFunc(utxo.PkScript)

}

func (tx *Transaction) GetContractTxType() MessageType {
	for _, msg := range tx.Messages() {
		if msg.App >= APP_CONTRACT_TPL_REQUEST && msg.App <= APP_CONTRACT_STOP_REQUEST {
			return msg.App
		}
	}
	log.Debugf("GetContractTxType, not contract Tx, txHash[%s]", tx.Hash().String())
	return APP_UNKNOW
}

// 获取locktime
func (tx *Transaction) GetLocktime() int64 {
	for _, msgcopy := range tx.Messages() {
		if msgcopy.App != APP_PAYMENT {
			continue
		}
		if msg, ok := msgcopy.Payload.(*PaymentPayload); ok {
			return int64(msg.LockTime)
		}
	}
	return 0
}

func (tx *Transaction) String() string {
	data, err := json.Marshal(tx)
	if err != nil {
		log.Errorf("tx[%s] Marshal error:%s", tx.Hash(), err.Error())
		return ""
	}
	return string(data)
}

type Addition struct {
	Addr   common.Address `json:"address"`
	Amount uint64         `json:"amount"`
	Asset  *Asset         `json:"asset"`
}

type OutPoint struct {
	TxHash       common.Hash `json:"txhash"`        // reference Utxo struct key field
	MessageIndex uint32      `json:"message_index"` // message index in transaction
	OutIndex     uint32      `json:"out_index"`
}

func (outpoint *OutPoint) String() string {
	return fmt.Sprintf("Outpoint[TxId:{%#x},MsgIdx:{%d},OutIdx:{%d}]",
		outpoint.TxHash, outpoint.MessageIndex, outpoint.OutIndex)
}
func (outpoint *OutPoint) Clone() *OutPoint {
	return NewOutPoint(outpoint.TxHash, outpoint.MessageIndex, outpoint.OutIndex)
}
func NewOutPoint(hash common.Hash, messageindex uint32, outindex uint32) *OutPoint {
	return &OutPoint{
		TxHash:       hash,
		MessageIndex: messageindex,
		OutIndex:     outindex,
	}
}

// VarIntSerializeSize returns the number of bytes it would take to serialize
// val as a variable length integer.
func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}
	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}
	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}
	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *Output) SerializeSize() int {
	// Value 8 bytes + serialized varint size for the length of PkScript +
	// PkScript bytes.
	return 8 + VarIntSerializeSize(uint64(len(t.PkScript))) + len(t.PkScript)
}
func (t *Input) SerializeSize() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Sequence 4 bytes +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 40 + VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}

func (msg *Transaction) SerializeSize() int {
	n := msg.baseSize()
	return n
}
func (tx *Transaction) DataPayloadSize() int {
	size := 0
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_DATA {
			data := msg.Payload.(*DataPayload)
			size += len(data.MainData) + len(data.ExtraData) + len(data.Reference)
		}
	}
	return size
}

//Deep copy transaction to a new object
func (tx *Transaction) Clone() *Transaction {
	newTx := new(Transaction)
	data, _ := rlp.EncodeToBytes(tx)
	rlp.DecodeBytes(data, newTx)

	return newTx
}

// SerializeNoWitness encodes the transaction to w in an identical manner to
// Serialize, however even if the source transaction has inputs with witness
// data, the old serialization format will still be used.
func (msg *PaymentPayload) SerializeNoWitness(w io.Writer) error {
	//return msg.BtcEncode(w, 0, BaseEncoding)
	return nil
}

func (msg *Transaction) baseSize() int {
	b, _ := rlp.EncodeToBytes(msg)
	return len(b)
}

//是否是合约交易
func (tx *Transaction) IsContractTx() bool {
	for _, m := range tx.txdata.TxMessages {
		if m.App >= APP_CONTRACT_TPL_REQUEST && m.App <= APP_CONTRACT_STOP_REQUEST {
			return true
		}
	}
	return false
}

//是否是系统合约调用，只有在具有InvokeRequest或者TemplateRequest的时候才算系统合约
func (tx *Transaction) IsSystemContract() bool {
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			contractId := msg.Payload.(*ContractInvokeRequestPayload).ContractId
			contractAddr := common.NewAddress(contractId, common.ContractHash)
			//log.Debug("isSystemContract", "contract id", contractAddr, "len", len(contractAddr))
			return contractAddr.IsSystemContractAddress() //, nil

		}
	}
	return false //没有Request，当然就不是系统合约
}

//是否是一个用户合约的交易
func (tx *Transaction) IsUserContract() bool {
	for _, msg := range tx.txdata.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			contractId := msg.Payload.(*ContractInvokeRequestPayload).ContractId
			contractAddr := common.NewAddress(contractId, common.ContractHash)
			//log.Debug("isSystemContract", "contract id", contractAddr, "len", len(contractAddr))
			return !contractAddr.IsSystemContractAddress() //, nil

		} else if msg.App == APP_CONTRACT_DEPLOY_REQUEST || msg.App == APP_CONTRACT_STOP_REQUEST {
			return true //只有用户合约才有deploy和stop
		}
	}
	return false //没有Request，当然就不是合约
}

//判断一个交易是否是一个合约请求交易，并且还没有被执行
func (tx *Transaction) IsOnlyContractRequest() bool {
	if tx.txdata.TxMessages[0].App == APP_PAYMENT {
		lastMsg := tx.txdata.TxMessages[len(tx.txdata.TxMessages)-1]
		return lastMsg.App.IsRequest()
	}
	//no gas fee 判断第一个SignaturePayload是不是最后一个Message
	hasRequestPayload := false
	for i, msg := range tx.txdata.TxMessages {
		if msg.App.IsRequest() {
			hasRequestPayload = true
		}
		if msg.App == APP_SIGNATURE && hasRequestPayload {
			if i == len(tx.txdata.TxMessages)-1 {
				return true
			} else { //SignaturePayload下面还有其他Message
				return false
			}
		}
	}
	return false
}

//获得合约请求Msg的Index
func (tx *Transaction) GetContractInvokeReqMsgIdx() int {
	for idx, msg := range tx.txdata.TxMessages {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST {
			return idx
		}
	}
	return -1
}

//之前的费用分配有Bug，在ContractInstall的时候会分配错误。在V2中解决了这个问题，但是由于测试网已经有历史数据了，所以需要保留历史计算方法。
func (tx *Transaction) GetTxFeeAllocateLegacyV1(queryUtxoFunc QueryUtxoFunc, getSignerFunc GetScriptSignersFunc,
	mediatorReward common.Address) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc)
	result := make([]*Addition, 0)
	if err != nil {
		return nil, err
	}
	if fee.Amount == 0 {
		return result, nil
	}

	isResultMsg := false
	jury := []common.Address{}
	for msgIdx, msg := range tx.TxMessages() {
		if msg.App.IsRequest() {
			isResultMsg = true
			continue
		}
		if isResultMsg && msg.App == APP_SIGNATURE {
			payload := msg.Payload.(*SignaturePayload)
			for _, sig := range payload.Signatures {
				jury = append(jury, crypto.PubkeyBytesToAddress(sig.PubKey))
			}
		}
		if isResultMsg && msg.App == APP_PAYMENT {
			payment := msg.Payload.(*PaymentPayload)
			if !payment.IsCoinbase() {
				jury, err = getSignerFunc(tx, msgIdx, 0)
				if err != nil {
					return nil, errors.New("Parse unlock script to get signers error:" + err.Error())
				}
			}
		}
	}

	juryAllocatedAmt := uint64(0)
	if isResultMsg { //合约执行，Fee需要分配给Jury
		juryAmount := float64(fee.Amount) * parameter.CurrentSysParameters.ContractFeeJuryPercent
		juryCount := float64(len(jury))
		for _, juror := range jury {
			jIncome := &Addition{
				Addr:   juror,
				Amount: uint64(juryAmount / juryCount),
				Asset:  fee.Asset,
			}
			juryAllocatedAmt += jIncome.Amount
			result = append(result, jIncome)
		}
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount - juryAllocatedAmt,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
		//} else { //没有合约执行，全部分配给Mediator
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
	}

	mediatorIncome := &Addition{
		Addr:   mediatorReward,
		Amount: fee.Amount - juryAllocatedAmt,
		Asset:  fee.Asset,
	}
	result = append(result, mediatorIncome)

	return result, nil
}

//获得一笔交易的手续费分配情况,包括Mediator的打包费，Juror的合约执行费
func (tx *Transaction) GetTxFeeAllocate(queryUtxoFunc QueryUtxoFunc, getSignerFunc GetScriptSignersFunc,
	mediatorReward common.Address, getJurorRewardFunc GetJurorRewardAddFunc) ([]*Addition, error) {
	fee, err := tx.GetTxFee(queryUtxoFunc)
	result := make([]*Addition, 0)
	if err != nil {
		return nil, err
	}
	if fee.Amount == 0 {
		return result, nil
	}

	isJuryInside := false
	jury := []common.Address{}
	for msgIdx, msg := range tx.TxMessages() {
		if msg.App == APP_CONTRACT_INVOKE_REQUEST ||
			msg.App == APP_CONTRACT_DEPLOY_REQUEST ||
			msg.App == APP_CONTRACT_STOP_REQUEST {
			isJuryInside = true
			//只有合约部署和调用的时候会涉及到Jury，才会分手续费给Jury
			continue
		}
		if isJuryInside && msg.App == APP_SIGNATURE {
			payload := msg.Payload.(*SignaturePayload)
			for _, sig := range payload.Signatures {
				jury = append(jury, crypto.PubkeyBytesToAddress(sig.PubKey))
			}
		}
		if isJuryInside && msg.App == APP_PAYMENT {
			payment := msg.Payload.(*PaymentPayload)
			if !payment.IsCoinbase() {
				jury, err = getSignerFunc(tx, msgIdx, 0)
				if err != nil {
					return nil, errors.New("Parse unlock script to get signers error:" + err.Error())
				}
			}
		}
	}

	juryAllocatedAmt := uint64(0)
	if isJuryInside { //合约执行，Fee需要分配给Jury
		juryAmount := float64(fee.Amount) * parameter.CurrentSysParameters.ContractFeeJuryPercent
		juryCount := float64(len(jury))
		for _, juror := range jury {
			jIncome := &Addition{
				Addr:   getJurorRewardFunc(juror),
				Amount: uint64(juryAmount / juryCount),
				Asset:  fee.Asset,
			}
			juryAllocatedAmt += jIncome.Amount
			result = append(result, jIncome)
		}
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount - juryAllocatedAmt,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
		//} else { //没有合约部署或者执行，全部分配给Mediator
		//	mediatorIncome := &Addition{
		//		Addr:   mediatorAddr,
		//		Amount: fee.Amount,
		//		Asset:  fee.Asset,
		//	}
		//	result = append(result, mediatorIncome)
	}

	mediatorIncome := &Addition{
		Addr:   mediatorReward,
		Amount: fee.Amount - juryAllocatedAmt,
		Asset:  fee.Asset,
	}
	result = append(result, mediatorIncome)

	return result, nil
}

// SerializeSizeStripped returns the number of bytes it would take to serialize
// the transaction, excluding any included witness data.
func (tx *Transaction) SerializeSizeStripped() int {
	return tx.baseSize()
}

func (a *Addition) IsEqualStyle(b *Addition) (bool, error) {
	if b == nil {
		return false, errors.New("Addition isEqual err, param is nil")
	}
	if a.Addr == b.Addr && a.Asset == b.Asset {
		return true, nil
	}
	return false, nil
}
func (a *Addition) Key() string {
	if a.Asset != nil {
		return hex.EncodeToString(append(a.Addr.Bytes21(), a.Asset.Bytes()...))
	}
	return hex.EncodeToString(a.Addr.Bytes21())
}

//传入一堆交易，按依赖关系进行排序，并根据UTXO的使用情况，分为3类Tx：
//1.排序后的正常交易，2.孤儿交易，3.因为双花需要丢弃的交易
func SortTxs(txs map[common.Hash]*Transaction, utxoFunc QueryUtxoFunc) ([]*Transaction, []*Transaction, []*Transaction) {
	sortedTxs := make([]*Transaction, 0)
	doubleSpendTxs := make([]*Transaction, 0)
	orphanTxs := make([]*Transaction, 0)
	map_orphans := make(map[common.Hash]common.Hash)
	map_pretxs := make(map[common.Hash]int)
	map_doubleTxs := make(map[*OutPoint]common.Hash)
	//map_utxos := make(map[*OutPoint]common.Hash)
	for hash, tx := range txs {
		ops := tx.GetSpendOutpoints()
		isOrphan := false
		hasDouble := false
		for _, op := range ops {
			if _, has := map_orphans[op.TxHash]; has {
				map_orphans[hash] = op.TxHash
				orphanTxs = append(orphanTxs, tx)
				isOrphan = true
				break
			}
			if tx.isOrphanTx(txs, utxoFunc) {
				map_orphans[hash] = op.TxHash
				orphanTxs = append(orphanTxs, tx)
				isOrphan = true
				break
			}
			// 在双花交易中择优选择一个交易作为有效交易。
			if d_hash, has := map_doubleTxs[op]; has {
				hasDouble = true
				p_tx := txs[d_hash]
				if isPrefer, err := tx.preferTx(p_tx, utxoFunc); err == nil {
					if isPrefer {
						// delete p_tx
						temp := make([]*Transaction, 0)
						for i, stx := range sortedTxs {
							if stx.Hash() == p_tx.Hash() {
								temp = sortedTxs[:i]
								temp = append(temp, sortedTxs[i+1:]...)
								break
							}
						}
						temp = append(temp, tx)
						sortedTxs = temp[:]

						doubleSpendTxs = append(doubleSpendTxs, p_tx)
						for _, op := range tx.GetSpendOutpoints() {
							map_doubleTxs[op] = tx.Hash()
						}
					} else {
						doubleSpendTxs = append(doubleSpendTxs, tx)
					}
				}
				continue
			}
			map_doubleTxs[op] = hash
		}
		if !isOrphan && !hasDouble {
			pre_txs := tx.GetPrecusorTxs(txs)
			for _, tx := range pre_txs {
				if _, has := map_pretxs[tx.Hash()]; !has {
					map_pretxs[tx.Hash()] = len(sortedTxs)
					//fmt.Println("add sorted tx:", tx.Hash().String())
					sortedTxs = append(sortedTxs, tx)
				}
			}
		}
	}

	//for i, tx := range sortedTxs {
	//	fmt.Println("sorted tx:", i, tx.Hash().String())
	//}
	//for i, tx := range orphanTxs {
	//	fmt.Println("orphan tx:", i, tx.Hash().String())
	//}
	//for i, tx := range doubleSpendTxs {
	//	fmt.Println("double spend tx:", i, tx.Hash().String())
	//}

	return sortedTxs, orphanTxs, doubleSpendTxs
}

func (tx *Transaction) GetPrecusorTxs(poolTxs map[common.Hash]*Transaction) []*Transaction {
	pretxs := make([]*Transaction, 0)
	for _, msg := range tx.Messages() {
		if msg.App == APP_PAYMENT {
			payment, ok := msg.Payload.(*PaymentPayload)
			if ok {
				for _, input := range payment.Inputs {
					if input.PreviousOutPoint != nil {
						sort_tx, has := poolTxs[input.PreviousOutPoint.TxHash]
						isRequest := false
						if !has {
							for _, sort_tx = range poolTxs {
								if sort_tx.RequestHash() == input.PreviousOutPoint.TxHash {
									isRequest = true
									break
								}
							}
							if !isRequest {
								continue
							}
						}
						if sort_tx != nil {
							list := sort_tx.GetPrecusorTxs(poolTxs)
							if len(list) > 0 {
								pretxs = append(pretxs, list...)
							}
							pretxs = append(pretxs, sort_tx)
						}
					}
				}
			}
		}
	}
	//返回自己
	pretxs = append(pretxs, tx)
	return pretxs
}
func (tx *Transaction) isOrphanTx(txs map[common.Hash]*Transaction, utxoFunc QueryUtxoFunc) bool {
	for _, op := range tx.GetSpendOutpoints() {
		if _, err := utxoFunc(op); err == nil {
			continue
		}
		// db里没有该utxo,则依次从tx的前驱交易里找,如果列表里没有前驱交易返回true
		if p_tx, has := txs[op.TxHash]; !has {
			for _, otx := range txs {
				if otx.RequestHash() == op.TxHash {
					// 若requeshash等于op的txhash,则从otx的output里找utxo
					isfound := false
					for i, msg := range otx.txdata.TxMessages {
						if msg.App == APP_PAYMENT {
							payment := msg.Payload.(*PaymentPayload)
							for j := range payment.Outputs {
								if op.OutIndex == uint32(j) && op.MessageIndex == uint32(i) {
									isfound = true
									break
								}
							}
						}
					}
					if !isfound {
						return true
					}
					return false
				}
			}
			return true
		} else {
			if p_tx.isOrphanTx(txs, utxoFunc) {
				return true
			} else {
				// 找到该utxo
				isfound := false
				for i, msg := range p_tx.txdata.TxMessages {
					if msg.App == APP_PAYMENT {
						payment := msg.Payload.(*PaymentPayload)
						for j := range payment.Outputs {
							if op.OutIndex == uint32(j) && op.MessageIndex == uint32(i) {
								isfound = true
								break
							}
						}
					}
				}
				if !isfound {
					return true
				} else {
					continue
				}
			}
		}
	}
	return false
}

func (tx *Transaction) preferTx(tx1 *Transaction, utxoFunc QueryUtxoFunc) (bool, error) {
	fee, err := tx.GetTxFee(utxoFunc)
	fee1, err1 := tx1.GetTxFee(utxoFunc)
	if err != nil && err1 != nil {
		return false, fmt.Errorf("all utxos are illegal.")
	}
	if err != nil {
		return false, nil
	}
	if err1 != nil {
		return true, nil
	}
	if fee.Asset.String() != fee1.Asset.String() {
		return false, fmt.Errorf("this two transaction asset is different.")
	}
	if fee1.Amount > fee.Amount {
		return false, nil
	}
	return true, nil

}
