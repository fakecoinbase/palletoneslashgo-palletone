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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package dag

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

type Txo4Greedy struct {
	modules.OutPoint
	Amount uint64
}

func (txo *Txo4Greedy) GetAmount() uint64 {
	return txo.Amount
}

func newTxo4Greedy(outPoint modules.OutPoint, amount uint64) *Txo4Greedy {
	return &Txo4Greedy{
		OutPoint: outPoint,
		Amount:   amount,
	}
}

func (dag *Dag) CreateBaseTransaction(from, to common.Address, daoAmount, daoFee uint64) (*modules.Transaction, error) {
	if daoFee == 0 {
		return &modules.Transaction{}, fmt.Errorf("Transaction's fee id zero!")
	}

	// 1. 获取转出账户所有的utxo
	//allUtxos, err := dag.GetAddrUtxos(from)
	coreUtxos, err := dag.GetAddrUtxos(from)
	if err != nil {
		return &modules.Transaction{}, err
	}

	if len(coreUtxos) == 0 {
		return &modules.Transaction{}, fmt.Errorf("%v 's uxto is null!", from.Str())
	}

	// 2. 利用贪心算法得到指定额度的utxo集合
	greedyUtxos := core.Utxos{}
	for outPoint, utxo := range coreUtxos {
		tg := newTxo4Greedy(outPoint, utxo.Amount)
		greedyUtxos = append(greedyUtxos, tg)
	}

	selUtxos, change, err := core.Select_utxo_Greedy(greedyUtxos, daoAmount+daoFee)
	if err != nil {
		return nil, fmt.Errorf("Select utxo err")
	}

	// 3. 构建PaymentPayload的Inputs
	pload := new(modules.PaymentPayload)
	pload.LockTime = 0

	for _, selTxo := range selUtxos {
		tg := selTxo.(*Txo4Greedy)
		txInput := modules.NewTxIn(&tg.OutPoint, []byte{})
		pload.AddTxIn(txInput)
	}

	// 4. 构建PaymentPayload的Outputs
	outAmount := map[common.Address]uint64{}
	outAmount[to] = daoAmount
	if change > 0 {
		outAmount[from] = change
	}

	for addr, amount := range outAmount {
		pkScript := tokenengine.GenerateLockScript(addr)
		txOut := modules.NewTxOut(amount, pkScript, modules.NewPTNAsset())
		pload.AddTxOut(txOut)
	}

	// 5. 构建Transaction
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))

	return tx, nil
}

func (dag *Dag) GetAddrCoreUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	// todo 待优化
	allUtxos, err := dag.GetAddrUtxos(addr)
	if err != nil {
		return nil, err
	}

	coreUtxos := make(map[modules.OutPoint]*modules.Utxo, len(allUtxos))
	for outPoint, utxo := range allUtxos {
		if utxo.Asset.IsSimilar(modules.CoreAsset) {
			continue
		}

		coreUtxos[outPoint] = utxo
	}

	return coreUtxos, nil
}

func (dag *Dag) GenMediatorCreateTx(account common.Address,
	op *modules.MediatorCreateOperation) (*modules.Transaction, error) {
	// 1. 组装 message
	msg := &modules.Message{
		App:     modules.OP_MEDIATOR_CREATE,
		Payload: op,
	}

	// 2. 组装 tx
	fee := dag.CurrentFeeSchedule().MediatorCreateFee
	tx, err := dag.CreateBaseTransaction(account, account, 0, fee)
	if err != nil {
		return nil, err
	}

	tx.TxMessages = append(tx.TxMessages, msg)
	//tx.TxHash = tx.Hash()

	return tx, nil
}
