/*
 *  This file is part of go-palletone.
 *  go-palletone is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *  go-palletone is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *  You should have received a copy of the GNU General Public License
 *  along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 *
 *  @author PalletOne core developer <dev@pallet.one>
 *  @date 2018-2020
 */

package txpool2

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
	"github.com/palletone/go-palletone/validator"
)

var (
	ErrNotFound    = errors.New("txpool: not found")
	ErrDoubleSpend = errors.New("txpool: double spend")
	ErrNotSupport  = errors.New("txpool: not support")
)

var Instance txspool.ITxPool

type TxPool struct {
	normals              *txList                                    //普通交易池
	orphans              map[common.Hash]*txspool.TxPoolTransaction //孤儿交易池
	userContractRequests map[common.Hash]*txspool.TxPoolTransaction //用户合约请求，只参与utxo运算，不会被打包
	txValidator          txspool.IValidator
	dag                  txspool.IDag
	tokenengine          tokenengine.ITokenEngine
	sync.RWMutex
	txFeed event.Feed
	scope  event.SubscriptionScope
}

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewTxPool(config txspool.TxPoolConfig, cachedb palletcache.ICache, unit txspool.IDag) *TxPool {
	tokenEngine := tokenengine.Instance
	val := validator.NewValidate(unit, unit, unit, unit, nil, cachedb, false)
	pool := NewTxPool4DI(config, cachedb, unit, tokenEngine, val)
	//pool.startJournal(config)
	return pool
}

func NewTxPool4DI(config txspool.TxPoolConfig, cachedb palletcache.ICache, dag txspool.IDag,
	tokenEngine tokenengine.ITokenEngine, txValidator txspool.IValidator) *TxPool {
	return &TxPool{
		normals:              newTxList(),
		orphans:              make(map[common.Hash]*txspool.TxPoolTransaction),
		userContractRequests: make(map[common.Hash]*txspool.TxPoolTransaction),
		txValidator:          txValidator,
		dag:                  dag,
		tokenengine:          tokenEngine,
	}
}

//支持合约Request，普通FullTx，用户合约FullTx的加入，不支持系统合约FullTx
func (pool *TxPool) AddLocal(tx *modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		data, _ := rlp.EncodeToBytes(tx)
		return fmt.Sprintf("try to add tx[%s] to txpool, tx hex:%x", tx.Hash().String(), data)
	})
	err := pool.addLocal(tx)
	if err != nil {
		return err
	}

	return nil
}
func (pool *TxPool) AddRemote(tx *modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		data, _ := rlp.EncodeToBytes(tx)
		return fmt.Sprintf("try to add tx[%s] to txpool, tx hex:%x", tx.Hash().String(), data)
	})
	err := pool.addLocal(tx)
	if err != nil {
		return err
	}

	return nil
}
func (pool *TxPool) addLocal(tx *modules.Transaction) error {
	//check duplicate add
	if _, err := pool.normals.GetTx(tx.Hash()); err == nil { //found tx
		log.Info("try to add duplicate tx[%s] to tx pool", tx.Hash().String())
		return nil
	}
	if _, ok := pool.orphans[tx.Hash()]; ok { //found in orphans
		log.Info("try to add duplicate tx[%s] to tx pool", tx.Hash().String())
		return nil
	}

	if tx.IsSystemContract() && !tx.IsOnlyContractRequest() && tx.GetContractTxType() != modules.APP_CONTRACT_TPL_REQUEST {
		log.Infof("tx[%s] is a full system contract invoke tx, don't support", tx.Hash().String())
		return ErrNotSupport
	}

	//1.validate tx
	pool.txValidator.SetUtxoQuery(pool)
	fee, vcode, err := pool.txValidator.ValidateTx(tx, false)
	if err != nil {
		log.Warnf("validate tx[%s] get error:%s", tx.Hash().String(), err.Error())
		return err
	}
	tx2 := pool.convertTx(tx, fee)
	//2. process orphan
	if vcode == validator.TxValidationCode_ORPHAN {
		return pool.addOrphanTx(tx2)
	}
	if tx.IsUserContract() && tx.IsOnlyContractRequest() {
		log.Debugf("tx[%s] is an user contract invoke request", tx.Hash().String())
		pool.userContractRequests[tx2.TxHash] = tx2
		return nil
	}
	//3. process normal tx
	err = pool.normals.AddTx(tx2)
	if err != nil {
		log.Errorf("add tx[%s] to normal pool error:%s", tx2.TxHash.String(), err.Error())
		return err
	}
	pool.txFeed.Send(modules.TxPreEvent{Tx: tx})
	//4. check orphan txpool
	return pool.checkOrphanTxToNormal(tx2.TxHash)
}

