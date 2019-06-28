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

package deposit

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/contracts/syscontract"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

//  保存相关列表
func saveList(stub shim.ChaincodeStubInterface, key string, list map[string]bool) error {
	listByte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(key, listByte)
	if err != nil {
		return err
	}
	return nil
}

//  获取其他list
func getList(stub shim.ChaincodeStubInterface, typeList string) (map[string]bool, error) {
	byte, err := stub.GetState(typeList)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]bool)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//  判断 invokeTokens 是否包含保证金合约地址
func isContainDepositContractAddr(stub shim.ChaincodeStubInterface) (invokeToken *modules.InvokeTokens, err error) {
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return nil, err
	}
	for _, invokeTo := range invokeTokens {
		if strings.Compare(invokeTo.Address, syscontract.DepositContractAddress.String()) == 0 {
			return invokeTo, nil
		}
	}
	return nil, fmt.Errorf("it is not a depositContract invoke transaction")
}

//  处理部分保证金逻辑
func applyQuitList(role string, stub shim.ChaincodeStubInterface, args []string) error {
	//  获取请求调用地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return err
	}
	//  先获取申请列表
	listForQuit, err := GetListForQuit(stub)
	if err != nil {
		return err
	}
	// 判断列表是否为空
	if listForQuit == nil {
		listForQuit = make(map[string]*QuitNode)
	}
	quitNode := &QuitNode{
		Address: invokeAddr.String(),
		Role:    role,
		Time:    getTiem(stub),
	}

	//  保存退还列表
	listForQuit[invokeAddr.String()] = quitNode
	err = SaveListForQuit(stub, listForQuit)
	if err != nil {
		return err
	}
	return nil

}

//  加入相应候选列表，mediator jury dev
func addCandaditeList(stub shim.ChaincodeStubInterface, invokeAddr common.Address, candidate string) error {
	//  获取列表
	list, err := getList(stub, candidate)
	if err != nil {
		return err
	}
	if list == nil {
		list = make(map[string]bool)
	}
	if list[invokeAddr.String()] {
		return fmt.Errorf("node was in the list")
	}
	list[invokeAddr.String()] = true
	listByte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(candidate, listByte)
	if err != nil {
		return err
	}
	return nil
}

//  从候选列表删除mediator jury dev
func moveCandidate(candidate string, invokeFromAddr string, stub shim.ChaincodeStubInterface) error {
	//
	list, err := getList(stub, candidate)
	if err != nil {
		log.Error("stub.GetCandidateList err:", "error", err)
		return err
	}
	//
	if list == nil {
		log.Error("stub.GetCandidateList err: list is nil")
		return fmt.Errorf("%s", "list is nil")
	}
	if !list[invokeFromAddr] {
		return fmt.Errorf("node was not in the list")
	}
	delete(list, invokeFromAddr)
	//
	err = saveList(stub, candidate, list)
	if err != nil {
		return err
	}
	return nil

}

//  保存没收列表
func SaveListForForfeiture(stub shim.ChaincodeStubInterface, list map[string]*Forfeiture) error {
	byte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(ListForForfeiture, byte)
	if err != nil {
		return err
	}
	return nil
}

//  获取没收列表
func GetListForForfeiture(stub shim.ChaincodeStubInterface) (map[string]*Forfeiture, error) {
	byte, err := stub.GetState(ListForForfeiture)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]*Forfeiture)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//  保存退款列表
func SaveListForQuit(stub shim.ChaincodeStubInterface, list map[string]*QuitNode) error {
	byte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(ListForQuit, byte)
	if err != nil {
		return err
	}
	return nil
}

