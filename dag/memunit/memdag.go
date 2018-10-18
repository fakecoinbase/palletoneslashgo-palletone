/*
 *
 *     This file is part of go-palletone.
 *     go-palletone is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *     go-palletone is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *     You should have received a copy of the GNU General Public License
 *     along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package memunit

import (
	"fmt"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	dagCommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

/*********************************************************************/
// TODO MemDag
type MemDag struct {
	//db                ptndb.Database
	dagdb             storage.IDagDb
	unitRep           dagCommon.IUnitRepository
	lastValidatedUnit map[string]common.Hash // the key is asset id
	forkIndex         map[string]*ForkIndex  // the key is asset id
	mainChain         map[string]int         // the key is asset id, value is fork index
	memUnit           *MemUnit
	memSize           uint8
}

func NewMemDag(db storage.IDagDb, unitRep dagCommon.IUnitRepository) *MemDag {
	memdag := MemDag{
		lastValidatedUnit: make(map[string]common.Hash),
		forkIndex:         make(map[string]*ForkIndex),
		memUnit:           InitMemUnit(),
		memSize:           dagconfig.DefaultConfig.MemoryUnitSize,
		dagdb:             db,
		unitRep:           unitRep,
		mainChain:         make(map[string]int),
	}
	// get genesis Last Irreversible Unit
	genesisUnit, err := unitRep.GetGenesisUnit(0)
	if err != nil {
		log.Error("NewMemDag when GetGenesisUnit", "error", err.Error())
		return nil
	}
	if genesisUnit == nil {
		log.Error("Get genesis unit failed, unit of genesis is nil.")
		return nil
	}
	lastIrreUnit, _ := db.GetLastIrreversibleUnit(genesisUnit.UnitHeader.Number.AssetID)
	if lastIrreUnit != nil {
		memdag.lastValidatedUnit[genesisUnit.UnitHeader.Number.AssetID.String()] = lastIrreUnit.UnitHash
	}

	return &memdag
}

func (chain *MemDag) validateMemory() bool {
	length := chain.memUnit.Lenth()
	//log.Info("MemDag", "validateMemory unit length:", length, "chain.memSize:", chain.memSize)
	if length >= uint64(chain.memSize) {
		return false
	}
	return true
}

func (chain *MemDag) Save(unit *modules.Unit) error {
	if unit == nil {
		return fmt.Errorf("Save mem unit: unit is null")
	}
	if chain.memUnit.Exists(unit.UnitHash) {
		return fmt.Errorf("Save mem unit: unit is already exists in memory")
	}

	//TODO must recover
	//if !chain.validateMemory() {
	//	return fmt.Errorf("Save mem unit: size is out of limit")
	//}

	assetId := unit.UnitHeader.Number.AssetID.String()

	// save fork index
	forkIndex, ok := chain.forkIndex[assetId]
	if !ok {
		// create forindex
		chain.forkIndex[assetId] = &ForkIndex{}
		forkIndex = chain.forkIndex[assetId]
	}

	// get asset chain's las irreversible unit
	irreUnitHash, ok := chain.lastValidatedUnit[assetId]
	if !ok {
		lastIrreUnit, _ := chain.dagdb.GetLastIrreversibleUnit(unit.UnitHeader.Number.AssetID)
		if lastIrreUnit != nil {
			irreUnitHash = lastIrreUnit.UnitHash
			chain.lastValidatedUnit[assetId] = irreUnitHash
		}
	}
	// save unit to index
	index, err := forkIndex.AddData(unit.UnitHash, unit.UnitHeader.ParentsHash)
	switch index {
	case -1:
		return err
	case -2:
		// check last irreversible unit
		// if it is not null, check continuously
		if strings.Compare(irreUnitHash.String(), "") != 0 {
			if common.CheckExists(irreUnitHash, unit.UnitHeader.ParentsHash) < 0 {
				return fmt.Errorf("The unit(%s) is not continious.", unit.UnitHash.String())
			}
		}
		// add new fork into index
		forkData := ForkData{}
		forkData = append(forkData, unit.UnitHash)
		index = len(*forkIndex)
		*forkIndex = append(*forkIndex, &forkData)
	}
	// save memory unit
	if err := chain.memUnit.Add(unit); err != nil {
		return err
	}
	// Check if the irreversible height has been reached
	if forkIndex.IsReachedIrreversibleHeight(index) {
		// set unit irreversible
		unitHash := forkIndex.GetReachedIrreversibleHeightUnitHash(index)
		// prune fork if the irreversible height has been reached
		if err := chain.Prune(assetId, unitHash); err != nil {
			log.Error("Check if the irreversible height has been reached", "error", err.Error())
			return err
		}
		// save the matured unit into leveldb
		if err := chain.unitRep.SaveUnit(unit, false); err != nil {
			log.Error("save the matured unit into leveldb", "error", err.Error())
			return err
		}
	}
	return nil
}