//检查如果将一个Tx加入Normal后，有没有后续的孤儿Tx需要连带加入
func (pool *TxPool) checkOrphanTxToNormal(txHash common.Hash) error {
	readyTx := []*modules.Transaction{}
	for hash, otx := range pool.orphans {
		if otx.IsFineToNormal(txHash) { //满足Normal的条件了
			log.Debugf("move tx[%s] from orphans to normals", otx.TxHash.String())
			delete(pool.orphans, hash) //从孤儿池删除
			readyTx = append(readyTx, otx.Tx)
			//otx.Status= TxPoolTxStatus_Unpacked
			//err := pool.normals.AddTx(otx)
			//if err != nil {
			//	log.Errorf("add tx[%s] to normal pool error:%s", otx.TxHash.String(), err.Error())
			//	return err
			//}
			//return pool.checkOrphanTxToNormal(otx.TxHash)
		}
	}
	for _, tx := range readyTx {
		err := pool.addLocal(tx) //因为之前孤儿交易没有手续费，UTXO等，所以需要重新计算
		if err != nil {
			log.Warnf("add tx[%s] to pool fail:%s", tx.Hash().String(), err.Error())
		}
	}
	return nil
}

func (pool *TxPool) convertTx(tx *modules.Transaction, fee []*modules.Addition) *txspool.TxPoolTransaction {
	fromAddr, _ := tx.GetFromAddrs(pool.GetUtxoEntry, pool.tokenengine.GetAddressFromScript)
	dependOnTxs := make(map[common.Hash]bool)
	for _, o := range tx.GetSpendOutpoints() {
		dependOnTxs[o.TxHash] = false
	}
	txAddr, _ := tx.GetToAddrs(pool.tokenengine.GetAddressFromScript)

	return &txspool.TxPoolTransaction{
		Tx:                   tx,
		TxHash:               tx.Hash(),
		ReqHash:              tx.RequestHash(),
		TxFee:                fee,
		CreationDate:         time.Now(),
		FromAddr:             fromAddr,
		DependOnTxs:          dependOnTxs,
		From:                 tx.GetSpendOutpoints(),
		ToAddr:               txAddr,
		IsSysContractRequest: tx.IsOnlyContractRequest() && tx.IsSystemContract(),
		IsUserContractFullTx: tx.IsUserContract() && !tx.IsOnlyContractRequest(),
	}
}
func (pool *TxPool) addOrphanTx(tx *txspool.TxPoolTransaction) error {
	log.Debugf("add tx[%s] to orphan pool", tx.TxHash.String())
	tx.Status = txspool.TxPoolTxStatus_Orphan
	pool.orphans[tx.TxHash] = tx
	return nil
}
func (pool *TxPool) GetSortedTxs(processor func(transaction *txspool.TxPoolTransaction) (getNext bool, err error)) error {
	pool.RLock()
	defer pool.RUnlock()
	return pool.normals.GetSortedTxs(processor)
}

//带锁的对外暴露的查询
func (pool *TxPool) GetUtxo(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	pool.RLock()
	defer pool.RUnlock()
	return pool.GetUtxoEntry(outpoint)
}

//主要用于Validator，不带锁
func (pool *TxPool) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	poolUtxo, err := pool.normals.GetUtxoEntry(outpoint)
	if err != nil {
		if len(pool.userContractRequests) > 0 {
			reqUtxo, err := getUtxoFromTxs(pool.userContractRequests, outpoint)
			if err == nil {
				return reqUtxo, nil
			}
		}
		//log.Warnf("GetUtxoEntry(%s) not found in pool",outpoint.String())
		log.DebugDynamic(func() string {
			return fmt.Sprintf("GetUtxoEntry(%s) not found in pool", outpoint.String())
		})
		return pool.dag.GetUtxoEntry(outpoint)
		//return nil,ErrNotFound

	}
	return poolUtxo, nil
}

//func (pool *TxPool) GetUtxoFromPoolAndDag(outpoint *modules.OutPoint) (*modules.Utxo, error) {
//	utxo,err:=pool.GetUtxoEntry(outpoint)
//	if err!=nil{
//		return pool.dag.GetUtxoEntry(outpoint)
//	}
//	return utxo,nil
//}

func getUtxoFromTxs(txs map[common.Hash]*txspool.TxPoolTransaction, outpoint *modules.OutPoint) (*modules.Utxo, error) {
	newUtxo := make(map[modules.OutPoint]*modules.Utxo)
	spendUtxo := make(map[modules.OutPoint]bool)
	for _, tx := range txs {
		for _, o := range tx.Tx.GetSpendOutpoints() {
			spendUtxo[*o] = true
		}
		for o, u := range tx.Tx.GetNewUtxos() {
			newUtxo[o] = u
		}
	}
	if _, ok := spendUtxo[*outpoint]; ok {
		return nil, ErrDoubleSpend
	}
	if utxo, ok := newUtxo[*outpoint]; ok {
		return utxo, nil
	}
	return nil, ErrNotFound
}

func (pool *TxPool) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	pool.RLock()
	defer pool.RUnlock()
	return pool.dag.GetStxoEntry(outpoint)
}

