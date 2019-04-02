package sysconfigcc

import (
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

const sysParam = "sysParam"
const sysParams = "sysParams"

//one topic
type SysVoteTopic struct {
	TopicTitle    string
	SelectOptions []string
	SelectMax     uint64
}

//topic support result
type SysTopicSupports struct {
	TopicTitle  string
	VoteResults []*modules.SysVoteResult
	SelectMax   uint64
	//SelectOptionsNum  uint64
}

//vote token information
type SysTokenInfo struct {
	Name        string
	Symbol      string
	CreateAddr  string
	VoteType    byte
	TotalSupply uint64
	VoteEndTime time.Time
	VoteContent []byte
	AssetID     modules.AssetId
}

//one user's support
type SysSupportRequest struct {
	TopicIndex   uint64
	SelectIndexs []uint64
}