func (chain *MemDag) Exists(uHash common.Hash) bool {
	if chain.memUnit.Exists(uHash) {
		return true
	}
	return false
}

/**
对分叉数据进行剪支
Prune fork data
*/
func (chain *MemDag) Prune(assetId string, maturedUnitHash common.Hash) error {
	// get fork index
	index, subindex := chain.QueryIndex(assetId, maturedUnitHash)
	if index < 0 {
		return fmt.Errorf("Prune error: matured unit is not found in memory")
	}
	// save all the units before matured unit into db
	forkdata := (*(chain.forkIndex[assetId]))[index]
	for i := 0; i < subindex; i++ {
		unitHash := (*forkdata)[i]
		unit := (*chain.memUnit)[unitHash]
		if err := chain.unitRep.SaveUnit(unit, false); err != nil {
			return fmt.Errorf("Prune error when save unit: %s", err.Error())
		}
	}
	// rollback transaction pool

	// refresh forkindex
	if lenth := len(*forkdata); lenth > subindex {
		newForkData := ForkData{}
		for i := subindex + 1; i < lenth; i++ {
			newForkData = append(newForkData, (*forkdata)[i])
		}
		// prune other forks
		newForkindex := ForkIndex{}
		newForkindex = append(newForkindex, &newForkData)
		chain.forkIndex[assetId] = &newForkindex
	}
	// save the matured unit
	chain.lastValidatedUnit[assetId] = maturedUnitHash

	return nil
}

/**
切换主链：将最长链作为主链
Switch to the longest fork
*/
func (chain *MemDag) SwitchMainChain() error {
	// chose the longest fork as the main chain
	for assetid, forkindex := range chain.forkIndex {
		maxLenth := 0
		for index, forkdata := range *forkindex {
			if len(*forkdata) > maxLenth {
				chain.mainChain[assetid] = index
				maxLenth = len(*forkdata)
			}
		}
	}
	return nil
}

func (chain *MemDag) QueryIndex(assetId string, maturedUnitHash common.Hash) (int, int) {
	forkindex, ok := chain.forkIndex[assetId]
	if !ok {
		return -1, -1
	}
	for index, forkdata := range *forkindex {
		for subindex, unitHash := range *forkdata {
			if strings.Compare(unitHash.String(), maturedUnitHash.String()) == 0 {
				return index, subindex
			}
		}
	}
	return -1, -1
}

func (chain *MemDag) GetCurrentUnit(assetid modules.IDType16) (*modules.Unit, error) {
	sAssetID := assetid.String()
	bAssetID, _ := hexutil.Decode(sAssetID)
	fmt.Println("GetCurrentUnit", "assetid", sAssetID, "byte assetid", bAssetID)
	mainIndex, ok := chain.mainChain[sAssetID]
	for k, _ := range chain.lastValidatedUnit {
		fmt.Println("string key=", k)
		bk, _ := hexutil.Decode(k)
		fmt.Println("byte key=", bk)
	}
	if !ok {
		// to get from lastValidatedUnit
		lastValidatedUnitHash, ok := chain.lastValidatedUnit[sAssetID]
		if !ok {
			return nil, nil
		}
		unit, _ := chain.dagdb.GetUnit(lastValidatedUnitHash)
		return unit, nil
	}
	fork, ok := chain.forkIndex[sAssetID]
	if !ok {
		return nil, fmt.Errorf("MemDag.GetCurrentUnit error: forkIndex has no asset(%s) info.", assetid.String())
	}
	if mainIndex >= fork.Lenth() {
		return nil, fmt.Errorf("MemDag.GetCurrentUnit error: forkindex is out of range")
	}
	forkdata := *(*fork)[mainIndex]

	if len(forkdata) > 0 {
		curHash := forkdata[len(forkdata)-1]
		curUnit, ok := (*chain.memUnit)[curHash]
		if !ok {
			return nil, fmt.Errorf("MemDag.GetCurrentUnit error: get no unit hash(%s) in memUnit", curHash.String())
		}
		return curUnit, nil
	}
	return nil, nil
}