func (pool *TxPool) DiscardTxs(txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("discard txs: %s", hashes)
	})
	if pool.normals.Count() == 0 {
		return nil
	}
	for _, tx := range txs {
		if tx.IsContractTx() {
			err := pool.normals.DiscardTx(tx.RequestHash())
			if err != nil {
				if err == ErrNotFound {
					continue
				} else {
					return err
				}
			}
			delete(pool.orphans, tx.RequestHash())
		}
		err := pool.normals.DiscardTx(tx.Hash())
		if err != nil {
			if err == ErrNotFound {
				continue
			} else {
				return err
			}
		}
		delete(pool.orphans, tx.Hash())
	}
	return nil
}

func (pool *TxPool) GetUnpackedTxsByAddr(addr common.Address) ([]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	txs, err := pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Unpacked)
	if err != nil {
		return nil, err
	}
	result := []*txspool.TxPoolTransaction{}
	for _, tx := range txs {
		if tx.IsFrom(addr) || tx.IsTo(addr) {
			result = append(result, tx)
		}
	}
	return result, nil
}

//func (pool *TxPool) GetUnpackedTxs() (map[common.Hash]*txspool.TxPoolTransaction, error) {
//	return pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Unpacked)
//}
func (pool *TxPool) Pending() (map[common.Hash][]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	packedTxs, err := pool.normals.GetTxsByStatus(txspool.TxPoolTxStatus_Packed)
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash][]*txspool.TxPoolTransaction)
	for _, tx := range packedTxs {
		if txs, ok := result[tx.UnitHash]; ok {
			result[tx.UnitHash] = append(txs, tx)
		} else {
			result[tx.UnitHash] = []*txspool.TxPoolTransaction{tx}
		}
	}
	return result, nil
}
func (pool *TxPool) Queued() ([]*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	result := []*txspool.TxPoolTransaction{}
	for _, tx := range pool.orphans {
		result = append(result, tx)
	}
	return result, nil
}
func (pool *TxPool) Stop() {
	pool.scope.Close()
	log.Info("Transaction pool stopped")
}

//基本状态(未打包，已打包，孤儿)
func (pool *TxPool) Status() (int, int, int) {
	pool.RLock()
	defer pool.RUnlock()
	normals := pool.normals.GetAllTxs()
	packed := 0
	unpacked := 0
	for _, tx := range normals {
		if tx.Status == txspool.TxPoolTxStatus_Packed {
			packed++
		}
		if tx.Status == txspool.TxPoolTxStatus_Unpacked {
			unpacked++
		}
	}
	return unpacked, packed, len(pool.orphans)
}
func (pool *TxPool) Content() (map[common.Hash]*txspool.TxPoolTransaction, map[common.Hash]*txspool.TxPoolTransaction) {
	pool.RLock()
	defer pool.RUnlock()
	return pool.normals.GetAllTxs(), pool.orphans
}

//将交易状态改为已打包
func (pool *TxPool) SetPendingTxs(unit_hash common.Hash, num uint64, txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("update status to packed txs: %s", hashes)
	})
	if pool.normals.Count() == 0 {
		return nil
	}
	for _, tx := range txs {
		if tx.IsContractTx() {
			err := pool.normals.UpdateTxStatusPacked(tx.RequestHash(), unit_hash, num)
			if err != nil && err != ErrNotFound {
				return err
			}
		}
		err := pool.normals.UpdateTxStatusPacked(tx.Hash(), unit_hash, num)
		if err != nil && err != ErrNotFound {
			return err
		}
	}
	return nil
}

//将交易状态改为未打包
func (pool *TxPool) ResetPendingTxs(txs []*modules.Transaction) error {
	pool.Lock()
	defer pool.Unlock()
	log.DebugDynamic(func() string {
		hashes := ""
		for _, tx := range txs {
			hashes += tx.Hash().String() + ";"
		}
		return fmt.Sprintf("update status to unpacked txs: %s", hashes)
	})
	if pool.normals.Count() == 0 {
		return nil
	}
	for _, tx := range txs {
		if tx.IsContractTx() {
			err := pool.normals.UpdateTxStatusUnpacked(tx.RequestHash())
			if err != nil && err != ErrNotFound {
				return err
			}
		}
		err := pool.normals.UpdateTxStatusUnpacked(tx.Hash())
		if err != nil && err != ErrNotFound {
			return err
		}
	}
	return nil
}
func (pool *TxPool) GetTx(hash common.Hash) (*txspool.TxPoolTransaction, error) {
	pool.RLock()
	defer pool.RUnlock()
	tx, err := pool.normals.GetTx(hash)
	if err == ErrNotFound {
		tx, ok := pool.orphans[hash]
		if ok {
			return tx, nil
		}
		return nil, ErrNotFound
	}
	return tx, err
}

// SubscribeTxPreEvent registers a subscription of TxPreEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeTxPreEvent(ch chan<- modules.TxPreEvent) event.Subscription {
	//return pool.txFeed.Subscribe(ch)
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}