//  获取退出列表
func GetListForQuit(stub shim.ChaincodeStubInterface) (map[string]*QuitNode, error) {
	byte, err := stub.GetState(ListForQuit)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	list := make(map[string]*QuitNode)
	err = json.Unmarshal(byte, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func mediatorDepositKey(medAddr string) string {
	return string(constants.MEDIATOR_INFO_PREFIX) + string(constants.DEPOSIT_BALANCE_PREFIX) + medAddr
}

//  获取mediator
func GetMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string) (*MediatorDeposit, error) {
	byte, err := stub.GetState(mediatorDepositKey(medAddr))
	if err != nil || byte == nil {
		return nil, err
	}
	balance := NewMediatorDeposit()
	err = json.Unmarshal(byte, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

//  保存mediator
func SaveMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string, balance *MediatorDeposit) error {
	byte, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	err = stub.PutState(mediatorDepositKey(medAddr), byte)
	if err != nil {
		return err
	}

	return nil
}

//  删除mediator
func DelMediatorDeposit(stub shim.ChaincodeStubInterface, medAddr string) error {
	err := stub.DelState(mediatorDepositKey(medAddr))
	if err != nil {
		return err
	}

	return nil
}

//  保存jury/dev
func SaveNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string, balance *DepositBalance) error {
	balanceByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	err = stub.PutState(string(constants.DEPOSIT_BALANCE_PREFIX)+balanceAddr, balanceByte)
	if err != nil {
		return err
	}
	return nil
}

//  获取jury/dev
func GetNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string) (*DepositBalance, error) {
	byte, err := stub.GetState(string(constants.DEPOSIT_BALANCE_PREFIX) + balanceAddr)
	if err != nil {
		return nil, err
	}
	if byte == nil {
		return nil, nil
	}
	balance := &DepositBalance{}
	err = json.Unmarshal(byte, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

//  删除jury/dev
func DelNodeBalance(stub shim.ChaincodeStubInterface, balanceAddr string) error {
	err := stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + balanceAddr)
	if err != nil {
		return err
	}
	return nil
}

//  判断是否基金会发起的
func isFoundationInvoke(stub shim.ChaincodeStubInterface) bool {
	//  判断是否基金会发起的
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return false
	}
	//  获取
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return false
	}
	foundationAddress := cp.FoundationAddress
	// 判断当前请求的是否为基金会
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return false
	}
	return true
}

//  获取普通节点
//func getAccountMediatorVote(stub shim.ChaincodeStubInterface, invokeA string) ([]string, error) {
//	b, err := stub.GetState(string(constants.DEPOSIT_MEDIATOR_VOTE_PREFIX) + invokeA)
//	if err != nil {
//		return nil, err
//	}
//	if b == nil {
//		return nil, nil
//	}
//	mediators := []string{}
//	err = json.Unmarshal(b, &mediators)
//	if err != nil {
//		return nil, err
//	}
//	return mediators, nil
//}

//  保存普通节点
//func saveMediatorVote(stub shim.ChaincodeStubInterface, invokeA string, mediators []string) error {
//	config, _ := stub.GetSystemConfig()
//	maxVote := config.MaximumMediatorCount
//	if len(mediators) > int(maxVote) {
//		return errors.New(fmt.Sprintf("Too many mediators, must less or equal than %d", maxVote))
//	}
//	b, err := json.Marshal(mediators)
//	if err != nil {
//		return err
//	}
//	err = stub.PutState(string(constants.DEPOSIT_MEDIATOR_VOTE_PREFIX)+invokeA, b)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//  获取普通节点提取PTN
//func getExtPtn(stub shim.ChaincodeStubInterface) (map[string]*extractPtn, error) {
//	b, err := stub.GetState(ExtractPtnList)
//	if err != nil {
//		return nil, err
//	}
//	if b == nil {
//		return nil, nil
//	}
//	extP := make(map[string]*extractPtn)
//	err = json.Unmarshal(b, &extP)
//	if err != nil {
//		return nil, err
//	}
//	return extP, nil
//}

