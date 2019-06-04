package cors

import (
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptn/downloader"
)

func (pm *ProtocolManager) CorsHeaderMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	if pm.fetcher != nil {
		//TODO start lps broadcast
		for _, header := range headers {
			log.Trace("CorsHeaderMsg message content", "assetid:", header.Number.AssetID, "index:", header.Number.Index)
			pm.fetcher.Enqueue(p, header)
		}
	}
	return nil
}

func (pm *ProtocolManager) CorsHeadersMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Debug("CorsHeadersMsg message length", "len(headers)", len(headers))
	if pm.fetcher != nil {
		for _, header := range headers {
			//log.Trace("CorsHeadersMsg message content", "header:", header)
			pm.fetcher.Enqueue(p, header)
		}
		if len(headers) < MaxHeaderFetch {
			pm.bdlock.Lock() //TODO modify
			log.Info("CorsHeadersMsg message needboradcast", "assetid", headers[len(headers)-1].Number.AssetID,
				"index", headers[len(headers)-1].Number.Index)
			pm.needboradcast[p.id] = headers[len(headers)-1].Number.Index
			pm.bdlock.Unlock()
		}
	}
	return nil
}
func (pm *ProtocolManager) GetCurrentHeaderMsg(msg p2p.Msg, p *peer) error {
	var number modules.ChainIndex
	if err := msg.Decode(&number); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	header := pm.dag.CurrentHeader(number.AssetID)
	log.Trace("GetCurrentHeaderMsg message content", "number", number.AssetID, "header", header)
	var headers []*modules.Header
	headers = append(headers, header)
	return p.SendCurrentHeader(headers)
}

