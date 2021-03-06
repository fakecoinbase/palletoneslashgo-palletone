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
 */

package ptn

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/modules"
)

// @author Albert·Gou
type producer interface {
	// SubscribeNewProducedUnitEvent should return an event subscription of
	// NewProducedUnitEvent and send events to the given channel.
	SubscribeNewProducedUnitEvent(ch chan<- mp.NewProducedUnitEvent) event.Subscription

	// AddToTBLSSignBufs is to TBLS sign the unit
	AddToTBLSSignBufs(newHash common.Hash)

	SubscribeSigShareEvent(ch chan<- mp.SigShareEvent) event.Subscription
	AddToTBLSRecoverBuf(sigShare *mp.SigShareEvent, header *modules.Header)

	SubscribeVSSDealEvent(ch chan<- mp.VSSDealEvent) event.Subscription
	AddToDealBuf(deal *mp.VSSDealEvent)

	SubscribeVSSResponseEvent(ch chan<- mp.VSSResponseEvent) event.Subscription
	AddToResponseBuf(resp *mp.VSSResponseEvent)

	LocalHaveActiveMediator() bool
	LocalHavePrecedingMediator() bool

	SubscribeGroupSigEvent(ch chan<- mp.GroupSigEvent) event.Subscription
	UpdateMediatorsDKG(isRenew bool)

	IsLocalMediator(add common.Address) bool
	ClearGroupSignBufs(stableUnit *modules.Unit)
}

