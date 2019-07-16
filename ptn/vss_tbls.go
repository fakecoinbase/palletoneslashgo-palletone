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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func (self *ProtocolManager) newProducedUnitBroadcastLoop() {
	for {
		select {
		case event := <-self.newProducedUnitCh:
			self.BroadcastUnit(event.Unit, true)
			//self.BroadcastCorsHeader(event.Unit.Header(), self.SubProtocols[0].Name)

		case <-self.newProducedUnitSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) toGroupSignEventRecvLoop() {
	log.Debugf("toGroupSignEventRecvLoop")
	for {
		select {
		case event := <-pm.toGroupSignCh:
			go pm.toGroupSign(event)

		// Err() channel will be closed when unsubscribing.
		case <-pm.toGroupSignSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) toGroupSign(event modules.ToGroupSignEvent) {
	log.Debugf("receive toGroupSign event")

	// 判断是否满足群签名的条件
	if !pm.dag.IsSynced() {
		log.Debugf("dag is not synced")
		return
	}

	if !pm.producer.LocalHaveActiveMediator() && !pm.producer.LocalHavePrecedingMediator() {
		log.Debugf("the current node has no mediator")
		return
	}

	//if !pm.producer.IsEnabledGroupSign() {
	//	log.Debugf("the current node is enabled groupSign")
	//	return
	//}

	// 获取最高稳定单元的高度
	gasToken := dagconfig.DagConfig.GetGasToken()
	iun := pm.dag.GetIrreversibleUnitNum(gasToken)

	// 对稳定单元后一个unit进行群签名
	newUnit, err := pm.dag.GetUnitByNumber(&modules.ChainIndex{gasToken, iun + 1})
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	go pm.producer.AddToTBLSSignBufs(newUnit)
}

// @author Albert·Gou
func (self *ProtocolManager) sigShareTransmitLoop() {
	for {
		select {
		case event := <-self.sigShareCh:
			unit, err := self.dag.GetUnitByHash(event.UnitHash)
			if unit != nil && err == nil {
				med := unit.Author()
				node := self.dag.GetActiveMediator(med).Node
				self.TransmitSigShare(node, &event)
			}

			// Err() channel will be closed when unsubscribing.
		case <-self.sigShareSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitSigShare(node *discover.Node, sigShare *mp.SigShareEvent) {
	peer, self := pm.GetPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(sigShare)
		//if err != nil {
		//	log.Debug(err.Error())
		//}
		//
		//var s mp.SigShareEvent
		//stream := rlp.NewStream(reader, uint64(size))
		//if err := stream.Decode(&s); err != nil {
		//	log.Debug(err.Error())
		//}
		//pm.producer.AddToTBLSRecoverBuf(s.UnitHash, s.SigShare)

		pm.producer.AddToTBLSRecoverBuf(sigShare.UnitHash, sigShare.SigShare)
		return
	}

	if peer == nil {
		return
	}

	err := peer.SendSigShare(sigShare)
	if err != nil {
		log.Debug(err.Error())
	}
}

// @author Albert·Gou
func (self *ProtocolManager) groupSigBroadcastLoop() {
	for {
		select {
		case event := <-self.groupSigCh:
			self.BroadcastGroupSig(&event)

		// Err() channel will be closed when unsubscribing.
		case <-self.groupSigSub.Err():
			return
		}
	}
}

// @author Albert·Gou
// BroadcastGroupSig will propagate the group signature of unit to p2p network
func (pm *ProtocolManager) BroadcastGroupSig(groupSig *mp.GroupSigEvent) {
	peers := pm.peers.PeersWithoutGroupSig(groupSig.UnitHash)
	for _, peer := range peers {
		peer.SendGroupSig(groupSig)
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssDealTransmitLoop() {
	for {
		select {
		case event := <-self.vssDealCh:
			node := self.dag.GetActiveMediatorNode(int(event.DstIndex))
			self.TransmitVSSDeal(node, &event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssDealSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) TransmitVSSDeal(node *discover.Node, deal *mp.VSSDealEvent) {
	peer, self := pm.GetPeer(node)
	if self {
		//size, reader, err := rlp.EncodeToReader(deal)
		//if err != nil {
		//	log.Debug(err.Error())
		//}
		//
		//var d mp.VSSDealEvent
		//s := rlp.NewStream(reader, uint64(size))
		//if err := s.Decode(&d); err != nil {
		//	log.Debug(err.Error())
		//}
		//pm.producer.ProcessVSSDeal(&d)

		pm.producer.ProcessVSSDeal(deal)
		return
	}

	if peer == nil {
		return
	}

	// comment by Albert·Gou
	// // append by wangjiyou
	//if pm.peers.PeersWithoutVss(dstId) {
	//	return
	//}
	//pm.peers.MarkVss(dstId)

	//msg := &vssMsg{
	//	NodeId: dstId,
	//	Deal:   deal,
	//}
	//err := peer.SendVSSDeal(msg)

	err := peer.SendVSSDeal(deal)
	if err != nil {
		log.Debug(err.Error())
	}
}

// @author Albert·Gou
func (self *ProtocolManager) vssResponseBroadcastLoop() {
	for {
		select {
		case event := <-self.vssResponseCh:
			self.BroadcastVssResp(&event)

			// Err() channel will be closed when unsubscribing.
		case <-self.vssResponseSub.Err():
			return
		}
	}
}

// @author Albert·Gou
//func (pm *ProtocolManager) BroadcastVssResp(dstId string, resp *mp.VSSResponseEvent) {
func (pm *ProtocolManager) BroadcastVssResp(resp *mp.VSSResponseEvent) {
	// comment by Albert·Gou
	//dstId := node.ID.TerminalString()
	//peer := pm.peers.Peer(dstId)
	//if peer == nil {
	//	log.Debugf("peer not exist: %v", node.String())
	//}

	// comment by Albert·Gou
	//if pm.peers.PeersWithoutVssResp(dstId) {
	//	return
	//}
	//pm.peers.MarkVssResp(dstId)

	peers := pm.GetActiveMediatorPeers()
	//peers := pm.GetTransitionPeers()
	for _, peer := range peers {
		if peer == nil { // 此时为本节点
			//size, reader, err := rlp.EncodeToReader(resp)
			//if err != nil {
			//	log.Debug(err.Error())
			//}
			//
			//var r mp.VSSResponseEvent
			//s := rlp.NewStream(reader, uint64(size))
			//if err := s.Decode(&r); err != nil {
			//	log.Debug(err.Error())
			//}
			//go pm.producer.AddToResponseBuf(&r)

			go pm.producer.AddToResponseBuf(resp)
			continue
		}

		// comment by Albert·Gou
		//dstId := peer.id
		//if pm.peers.PeersWithoutVssResp(dstId) {
		//	return
		//}
		//pm.peers.MarkVssResp(dstId)

		// comment by Albert·Gou
		//msg := &vssRespMsg{
		//	NodeId: dstId,
		//	Resp:   resp,
		//}
		//
		//err := peer.SendVSSResponse(msg)

		err := peer.SendVSSResponse(resp)
		if err != nil {
			log.Info(err.Error())
		}
	}
}

// GetPeer, retrieve specified peer. If it is the node itself, p is nil and self is true
// @author Albert·Gou
func (pm *ProtocolManager) GetPeer(node *discover.Node) (p *peer, self bool) {
	id := node.ID
	if pm.srvr.Self().ID == id {
		self = true
	}

	p = pm.peers.Peer(id.TerminalString())
	if p == nil && !self {
		log.Debugf("the Peer is not exist: %v", node.String())
	}

	return
}

// GetActiveMediatorPeers retrieves a list of peers that active mediator.
// If the value is nil, it is the node itself
// @author Albert·Gou
func (pm *ProtocolManager) GetActiveMediatorPeers() map[string]*peer {
	nodes := pm.dag.GetActiveMediatorNodes()
	list := make(map[string]*peer, len(nodes))

	for id, node := range nodes {
		peer, self := pm.GetPeer(node)
		if peer != nil || self {
			list[id] = peer
		}
	}

	return list
}

// @author Albert·Gou
//func (p *peer) SendVSSDeal(deal *vssMsg) error {
func (p *peer) SendVSSDeal(deal *mp.VSSDealEvent) error {
	return p2p.Send(p.rw, VSSDealMsg, deal)
}

// @author Albert·Gou
//func (p *peer) SendVSSResponse(resp *vssRespMsg) error {
func (p *peer) SendVSSResponse(resp *mp.VSSResponseEvent) error {
	return p2p.Send(p.rw, VSSResponseMsg, resp)
}

// @author Albert·Gou
func (p *peer) SendSigShare(sigShare *mp.SigShareEvent) error {
	return p2p.Send(p.rw, SigShareMsg, sigShare)
}

//BroadcastGroupSig
func (p *peer) SendGroupSig(groupSig *mp.GroupSigEvent) error {
	p.knownGroupSig.Add(groupSig.UnitHash)
	return p2p.Send(p.rw, GroupSigMsg, groupSig)
}
