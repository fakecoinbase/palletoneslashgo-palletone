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
 * @date 2018-2020
 */

package rwset

import (
	"sync"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
)

type RWSetBuilder struct {
	pubRwBuilderMap map[string]*nsPubRwBuilder
	locker          sync.RWMutex
}

type nsPubRwBuilder struct {
	namespace   string
	readMap     map[string]map[string]*KVRead //map[contractId]map[key]*KVRead
	writeMap    map[string]map[string]*KVWrite
	tokenPayOut []*modules.TokenPayOut
	tokenSupply []*modules.TokenSupply
	tokenDefine *modules.TokenDefine
}

func NewRWSetBuilder() *RWSetBuilder {
	return &RWSetBuilder{
		pubRwBuilderMap: make(map[string]*nsPubRwBuilder),
		locker:          sync.RWMutex{},
	}
}

func (b *RWSetBuilder) AddToReadSet(contractId []byte, ns string, key string, version *modules.StateVersion) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	if nsPubRwBuilder.readMap == nil {
		nsPubRwBuilder.readMap = make(map[string]map[string]*KVRead)
	} else {
		kv, ok := nsPubRwBuilder.readMap[string(contractId)]
		if !ok {
			kv = make(map[string]*KVRead)
			nsPubRwBuilder.readMap[string(contractId)] = kv
		}
		// ReadSet
		kv[key] = NewKVRead(contractId, key, version)
	}
}
func (b *RWSetBuilder) AddTokenPayOut(ns string, address common.Address, asset *modules.Asset, amount uint64, lockTime uint32) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	if nsPubRwBuilder.tokenPayOut == nil {
		nsPubRwBuilder.tokenPayOut = []*modules.TokenPayOut{}
	}
	pay := &modules.TokenPayOut{Asset: asset, Amount: amount, PayTo: address, LockTime: lockTime}
	nsPubRwBuilder.tokenPayOut = append(nsPubRwBuilder.tokenPayOut, pay)

}
func (b *RWSetBuilder) AddToWriteSet(contractId []byte, ns string, key string, value []byte) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	if nsPubRwBuilder.writeMap == nil {
		nsPubRwBuilder.writeMap = make(map[string]map[string]*KVWrite)
	}
	kv, ok := nsPubRwBuilder.writeMap[string(contractId)]
	if !ok {
		kv = make(map[string]*KVWrite)
		nsPubRwBuilder.writeMap[string(contractId)] = kv
	}
	kv[key] = newKVWrite(contractId, key, value)
}
func (b *RWSetBuilder) GetWriteSet(contractId []byte, key string) ([]byte, error) {
	for _, builder := range b.pubRwBuilderMap {
		if kv, ok := builder.writeMap[string(contractId)]; ok {
			if value, ok2 := kv[key]; ok2 {
				if value.isDelete {
					return nil, nil
				}
				return value.value, nil
			}
		}
	}
	return nil, errors.ErrNotFound
}
func (b *RWSetBuilder) GetWriteSets(contractId []byte) ([]*KVWrite, error) {
	result := make([]*KVWrite, 0, 0)
	for _, builder := range b.pubRwBuilderMap {
		if kv, ok := builder.writeMap[string(contractId)]; ok {
			for _, v := range kv {
				result = append(result, v)
			}
		}
	}
	return result, nil
}
func (b *RWSetBuilder) GetTokenPayOut(ns string) []*modules.TokenPayOut {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)

	return nsPubRwBuilder.tokenPayOut
}
func (b *RWSetBuilder) GetTokenDefine(ns string) *modules.TokenDefine {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	return nsPubRwBuilder.tokenDefine
}
func (b *RWSetBuilder) GetTokenSupply(ns string) []*modules.TokenSupply {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	tokenSupply := make([]*modules.TokenSupply, 0)
	if nsPubRwBuilder.tokenSupply == nil {
		nsPubRwBuilder.tokenSupply = tokenSupply
	}
	// 上层对tokenSupply的更改不影响nsPubRwBuilder原值。
	tokenSupply = append(tokenSupply, nsPubRwBuilder.tokenSupply...)
	return tokenSupply
}
func (b *RWSetBuilder) DefineToken(ns string, tokenType int32, define []byte, createAddr common.Address) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	nsPubRwBuilder.tokenDefine = &modules.TokenDefine{TokenType: int(tokenType),
		TokenDefineJson: define, Creator: createAddr}
}
func (b *RWSetBuilder) AddSupplyToken(ns string, assetId, uniqueId []byte, amt uint64,
	createAddr common.Address) error {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	if nsPubRwBuilder.tokenSupply == nil {
		nsPubRwBuilder.tokenSupply = make([]*modules.TokenSupply, 0)
	}

	nsPubRwBuilder.tokenSupply = append(nsPubRwBuilder.tokenSupply, &modules.TokenSupply{AssetId: assetId,
		UniqueId: uniqueId, Amount: amt, Creator: createAddr})
	return nil
}

func (b *RWSetBuilder) getOrCreateNsPubRwBuilder(ns string) *nsPubRwBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	nsPubRwBuilder, ok := b.pubRwBuilderMap[ns]
	if !ok {
		nsPubRwBuilder = newNsPubRwBuilder(ns)
		b.pubRwBuilderMap[ns] = nsPubRwBuilder
	}
	return nsPubRwBuilder
}

func newNsPubRwBuilder(namespace string) *nsPubRwBuilder {
	return &nsPubRwBuilder{
		namespace,
		make(map[string]map[string]*KVRead),
		make(map[string]map[string]*KVWrite),
		[]*modules.TokenPayOut{},
		[]*modules.TokenSupply{},
		nil,
	}
}