//  保存普通节点提取PTN
//func saveExtPtn(stub shim.ChaincodeStubInterface, extPtnL map[string]*extractPtn) error {
//	b, err := json.Marshal(extPtnL)
//	if err != nil {
//		return err
//	}
//	err = stub.PutState(ExtractPtnList, b)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//  获取当前PTN总量
//func getVotes(stub shim.ChaincodeStubInterface) (int64, error) {
//	b, err := stub.GetState(AllPledgeVotes)
//	if err != nil {
//		return 0, err
//	}
//	if b == nil {
//		return 0, nil
//	}
//	votes, err := strconv.ParseInt(string(b), 10, 64)
//	if err != nil {
//		return 0, err
//	}
//	return votes, nil
//}

//  保存PTN总量
//func saveVotes(stub shim.ChaincodeStubInterface, votes int64) error {
//	cur, err := getVotes(stub)
//	if err != nil {
//		return err
//	}
//	cur += votes
//	str := strconv.FormatInt(cur, 10)
//	err = stub.PutState(AllPledgeVotes, []byte(str))
//	if err != nil {
//		return err
//	}
//	return nil
//}

//  每天计算各节点收益
func handlePledgeReward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("need 0 args")
	}
	err := handleRewardAllocation(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)

}

//  每天计算各节点收益
//func handleEachDayAward1(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	if len(args) != 0 {
//		return shim.Error("need 0 args")
//	}
//	//  判断是否是基金会
//	if !isFoundationInvoke(stub) {
//		return shim.Error("please use foundation address")
//	}
//	//  判断当天是否处理过
//	if isHandled(stub) {
//		return shim.Error("had handled")
//	}
//	//  通过前缀获取所有mediator
//	mediators, err := stub.GetStateByPrefix(string(constants.MEDIATOR_INFO_PREFIX) + string(constants.DEPOSIT_BALANCE_PREFIX))
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//  通过前缀获取所有jury/dev
//	juryAndDevs, err := stub.GetStateByPrefix(string(constants.DEPOSIT_BALANCE_PREFIX))
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//  通过前缀获取所有普通节点
//	normalNodes, err := stub.GetStateByPrefix(string(constants.DEPOSIT_MEDIATOR_VOTE_PREFIX))
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//  获取集合全部
//	allM, err := getLastPledgeList(stub)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	//  第一次,收集当前的质押情况
//	if a == nil {
//		a = &PledgeList{}
//		a.Amount = pledgeVotes
//		//  计算mediators
//		for _, me := range mediators {
//			m := Members{}
//			m.Key = mediatorDepositKey(me.Key)
//			m.Value = me.Value
//			m.Role = Mediator
//			a.Members = append(a.Members, &m)
//		}
//		//  计算jury/dev
//		for _, jd := range juryAndDevs {
//			m := Members{}
//			m.Key = string(constants.DEPOSIT_BALANCE_PREFIX) + jd.Key
//			m.Value = jd.Value
//			m.Role = JuryAndDev
//			a.Members = append(a.Members, &m)
//		}
//		//  计算normalNode
//		for _, nor := range normalNodes {
//			m := Members{}
//			m.Key = string(constants.DEPOSIT_MEDIATOR_VOTE_PREFIX) + nor.Key
//			m.Value = nor.Value
//			m.Role = NormalNode
//			a.Members = append(a.Members, &m)
//		}
//		err = saveMember(stub, a)
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//		return shim.Success(nil)
//	} else {
//		//  获取每天总奖励
//		cp, err := stub.GetSystemConfig()
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//		depositExtraReward := cp.DepositExtraReward
//		rate := float64(depositExtraReward) / float64(a.Amount)
//		//  计算集合利息
//		for i, m := range a.Members {
//			fmt.Println(i)
//			switch m.Role {
//			case Mediator:
//				mediatorDeposit := MediatorDeposit{}
//				_ = json.Unmarshal(m.Value, &mediators)
//				dayAward := rate * float64(mediatorDeposit.DepositBalance.Balance)
//				mediatorDeposit.DepositBalance.Balance += uint64(dayAward)
//				//  获取该节点
//				//  添加利息
//				//  放到集合统一更新
//			case JuryAndDev:
//			case NormalNode:
//
//			}
//		}
//		//  先计算前一天的收益，再累加当前的质押情况
//
//	}
//
//	//  计算mediators
//	for _, m := range mediators {
//		mediatorDeposit := MediatorDeposit{}
//		_ = json.Unmarshal(m.Value, &mediators)
//		dayAward := rate * float64(mediatorDeposit.DepositBalance.Balance)
//		mediatorDeposit.DepositBalance.Balance += uint64(dayAward)
//		_ = SaveMediatorDeposit(stub, m.Key, &mediatorDeposit)
//	}
//	//  计算jury/dev
//	for _, jd := range juryAndDevs {
//		depositBalance := DepositBalance{}
//		_ = json.Unmarshal(jd.Value, &depositBalance)
//		dayAward := rate * float64(depositBalance.Balance)
//		depositBalance.Balance += uint64(dayAward)
//		_ = SaveNodeBalance(stub, jd.Key, &depositBalance)
//	}
//	//  计算normalNode
//	for _, nor := range normalNodes {
//		norNodBal := NorNodBal{}
//		_ = json.Unmarshal(nor.Value, &norNodBal)
//		dayAward := rate * float64(norNodBal.AmountAsset.Amount)
//		norNodBal.AmountAsset.Amount += uint64(dayAward)
//		_ = saveNor(stub, nor.Key, &norNodBal)
//	}
//	nDay := time.Now().UTC().Day()
//	err = stub.PutState(HandleEachDay, []byte(strconv.Itoa(nDay)))
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(nil)
//}