func (pm *ProtocolManager) CurrentHeaderMsg(msg p2p.Msg, p *peer) error {
	var headers []*modules.Header
	if err := msg.Decode(&headers); err != nil {
		log.Info("msg.Decode", "err:", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	log.Trace("CurrentHeaderMsg message content", "len(headers)", len(headers))
	if len(headers) != 1 {
		log.Info("CurrentHeaderMsg len err", "len(headers)", len(headers))
		return errResp(ErrDecode, "msg %v: %v", msg, "len is err")
	}
	if headers[0].Number.AssetID.String() != pm.assetId.String() {
		log.Info("CurrentHeaderMsg", "assetid not equal response", headers[0].Number.AssetID.String(), "local", pm.assetId.String())
		return errBadPeer
	}
	pm.headerCh <- &headerPack{p.id, headers}
	return nil
}

func (pm *ProtocolManager) GetBlockHeadersMsg(msg p2p.Msg, p *peer) error {
	// Decode the complex header query
	log.Debug("===Enter Light GetBlockHeadersMsg===")
	defer log.Debug("===End Ligth GetBlockHeadersMsg===")

	var query getBlockHeadersData
	if err := msg.Decode(&query); err != nil {
		log.Info("GetBlockHeadersMsg Decode", "err:", err, "msg:", msg)
		return errResp(ErrDecode, "%v: %v", msg, err)
	}

	log.Debug("ProtocolManager", "GetBlockHeadersMsg getBlockHeadersData:", query)

	hashMode := query.Origin.Hash != (common.Hash{})
	log.Debug("ProtocolManager", "GetBlockHeadersMsg hashMode:", hashMode)
	// Gather headers until the fetch or network limits is reached
	var (
		bytes   common.StorageSize
		headers []*modules.Header
		unknown bool
	)

	for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
		// Retrieve the next header satisfying the query
		var origin *modules.Header
		if hashMode {
			origin, _ = pm.dag.GetHeaderByHash(query.Origin.Hash)
		} else {
			log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Origin.Number:", query.Origin.Number.Index)
			origin, _ = pm.dag.GetHeaderByNumber(&query.Origin.Number)
		}

		if origin == nil {
			break
		}
		log.Debug("ProtocolManager", "GetBlockHeadersMsg origin index:", origin.Number.Index)

		number := origin.Number.Index
		headers = append(headers, origin)
		bytes += estHeaderRlpSize

		// Advance to the next header of the query
		switch {
		case hashMode && query.Reverse:
			// Hash based traversal towards the genesis block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the genesis block")
			for i := 0; i < int(query.Skip)+1; i++ {
				if header, err := pm.dag.GetHeaderByHash(query.Origin.Hash); err == nil && header != nil {
					if number != 0 {
						query.Origin.Hash = header.ParentsHash[0]
					}
					number--
				} else {
					//log.Info("========GetBlockHeadersMsg========", "number", number, "err:", err)
					unknown = true
					break
				}
			}
		case hashMode && !query.Reverse:
			// Hash based traversal towards the leaf block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Hash based towards the leaf block")
			var (
				current = origin.Number.Index
				next    = current + query.Skip + 1
				index   = origin.Number
			)
			log.Debug("ProtocolManager", "GetBlockHeadersMsg next", next, "current:", current)
			if next <= current {
				infos, _ := json.MarshalIndent(p.Peer.Info(), "", "  ")
				log.Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
				unknown = true
			} else {
				index.Index = next
				log.Debug("ProtocolManager", "GetBlockHeadersMsg index.Index:", index.Index)
				if header, _ := pm.dag.GetHeaderByNumber(index); header != nil {
					hashs := pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)
					log.Debug("ProtocolManager", "GetUnitHashesFromHash len(hashs):", len(hashs), "header.index:", header.Number.Index, "header.hash:", header.Hash().String(), "query.Skip+1", query.Skip+1)
					if len(hashs) > int(query.Skip) && (hashs[query.Skip] == query.Origin.Hash) {
						query.Origin.Hash = header.Hash()
					} else {
						log.Debug("ProtocolManager", "GetBlockHeadersMsg unknown = true; pm.dag.GetUnitHashesFromHash not equal origin hash.", "")
						log.Debug("ProtocolManager", "GetBlockHeadersMsg header.Hash()", header.Hash(), "query.Skip+1:", query.Skip+1, "query.Origin.Hash:", query.Origin.Hash)
						//log.Debug("ProtocolManager", "GetBlockHeadersMsg pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip]:", pm.dag.GetUnitHashesFromHash(header.Hash(), query.Skip+1)[query.Skip])
						unknown = true
					}
				} else {
					log.Debug("ProtocolManager", "GetBlockHeadersMsg unknown = true; pm.dag.GetHeaderByNumber not found. Index:", index.Index)
					unknown = true
				}
			}
		case query.Reverse:
			// Number based traversal towards the genesis block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the genesis block")
			if query.Origin.Number.Index >= query.Skip+1 {
				query.Origin.Number.Index -= query.Skip + 1
			} else {
				log.Info("ProtocolManager", "GetBlockHeadersMsg query.Reverse", "unknown is true")
				unknown = true
			}

		case !query.Reverse:
			// Number based traversal towards the leaf block
			log.Debug("ProtocolManager", "GetBlockHeadersMsg ", "Number based towards the leaf block")
			query.Origin.Number.Index += query.Skip + 1
		}
	}
	start := uint64(0)
	end := uint64(0)
	number := len(headers)
	if number > 0 {
		start = uint64(headers[0].Number.Index)
		end = uint64(headers[number-1].Number.Index)
	}
	log.Debug("ProtocolManager", "GetBlockHeadersMsg query.Amount", query.Amount, "send number:", len(headers), "start:", start, "end:", end, " getBlockHeadersData:", query)
	return p.SendUnitHeaders(headers)
}

func (pm *ProtocolManager) BlockHeadersMsg(msg p2p.Msg, p *peer) error {
	if pm.downloader == nil {
		return errResp(ErrUnexpectedResponse, "")
	}

	log.Trace("Received block header response message")
	// A batch of headers arrived to one of our previous requests
	var headers []*modules.Header

	if err := msg.Decode(&headers); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	err := pm.downloader.DeliverHeaders(p.id, headers)
	if err != nil {
		log.Debug(fmt.Sprint(err))
	}
	//p.fcServer.GotReply(resp.ReqID, resp.BV)
	//if pm.fetcher != nil && pm.fetcher.requestedID(resp.ReqID) {
	//	pm.fetcher.deliverHeaders(p, resp.ReqID, resp.Headers)
	//} else {
	//	err := pm.downloader.DeliverHeaders(p.id, resp.Headers)
	//	if err != nil {
	//		log.Debug(fmt.Sprint(err))
	//	}
	//}
	return nil
}