func (pm *ProtocolManager) activeMediatorsUpdatedEventRecvLoop() {
	log.Debugf("activeMediatorsUpdatedEventRecvLoop")
	for {
		select {
		case event := <-pm.activeMediatorsUpdatedCh:
			go pm.switchMediatorConnect(event.IsChanged)

			// Err() channel will be closed when unsubscribing.
		case <-pm.activeMediatorsUpdatedSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) unstableRepositoryUpdatedRecvLoop() {
	log.Debugf("unstableRepositoryUpdatedRecvLoop")
	for {
		select {
		case <-pm.unstableRepositoryUpdatedCh:
			log.Debugf("receive UnstableRepositoryUpdatedEvent")
			pm.activeMediatorsUpdatedSub = pm.dag.SubscribeActiveMediatorsUpdatedEvent(pm.activeMediatorsUpdatedCh)

			// Err() channel will be closed when unsubscribing.
		case <-pm.unstableRepositoryUpdatedSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) saveStableUnitRecvLoop() {
	log.Debugf("saveStableUnitRecvLoop")
	for {
		select {
		case event := <-pm.saveStableUnitCh:
			log.Debugf("receive saveStableUnitEvent[%s]", event.Unit.DisplayId())
			txs := event.Unit.Transactions()
			if pm.enableGasFee {
				txs = event.Unit.TransactionsWithoutCoinbase()
			}
			if len(txs) > 0 {
				log.DebugDynamic(func() string {
					return fmt.Sprintf("discard txs %#x from txpool by stable unit[%s]",
						event.Unit.TxHashes(), event.Unit.DisplayId())
				})
				err := pm.txpool.DiscardTxs(txs)
				if err != nil {
					log.Error(err.Error())
				}
			}
			go pm.producer.ClearGroupSignBufs(event.Unit)
			go pm.delayDiscPrecedingMediator(event.Unit)

			// Err() channel will be closed when unsubscribing.
		case <-pm.saveStableUnitSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) saveUnitRecvLoop() {
	for {
		select {
		case u := <-pm.saveUnitCh:
			log.Debugf("SubscribeSaveUnitEvent received unit:%s", u.Unit.DisplayId())
			txs := u.Unit.Transactions()
			if pm.enableGasFee {
				txs = u.Unit.TransactionsWithoutCoinbase()
			}
			if len(txs) > 0 {
				err := pm.txpool.SetPendingTxs(u.Unit.Hash(), u.Unit.NumberU64(), txs) //UpdateTxStatusPacked
				if err != nil {
					log.Error(err.Error())
				}
			}

		case err := <-pm.saveUnitSub.Err():
			if err != nil {
				log.Error(err.Error())
			}
			return
		}

	}

}

func (pm *ProtocolManager) rollbackUnitRecvLoop() {

	for {
		select {
		case u := <-pm.rollbackUnitCh:
			log.Infof("SubscribeRollbackUnitEvent received unit:%s", u.Unit.DisplayId())
			txs := u.Unit.Transactions()
			if pm.enableGasFee {
				txs = u.Unit.TransactionsWithoutCoinbase()
			}
			if len(txs) > 0 {
				err := pm.txpool.ResetPendingTxs(txs) //UpdateTxStatusUnpacked
				if err != nil {
					log.Error(err.Error())
				}
			}

		case err := <-pm.rollbackUnitSub.Err():
			if err != nil {
				log.Error(err.Error())
			}
			return
		}

	}

}

func (pm *ProtocolManager) switchMediatorConnect(isChanged bool) {
	log.Debug("switchMediatorConnect", "isChanged", isChanged)

	dag := pm.dag
	// 防止重复接收消息
	if !(dag.LastMaintenanceTime() > pm.lastMaintenanceTime) {
		return
	}

	// 若干数据还没同步完成，则忽略本次切换，继续同步
	if !dag.IsSynced(true) {
		log.Debugf(errStr)
		return
	}

	headNum := dag.HeadUnitNum()
	stableNum := dag.StableUnitNum()
	isRenew := true
	// 如果 活跃mediator没有发生变化，并且群签名功能已生效，则不再重新做vss协议
	if !isChanged && !(headNum-stableNum > 1) && stableNum != 0 {
		isRenew = false
	}

	// 更新相关标记
	pm.lastMaintenanceTime = pm.dag.LastMaintenanceTime()

	if !isRenew {
		// 和新的活跃mediator节点相连
		go pm.connectWitchActiveMediators()
	}

	// 检查是否连接和同步，并更新DKG和VSS
	//go pm.checkConnectedAndSynced()
	go pm.producer.UpdateMediatorsDKG(isRenew)

	// 在其他地方当unit稳定后再关闭连接
	//// 延迟关闭和旧活跃mediator节点的连接
	//go pm.delayDiscPrecedingMediator()
}

func (pm *ProtocolManager) connectWitchActiveMediators() {
	// 判断本节点是否是活跃mediator
	log.Debugf("to connected with all active mediator nodes")
	if !pm.producer.LocalHaveActiveMediator() {
		return
	}

	// 和其他活跃mediator节点相连
	peers := pm.dag.GetActiveMediatorNodes()
	for _, peer := range peers {
		// 仅当不是本节点，才做处理
		if peer.ID != pm.srvr.Self().ID {
			pm.srvr.AddTrustedPeer(peer) // 加入Trusted列表
			pm.srvr.AddPeer(peer)        // 建立连接
		}
	}

	// 更新相关标记
	pm.isConnectedNewMediator = true
}

/*func (pm *ProtocolManager) checkConnectedAndSynced() {
	log.Debugf("check if it is connected to all active mediator peers")
	if !pm.producer.LocalHaveActiveMediator() {
		return
	}

	// 2. 是否和所有其他活跃mediator节点相连完成
	checkFn := func() bool {
		nodes := pm.dag.GetActiveMediatorNodes()
		for id, node := range nodes {
			// 仅当不是本节点，并还未连接完成时，或者未同步，返回false
			if node.ID == pm.srvr.Self().ID {
				continue
			}

			peer := pm.peers.Peer(id)
			if peer == nil {
				return false
			}
		}

		log.Debugf("connected with all active mediator peers")
		return true
	}

	// 3. 更新DKG和VSS
	processFn := func() {
		go pm.producer.UpdateMediatorsDKG(true)
	}

	// 1. 设置Ticker, 每隔一段时间检查一次
	checkTick := time.NewTicker(200 * time.Millisecond)

	defer checkTick.Stop()
	// 设置检查期限，防止死循环
	expiration := pm.dag.UnitIrreversibleTime()
	killLoop := time.NewTimer(expiration)

	for {
		select {
		case <-pm.quitSync:
			return
		case <-killLoop.C:
			return
		case <-checkTick.C:
			if checkFn() {
				processFn()
				return
			}
		}
	}
}*/

func (pm *ProtocolManager) delayDiscPrecedingMediator(stableUnit *modules.Unit) {
	if !pm.isConnectedNewMediator {
		return
	}

	if !(stableUnit.Timestamp() > pm.lastMaintenanceTime) {
		return
	}

	// 1. 判断当前节点是否是上一届活跃mediator
	if !pm.producer.LocalHavePrecedingMediator() {
		return
	}

	// 如果当前节点不是活跃mediator，则删除全部之前的mediator节点
	isActive := pm.producer.LocalHaveActiveMediator()

	// 2. 统计出需要断开连接的mediator节点
	delayDiscNodes := make(map[string]*discover.Node)

	activePeers := pm.dag.GetActiveMediatorNodes()
	precedingPeers := pm.dag.GetPrecedingMediatorNodes()
	for id, peer := range precedingPeers {
		// 仅当上一届mediator 不是本届活跃mediator，或者本节点不是活跃mediator
		if _, ok := activePeers[id]; !isActive || !ok /*&& pm.peers.Peer(id) != nil*/ {
			delayDiscNodes[id] = peer
		}
	}

	disconnectFn := func() {
		log.Debugf("disconnect with preceding mediator nodes")
		for _, peer := range delayDiscNodes {
			pm.srvr.RemoveTrustedPeer(peer)
		}
	}

	//// 3. 设置定时器延迟 将上一届的活跃mediator节点从Trusted列表中移除
	//expiration := pm.dag.UnitIrreversibleTime()
	//delayDisc := time.NewTimer(expiration)
	//
	//select {
	//case <-pm.quitSync:
	//	return
	//case <-delayDisc.C:
	//	disconnectFn()
	//}

	disconnectFn()
	pm.isConnectedNewMediator = false
}

func (p *peer) MarkVSSDeal(hash common.Hash) {
	for p.knownVSSDeal.Cardinality() >= maxKnownVSSDeal {
		p.knownVSSDeal.Pop()
	}
	p.knownVSSDeal.Add(hash)
}

func (pm *ProtocolManager) BroadcastVSSDeal(deal *mp.VSSDealEvent) {
	now := uint64(time.Now().Unix())
	if now > deal.Deadline {
		return
	}

	peers := pm.peers.PeersWithoutVSSDeal(deal.Hash())
	for _, peer := range peers {
		go peer.SendVSSDeal(deal)
	}

	return
}

func (ps *peerSet) PeersWithoutVSSDeal(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownVSSDeal.Contains(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (p *peer) MarkVSSResponse(hash common.Hash) {
	for p.knownVSSResponse.Cardinality() >= maxKnownVSSResponse {
		p.knownVSSResponse.Pop()
	}
	p.knownVSSResponse.Add(hash)
}

func (pm *ProtocolManager) BroadcastVSSResponse(deal *mp.VSSResponseEvent) {
	now := uint64(time.Now().Unix())
	if now > deal.Deadline {
		return
	}

	peers := pm.peers.PeersWithoutVSSResponse(deal.Hash())
	for _, peer := range peers {
		go peer.SendVSSResponse(deal)
	}

	return
}

func (ps *peerSet) PeersWithoutVSSResponse(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownVSSResponse.Contains(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (p *peer) MarkSigShare(hash common.Hash) {
	for p.knownSigShare.Cardinality() >= maxKnownSigShare {
		p.knownSigShare.Pop()
	}
	p.knownSigShare.Add(hash)
}

func (pm *ProtocolManager) BroadcastSigShare(sigShare *mp.SigShareEvent) {
	//now := uint64(time.Now().Unix())
	//if now > sigShare.Deadline {
	//	return
	//}

	peers := pm.peers.PeersWithoutSigShare(sigShare.UnitHash)
	for _, peer := range peers {
		go peer.SendSigShare(sigShare)
	}

	return
}

func (ps *peerSet) PeersWithoutSigShare(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownSigShare.Contains(hash) {
			list = append(list, p)
		}
	}
	return list
}