//func getAllMember1(stub shim.ChaincodeStubInterface) (map[string][]*Member, error) {
//	b, err := stub.GetState(MemberList)
//	if err != nil {
//		return nil, err
//	}
//	if b == nil {
//		return nil, nil
//	}
//	allM := make(map[string][]*Member)
//	err = json.Unmarshal(b, allM)
//	if err != nil {
//		return nil, err
//	}
//	return allM, nil
//}

//func saveLastPledgeList(stub shim.ChaincodeStubInterface, allM map[string][]*Members) error {
//	b, err := json.Marshal(allM)
//	if err != nil {
//		return err
//	}
//	err = stub.PutState(MemberList, b)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func getNormalNodeFromAllMember(stub shim.ChaincodeStubInterface) ([]*NorNodBal, error) {
	//mapAll, err := getLastPledgeList(stub)
	//if err != nil {
	//	return nil, err
	//}
	//if normalNodes, ok := mapAll[NormalNode]; ok {
	//	var norl []*NorNodBal
	//	for i, nor := range normalNodes {
	//		fmt.Println("===normalNode===>   ", i)
	//		norNodBal := NorNodBal{}
	//		_ = json.Unmarshal(nor.Value, &norNodBal)
	//		norl = append(norl, &norNodBal)
	//	}
	//	return norl, nil
	//}
	return nil, nil
}

func getJuryAndDevFromAllMember(stub shim.ChaincodeStubInterface) ([]*DepositBalance, error) {
	//mapAll, err := getLastPledgeList(stub)
	//if err != nil {
	//	return nil, err
	//}
	//if juryAndDevs, ok := mapAll[JuryAndDev]; ok {
	//	var jdl []*DepositBalance
	//	for i, jd := range juryAndDevs {
	//		fmt.Println("===jury and dev===>   ", i)
	//		depositBalance := DepositBalance{}
	//		_ = json.Unmarshal(jd.Value, &depositBalance)
	//		jdl = append(jdl, &depositBalance)
	//	}
	//	return jdl, nil
	//}
	return nil, nil

}

func getMediatorsFromeAllMember(stub shim.ChaincodeStubInterface) ([]*MediatorDeposit, error) {
	//mapAll, err := getLastPledgeList(stub)
	//if err != nil {
	//	return nil, err
	//}
	//if mediators, ok := mapAll[deposit.Mediator]; ok {
	//	var ml []*MediatorDeposit
	//	for i, m := range mediators {
	//		fmt.Println("===mediator===>   ", i)
	//		mediatorDeposit := MediatorDeposit{}
	//		_ = json.Unmarshal(m.Value, &mediatorDeposit)
	//		ml = append(ml, &mediatorDeposit)
	//	}
	//	return ml, nil
	//}
	return nil, nil
}

func getTiem(stub shim.ChaincodeStubInterface) string {
	t, _ := stub.GetTxTimestamp(10)
	ti := time.Unix(t.Seconds, 0)
	return ti.Format(Layout2)
}

func getToday(stub shim.ChaincodeStubInterface) string {
	t, _ := stub.GetTxTimestamp(10)

	ti := time.Unix(t.Seconds, 0)
	str := ti.Format("20060102")
	log.Debugf("getToday GetTxTimestamp 10 result:%d, format string:%s", t.Seconds, str)
	return str
}

//质押分红处理
func handleRewardAllocation(stub shim.ChaincodeStubInterface) error {
	//  判断当天是否处理过
	today := getToday(stub)
	lastDate, err := getLastPledgeListDate(stub)

	if lastDate == today {
		return fmt.Errorf("%s pledge reward has been allocated before", today)
	}
	allM, err := getLastPledgeList(stub)
	if err != nil {
		return err
	}
	//计算分红
	if allM != nil {
		cp, err := stub.GetSystemConfig()
		if err != nil {
			return err
		}
		depositDailyReward := cp.DepositDailyReward
		allM = pledgeRewardAllocation(allM, depositDailyReward)
	} else {
		allM = &modules.PledgeList{}
	}
	allM.Date = today
	// 增加新的质押
	depositList, err := getAllPledgeDepositRecords(stub)
	if err != nil {
		return err
	}

	for _, awardNode := range depositList {
		allM.Add(awardNode.Address, awardNode.Amount)
		err = delPledgeDepositRecord(stub, awardNode.Address)
		if err != nil {
			return err
		}
	}
	err = saveLastPledgeList(stub, allM)
	if err != nil {
		return err
	}

	//处理提币请求
	withdrawList, err := getAllPledgeWithdrawRecords(stub)
	if err != nil {
		return err
	}
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	for _, withdraw := range withdrawList {
		withdrawAmt, _ := allM.Reduce(withdraw.Address, withdraw.Amount)
		if withdrawAmt > 0 {
			err := stub.PayOutToken(withdraw.Address, modules.NewAmountAsset(withdraw.Amount, gasToken), 0)
			if err != nil {
				return err
			}
			err = delPledgeWithdrawRecord(stub, withdraw.Address) //清空提取请求列表
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//  社区申请没收某节点的保证金数量
func applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("applyForForfeitureDeposit")
	if len(args) != 3 {
		log.Error("args need four parameters")
		return shim.Error("args need four parameters")
	}
	//  需要判断是否基金会发起的
	//if !isFoundationInvoke(stub) {
	//	log.Error("please use foundation address")
	//	return shim.Error("please use foundation address")
	//}
	//  被没收地址
	forfeitureAddr := args[0]
	//  判断没收地址是否正确
	f, err := common.StringToAddress(forfeitureAddr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//  需要判断是否已经被没收过了
	listForForfeiture, err := GetListForForfeiture(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	//
	if listForForfeiture == nil {
		listForForfeiture = make(map[string]*Forfeiture)
	} else {
		//
		if _, ok := listForForfeiture[f.String()]; ok {
			return shim.Error("node was in the forfeiture list")
		}
	}
	//  被没收地址属于哪种类型
	role := args[1]
	//  没收理由
	extra := args[2]

	//  申请地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//  存储信息
	forfeiture := &Forfeiture{}
	forfeiture.ApplyAddress = invokeAddr.String()
	forfeiture.ForfeitureAddress = forfeitureAddr
	forfeiture.ForfeitureRole = role
	forfeiture.Extra = extra
	forfeiture.ApplyTime = getTiem(stub)
	listForForfeiture[f.String()] = forfeiture
	//  保存列表
	err = SaveListForForfeiture(stub, listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}
