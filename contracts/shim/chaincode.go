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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

// Package shim provides APIs for the chaincode to access its state
// variables, transaction context and call other chaincodes.
package shim

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/comm"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// Logger for the shim package.
//var log = log.New("shim")
//var logOutput = os.Stderr

var key string
var cert string
var GlobalStateContractId = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var ERROR_ONLY_SYS_CONTRACT = errors.New("Only system contract can call this function.")

//const (
//
//	minUnicodeRuneValue   = 0            //U+0000
//	maxUnicodeRuneValue   = utf8.MaxRune //U+10FFFF - maximum (and unallocated) code point
//	compositeKeyNamespace = "\x00"
//	emptyKeySubstitute    = "\x01"
//)

// ChaincodeStub is an object passed to chaincode for shim side handling of
// APIs.
type ChaincodeStub struct {
	ContractId     []byte
	TxID           string
	ChannelId      string
	chaincodeEvent *pb.PtnChaincodeEvent
	args           [][]byte
	handler        *Handler
	signedProposal *pb.PtnSignedProposal
	//proposal       *pb.PtnProposal

	// Additional fields extracted from the signedProposal
	//creator   []byte
	//transient map[string][]byte
	//binding   []byte

	decorations map[string][]byte
}

// Peer address derived from command line or env var
var peerAddress string

//this separates the chaincode stream interface establishment
//so we can replace it with a mock peer stream
type peerStreamGetter func(name string) (PeerChaincodeStream, error)

//UTs to setup mock peer stream getter
var streamGetter peerStreamGetter

//the non-mock user CC stream establishment func
func userChaincodeStreamGetter(name string) (PeerChaincodeStream, error) {
	flag.StringVar(&peerAddress, "peer.address", "", "peer address")
	if comm.TLSEnabled() {
		keyPath := viper.GetString("tls.client.key.path")
		certPath := viper.GetString("tls.client.cert.path")
		data, err1 := ioutil.ReadFile(keyPath)
		if err1 != nil {
			err1 = errors.Wrap(err1, fmt.Sprintf("error trying to read file content %s", keyPath))
			log.Errorf("%+v", err1)
			return nil, err1
		}
		key = string(data)
		data, err1 = ioutil.ReadFile(certPath)
		if err1 != nil {
			err1 = errors.Wrap(err1, fmt.Sprintf("error trying to read file content %s", certPath))
			log.Errorf("%+v", err1)
			return nil, err1
		}
		cert = string(data)
	}
	flag.Parse()
	//TODO peer
	log.Debugf("Peer address: %s", viper.GetString("chaincode.peer.address"))
	// Establish connection with validating peer
	clientConn, err := newPeerClientConnection()
	if err != nil {
		err = errors.Wrap(err, "error trying to connect to local peer")
		log.Errorf("%+v", err)
		return nil, err
	}
	log.Debugf("os.Args returns: %s", os.Args)
	chaincodeSupportClient := pb.NewChaincodeSupportClient(clientConn)
	// Establish stream with validating peer
	stream, err := chaincodeSupportClient.Register(context.Background())
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error chatting with leader at address=%s",
			getPeerAddress()))
	}
	return stream, nil
}

// chaincodes.
func Start(cc Chaincode) error {
	// If Start() is called, we assume this is a standalone chaincode and set
	// up formatted logging.
	//SetupChaincodeLogging()
	viper.SetEnvPrefix("CORE")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	chaincodename := viper.GetString("chaincode.id.name")
	if chaincodename == "" {
		return errors.New("error chaincode id not provided")
	}
	//mock stream not set up ... get real stream
	if streamGetter == nil {
		streamGetter = userChaincodeStreamGetter
	}
	stream, err := streamGetter(chaincodename)
	if err != nil {
		return err
	}
	err = chatWithPeer(chaincodename, stream, cc)
	return err
}

// IsEnabledForLogLevel checks to see if the log is enabled for a specific logging level
// used primarily for testing
//func IsEnabledForLogLevel(logLevel string) bool {
//	lvl, _ := logging.LogLevel(logLevel)
//	return log.IsEnabledFor(lvl)
//}

// SetupChaincodeLogging sets the chaincode logging format and the level
// to the values of CORE_CHAINCODE_LOGFORMAT and CORE_CHAINCODE_LOGLEVEL set
// from core.yaml by chaincode_support.go
//func SetupChaincodeLogging() {
//	viper.SetEnvPrefix("CORE")
//	viper.AutomaticEnv()
//	replacer := strings.NewReplacer(".", "_")
//	viper.SetEnvKeyReplacer(replacer)
//	// setup system-wide logging backend
//	logFormat := flogging.SetFormat(viper.GetString("chaincode.logging.format"))
//	flogging.InitBackend(logFormat, logOutput)
//	// set default log level for all modules
//	chaincodeLogLevelString := viper.GetString("chaincode.logging.level")
//	if chaincodeLogLevelString == "" {
//		log.Infof("Chaincode log level not provided; defaulting to: %s", flogging.DefaultLevel())
//		flogging.InitFromSpec(flogging.DefaultLevel())
//	} else {
//		_, err := LogLevel(chaincodeLogLevelString)
//		if err == nil {
//			flogging.InitFromSpec(chaincodeLogLevelString)
//		} else {
//			log.Warningf("Error: '%s' for chaincode log level: %s; defaulting to %s", err, chaincodeLogLevelString,
// flogging.DefaultLevel())
//			flogging.InitFromSpec(flogging.DefaultLevel())
//		}
//	}
//	// override the log level for the shim logging module - note: if this value is
//	// blank or an invalid log level, then the above call to
//	// `flogging.InitFromSpec` already set the default log level so no action
//	// is required here.
//	shimLogLevelString := viper.GetString("chaincode.logging.shim")
//	if shimLogLevelString != "" {
//		shimLogLevel, err := LogLevel(shimLogLevelString)
//		if err == nil {
//			SetLoggingLevel(shimLogLevel)
//		} else {
//			log.Warningf("Error: %s for shim log level: %s", err, shimLogLevelString)
//		}
//	}
//	//now that logging is setup, print build level. This will help making sure
//	//chaincode is matched with peer.
//	buildLevel := viper.GetString("chaincode.buildlevel")
//	log.Infof("Chaincode (build level: %s) starting up ...", buildLevel)
//}

// StartInProc is an entry point for system chaincodes bootstrap. It is not an
// API for chaincodes.
func StartInProc(env []string, args []string, cc Chaincode, recv <-chan *pb.PtnChaincodeMessage,
	send chan<- *pb.PtnChaincodeMessage) error {
	log.Debugf("in proc %v", args)
	var chaincodename string
	for _, v := range env {
		if strings.Index(v, "CORE_CHAINCODE_ID_NAME=") == 0 {
			p := strings.SplitAfter(v, "CORE_CHAINCODE_ID_NAME=")
			chaincodename = p[1]
			break
		}
	}
	if chaincodename == "" {
		return errors.New("error chaincode id not provided")
	}
	stream := newInProcStream(recv, send)
	log.Debugf("starting chat with peer using name=%s", chaincodename)
	err := chatWithPeer(chaincodename, stream, cc)
	return err
}

func getPeerAddress() string {
	if peerAddress != "" {
		return peerAddress
	}
	////if peerAddress = viper.GetString("peer.address"); peerAddress == "" {
	//if peerAddress = cfg.GetConfig().ContractAddress; peerAddress == "" {
	//	log.Error("peer.address not configured, can't connect to peer")
	//}
	peerAddress = viper.GetString("chaincode.peer.address")
	return peerAddress
}

func newPeerClientConnection() (*grpc.ClientConn, error) {
	var peerAddress = getPeerAddress()
	// set the keepalive options to match static settings for chaincode server
	kaOpts := &comm.KeepaliveOptions{
		ClientInterval: time.Duration(1) * time.Minute,
		ClientTimeout:  time.Duration(20) * time.Second,
	}
	if comm.TLSEnabled() {
		return comm.NewClientConnectionWithAddress(peerAddress, true, true,
			comm.InitTLSForShim(key, cert), kaOpts)
	}
	//TODO peer
	log.Debugf("PeerClient: %s", getPeerAddress())
	return comm.NewClientConnectionWithAddress(peerAddress, true, false, nil, kaOpts)
}

func chatWithPeer(chaincodename string, stream PeerChaincodeStream, cc Chaincode) error {
	// Create the shim handler responsible for all control logic
	handler := newChaincodeHandler(stream, cc)
	defer stream.CloseSend()
	// Send the ChaincodeID during register.
	chaincodeID := &pb.PtnChaincodeID{Name: chaincodename}
	payload, err := proto.Marshal(chaincodeID)
	if err != nil {
		return errors.Wrap(err, "error marshaling chaincodeID during chaincode registration")
	}
	// Register on the stream
	log.Debugf("Registering.. sending %s", pb.PtnChaincodeMessage_REGISTER)
	err = handler.serialSend(&pb.PtnChaincodeMessage{Type: pb.PtnChaincodeMessage_REGISTER, Payload: payload})
	if err != nil {
		return errors.WithMessage(err, "error sending chaincode REGISTER")
	}
	waitc := make(chan struct{})
	errc := make(chan error)
	go func() {
		defer close(waitc)
		msgAvail := make(chan *pb.PtnChaincodeMessage)
		var nsInfo *nextStateInfo
		var in *pb.PtnChaincodeMessage
		recv := true
		for {
			in = nil
			err = nil
			nsInfo = nil
			if recv {
				recv = false
				go func() {
					var in2 *pb.PtnChaincodeMessage
					in2, err = stream.Recv()
					msgAvail <- in2
				}()
			}
			select {
			case sendErr := <-errc:
				//serialSendAsync successful?
				if sendErr == nil {
					continue
				}
				//no, bail
				err = errors.Wrap(sendErr, fmt.Sprintf("error sending type nil"))
				return
			case in = <-msgAvail:
				if err == io.EOF {
					err = errors.Wrapf(err, "received EOF, ending chaincode stream")
					log.Debugf("%+v", err)
					return
				} else if err != nil {
					log.Errorf("Received error from server, ending chaincode stream: %+v", err)
					return
				} else if in == nil {
					//err = errors.New("received nil message, ending chaincode stream")
					//log.Debugf("%+v", err)
					return
				}
				log.Debugf("[%s]Received message %s from shim", shorttxid(in.Txid), in.Type.String())
				recv = true
			case nsInfo = <-handler.nextState:
				in = nsInfo.msg
				if in == nil {
					log.Debugf("nil msg")
					panic("nil msg")
				}
				log.Debugf("[%s]Move state message %s", shorttxid(in.Txid), in.Type.String())
			}
			// Call FSM.handleMessage()
			err = handler.handleMessage(in)
			if err != nil {
				err = errors.WithMessage(err, "error handling message")
				return
			}
			//keepalive messages are PONGs to the PINGs
			if in.Type == pb.PtnChaincodeMessage_KEEPALIVE {
				log.Debug("Sending KEEPALIVE response")
				//ignore any errors, maybe next KEEPALIVE will work
				handler.serialSendAsync(in, nil)
			} else if nsInfo != nil && nsInfo.sendToCC {
				log.Debugf("[%s]send state message %s", shorttxid(in.Txid), in.Type.String())
				handler.serialSendAsync(in, errc)
			}
		}
	}()
	<-waitc
	return err
}

// -- init stub ---
// ChaincodeInvocation functionality
func (stub *ChaincodeStub) init(handler *Handler, contractid []byte, channelId string, txid string,
	input *pb.PtnChaincodeInput, signedProposal *pb.PtnSignedProposal) error {
	stub.TxID = txid
	stub.ChannelId = channelId
	stub.args = input.Args
	stub.handler = handler
	stub.signedProposal = signedProposal
	stub.decorations = input.Decorations
	stub.ContractId = contractid
	//log.Info("args:", input.Args)
	for _, tp := range input.Args {
		log.Debugf("%s", tp)
	}
	// TODO: sanity check: verify that every call to init with a nil
	// signedProposal is a legitimate one, meaning it is an internal call
	// to system chaincodes.
	return nil
}

// GetTxID returns the transaction ID for the proposal
func (stub *ChaincodeStub) GetTxID() string {
	return stub.TxID
}

// GetChannelID returns the channel for the proposal
func (stub *ChaincodeStub) GetChannelID() string {
	return stub.ChannelId
}

// ------------- Call Chaincode functions ---------------

// InvokeChaincode documentation can be found in interfaces.go
func (stub *ChaincodeStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	// Internally we handle chaincode name as a composite name
	if channel != "" {
		chaincodeName = chaincodeName + "/" + channel
	}
	return stub.handler.handleInvokeChaincode(chaincodeName, args, stub.ChannelId, stub.TxID)
}

// --------- State functions ----------

// GetState documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetState(key string) ([]byte, error) {
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handleGetState(collection, key, stub.ContractId, stub.ChannelId, stub.TxID)
}
func (stub *ChaincodeStub) GetGlobalState(key string) ([]byte, error) {
	return stub.handler.handleGetState("", key, GlobalStateContractId, stub.ChannelId, stub.TxID)

}
func (stub *ChaincodeStub) GetContractState(contractAddr common.Address, key string) ([]byte, error) {
	contractId := contractAddr.Bytes()
	return stub.handler.handleGetState("", key, contractId, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) GetContractStateByPrefix(contractAddr common.Address, prefix string) ([]*modules.KeyValue, error) {
	contractId := contractAddr.Bytes()
	return stub.handler.handelGetStateByPrefix(prefix, contractId, stub.ChannelId, stub.TxID)

}

func (stub *ChaincodeStub) GetStateByPrefix(prefix string) ([]*modules.KeyValue, error) {
	return stub.handler.handelGetStateByPrefix(prefix, stub.ContractId, stub.ChannelId, stub.TxID)
}

// PutState documentation can be found in interfaces.go
func (stub *ChaincodeStub) PutState(key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be an empty string")
	}
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handlePutState(collection, nil, key, value, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) PutGlobalState(key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be an empty string")
	}
	if !common.IsSystemContractId(stub.ContractId) {
		return ERROR_ONLY_SYS_CONTRACT
	}
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handlePutState(collection, GlobalStateContractId, key, value, stub.ChannelId, stub.TxID)

}

func (stub *ChaincodeStub) PutContractState(contractAddr common.Address, key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be an empty string")
	}
	if !common.IsSystemContractId(stub.ContractId) {
		return ERROR_ONLY_SYS_CONTRACT
	}

	// Access public data by setting the collection to empty string
	collection := ""
	contractId := contractAddr.Bytes()
	return stub.handler.handlePutState(collection, contractId, key, value, stub.ChannelId, stub.TxID)
}

// DelState documentation can be found in interfaces.go
func (stub *ChaincodeStub) DelState(key string) error {
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handleDelState(collection, nil, key, stub.ChannelId, stub.TxID)
}
func (stub *ChaincodeStub) DelGlobalState(key string) error {
	if !common.IsSystemContractId(stub.ContractId) {
		return ERROR_ONLY_SYS_CONTRACT
	}
	return stub.handler.handleDelState("", GlobalStateContractId, key, stub.ChannelId, stub.TxID)

}

func (stub *ChaincodeStub) OutChainCall(outChainName string, method string, params []byte) ([]byte, error) {
	if outChainName == "" {
		return nil, errors.New("outChainName must not be an empty string")
	}
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handleOutCall(collection, outChainName, method, params, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) SendJury(msgType uint32, consultContent []byte, myAnswer []byte) ([]byte, error) {
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handleSendJury(collection, msgType, consultContent, myAnswer, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) RecvJury(msgType uint32, consultContent []byte, timeout uint32) ([]byte, error) {
	// Access public data by setting the collection to empty string
	collection := ""
	return stub.handler.handleRecvJury(collection, msgType, consultContent, timeout, stub.ChannelId, stub.TxID)
}

// GetArgs documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetArgs() [][]byte {
	if len(stub.args) <= 2 {
		return nil
	}
	return stub.args[2:]
}

// GetStringArgs documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetStringArgs() []string {
	args := stub.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

// GetFunctionAndParameters documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetFunctionAndParameters() (function string, params []string) {
	allargs := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}

//GetInvokeParameters documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetInvokeParameters() (invokeAddr common.Address, invokeTokens []*modules.InvokeTokens,
	invokeFees *modules.AmountAsset, funcName string, params []string, err error) {
	allargs := stub.args
	if len(allargs) <= 0 {
		return
	}
	invokeInfo := &modules.InvokeInfo{}
	err = json.Unmarshal(allargs[0], invokeInfo)
	if err != nil {
		return common.Address{}, nil, nil, "", nil, err
	}
	invokeAddr = invokeInfo.InvokeAddress
	invokeTokens = invokeInfo.InvokeTokens
	invokeFees = invokeInfo.InvokeFees
	if len(allargs) < 2 {
		return
	}
	strargs := make([]string, 0, len(allargs)-1)
	for _, barg := range allargs[1:] {
		strargs = append(strargs, string(barg))
	}
	if len(strargs) < 2 {
		return
	}
	funcName = strargs[1]
	if len(strargs) < 3 {
		return
	}
	params = strargs[2:]
	return
}

// GetArgsSlice documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetArgsSlice() ([]byte, error) {
	args := stub.GetArgs()
	res := []byte{}
	for _, barg := range args {
		res = append(res, barg...)
	}
	return res, nil
}

// GetTxTimestamp documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetTxTimestamp(rangeNumber uint32) (*timestamp.Timestamp, error) {
	//glh
	/*
		hdr, err := utils.GetHeader(stub.proposal.Header)
		if err != nil {
			return nil, err
		}
		chdr, err := utils.UnmarshalChannelHeader(hdr.ChannelHeader)
		if err != nil {
			return nil, err
		}

		return chdr.GetTimestamp(), nil
	*/
	headerTime, err := stub.handler.handleGetTimestamp("", rangeNumber, stub.ContractId, stub.ChannelId,
		stub.TxID)
	if err != nil {
		return nil, errors.New("handleGetState failed")
	}
	secs, _ := strconv.ParseInt(string(headerTime), 10, 64)
	timeStamp := &(timestamp.Timestamp{Seconds: secs, Nanos: 0})
	return timeStamp, nil
}

// ------------- ChaincodeEvent API ----------------------

// SetEvent documentation can be found in interfaces.go
func (stub *ChaincodeStub) SetEvent(name string, payload []byte) error {
	if name == "" {
		return errors.New("event name can not be nil string")
	}
	stub.chaincodeEvent = &pb.PtnChaincodeEvent{EventName: name, Payload: payload}
	return nil
}

//---------- Deposit API ----------
//func (stub *ChaincodeStub) GetSystemConfig(key string) (string, error) {
func (stub *ChaincodeStub) GetSystemConfig() (*modules.GlobalProperty, error) {
	return stub.handler.handleGetSystemConfig(stub.ChannelId, stub.TxID)
}
func (stub *ChaincodeStub) GetInvokeAddress() (common.Address, error) {
	invokeAddr, _, _, _, _, err := stub.GetInvokeParameters()
	return invokeAddr, err
}
func (stub *ChaincodeStub) GetInvokeTokens() ([]*modules.InvokeTokens, error) {
	_, invokeTokens, _, _, _, err := stub.GetInvokeParameters()
	return invokeTokens, err
}
func (stub *ChaincodeStub) GetContractAllState() (map[string]*modules.ContractStateValue, error) {
	return stub.handler.handleGetContractAllState(stub.ChannelId, stub.TxID, stub.ContractId)
}
func (stub *ChaincodeStub) GetInvokeFees() (*modules.AmountAsset, error) {
	_, _, invokeFees, _, _, err := stub.GetInvokeParameters()
	return invokeFees, err
}
func (stub *ChaincodeStub) GetContractID() ([]byte, string) {
	addr := common.NewAddress(stub.ContractId, common.ContractHash)
	return stub.ContractId, addr.Str()
}

//获得该合约的Token余额
func (stub *ChaincodeStub) GetTokenBalance(address string, token *modules.Asset) ([]*modules.InvokeTokens, error) {
	return stub.handler.handleGetTokenBalance(address, token, stub.ContractId, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) GetStableTransactionByHash(txHash string) (*modules.Transaction, error) {
	return stub.handler.handleGetStableTransactionByHash(txHash, stub.ContractId, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) GetStableUnit(unitHash string, unitNumber uint64) (*modules.Unit, error) {
	return stub.handler.handleGetStableUnit(unitHash, unitNumber, stub.ContractId, stub.ChannelId, stub.TxID)
}

func (stub *ChaincodeStub) DefineToken(tokenType byte, define []byte, creator string) error {
	if !common.IsSystemContractId(stub.ContractId) {
		return ERROR_ONLY_SYS_CONTRACT
	}
	return stub.handler.handleDefineToken(tokenType, define, creator, stub.ContractId, stub.ChannelId, stub.TxID)
}

//增发一种之前已经定义好的Token
//如果是ERC20增发，则uniqueId为空，如果是ERC721增发，则必须指定唯一的uniqueId
func (stub *ChaincodeStub) SupplyToken(assetId []byte, uniqueId []byte, amt uint64, creator string) error {
	if !common.IsSystemContractId(stub.ContractId) {
		return ERROR_ONLY_SYS_CONTRACT
	}
	return stub.handler.handleSupplyToken(assetId, uniqueId, amt, creator, stub.ContractId, stub.ChannelId, stub.TxID)
}

//将合约上锁定的某种Token支付出去
func (stub *ChaincodeStub) PayOutToken(addr string, invokeTokens *modules.AmountAsset, lockTime uint32) error {
	//TODO Devin return stub.handler.handlePayOutToken(  stub.ContractId, stub.ChannelId, stub.TxID)
	return stub.handler.handlePayOutToken("", addr, invokeTokens, lockTime, stub.ContractId, stub.ChannelId,
		stub.TxID)
}

// 根据证书ID获得证书字节数据，不包含BEGIN和EN两行字符
func (stub *ChaincodeStub) GetRequesterCert() (certBytes []byte, err error) {
	if len(stub.args) <= 1 {
		return nil, fmt.Errorf("args error: has no cert info")
	}
	// query cert bytes
	intCertID := new(big.Int).SetBytes(stub.args[1])
	if intCertID == nil {
		return nil, fmt.Errorf("certid bytes error")
	}
	// check ca state
	caCert, err := stub.handler.handleGetCACert(stub.ChannelId, stub.TxID)
	if err != nil {
		return nil, fmt.Errorf("query ca certificate error (%s)", err.Error())
	}
	if caCert.SerialNumber.String() == intCertID.String() {
		return caCert.Raw, nil
	}
	// other certs
	key := dagConstants.CERT_BYTES_SYMBOL + intCertID.String()
	resBytes, err := stub.handler.handleGetCertState(key, stub.ChannelId, stub.TxID)
	if err != nil {
		return nil, err
	}
	if len(resBytes) <= 0 {
		return nil, fmt.Errorf("query no cert bytes for certid(%s)", intCertID.String())
	}
	certDBInfo := modules.CertBytesInfo{}
	if err := json.Unmarshal(resBytes, &certDBInfo); err != nil {
		return nil, err
	}
	return certDBInfo.Raw, nil
}

func (stub *ChaincodeStub) IsRequesterCertValid() (bool, error) {
	if len(stub.args) <= 1 {
		return false, fmt.Errorf("args error: has no cert info")
	}
	caller, err := stub.GetInvokeAddress()
	if err != nil {
		return false, err
	}

	return stub.handler.handlerCheckCertValidation(caller.String(), stub.args[1], stub.ChannelId, stub.TxID)
}

// ------------- Logging Control and Chaincode Loggers ---------------

// As independent programs, Go language chaincodes can use any logging
// methodology they choose, from simple fmt.Printf() to os.Stdout, to
// decorated logs created by the author's favorite logging package. The
// chaincode "shim" interface, however, is defined by the palletone
// and implements its own logging methodology. This methodology currently
// includes severity-based logging control and a standard way of decorating
// the logs.
//
// The facilities defined here allow a Go language chaincode to control the
// logging level of its shim, and to create its own logs formatted
// consistently with, and temporally interleaved with the shim logs without
// any knowledge of the underlying implementation of the shim, and without any
// other package requirements. The lack of package requirements is especially
// important because even if the chaincode happened to explicitly use the same
// logging package as the shim, unless the chaincode is physically included as
// part of the source code tree it could actually end up
// using a distinct binary instance of the logging package, with different
// formats and severity levels than the binary package used by the shim.
//
// Another approach that might have been taken, and could potentially be taken
// in the future, would be for the chaincode to supply a logging object for
// the shim to use, rather than the other way around as implemented
// here. There would be some complexities associated with that approach, so
// for the moment we have chosen the simpler implementation below. The shim
// provides one or more abstract logging objects for the chaincode to use via
// the NewLogger() API, and allows the chaincode to control the severity level
// of shim logs using the SetLoggingLevel() API.

// LoggingLevel is an enumerated type of severity levels that control
// chaincode logging.
//type LoggingLevel logging.Level
//
//// These constants comprise the LoggingLevel enumeration
//const (
//	LogDebug    = LoggingLevel(logging.DEBUG)
//	LogInfo     = LoggingLevel(logging.INFO)
//	LogNotice   = LoggingLevel(logging.NOTICE)
//	LogWarning  = LoggingLevel(logging.WARNING)
//	LogError    = LoggingLevel(logging.ERROR)
//	LogCritical = LoggingLevel(logging.CRITICAL)
//)
//
//var shimLoggingLevel = LogInfo // Necessary for correct initialization; See Start()

// SetLoggingLevel allows a Go language chaincode to set the logging level of
// its shim.
//func SetLoggingLevel(level LoggingLevel) {
//	shimLoggingLevel = level
//	logging.SetLevel(logging.Level(level), "shim")
//}

// LogLevel converts a case-insensitive string chosen from CRITICAL, ERROR,
// WARNING, NOTICE, INFO or DEBUG into an element of the LoggingLevel
// type. In the event of errors the level returned is LogError.
//func LogLevel(levelString string) (LoggingLevel, error) {
//	l, err := logging.LogLevel(levelString)
//	level := LoggingLevel(l)
//	if err != nil {
//		level = LogError
//	}
//	return level, err
//}

// ------------- Chaincode Loggers ---------------
/*
// ChaincodeLogger is an abstraction of a logging object for use by
// chaincodes. These objects are created by the NewLogger API.
type ChaincodeLogger struct {
	log log.ILogger
}

// NewLogger allows a Go language chaincode to create one or more logging
// objects whose logs will be formatted consistently with, and temporally
// interleaved with the logs created by the shim interface. The logs created
// by this object can be distinguished from shim logs by the name provided,
// which will appear in the logs.
func NewLogger(name string) *ChaincodeLogger {
	return &ChaincodeLogger{log.New(name)}
}

// SetLevel sets the logging level for a chaincode log. Note that currently
// the levels are actually controlled by the name given when the log is
// created, so logs should be given unique names other than "shim".
//func (c *ChaincodeLogger) SetLevel(level LoggingLevel) {
//	logging.SetLevel(logging.Level(level), c.log.Module)
//}

// IsEnabledFor returns true if the log is enabled to creates logs at the
// given logging level.
//func (c *ChaincodeLogger) IsEnabledFor(level LoggingLevel) bool {
//	return c.log.IsEnabledFor(logging.Level(level))
//}

// Debug logs will only appear if the ChaincodeLogger LoggingLevel is set to
// LogDebug.
func (c *ChaincodeLogger) Debug(args ...interface{}) {
	c.log.Debug("Chaincode", args...)
}

// Info logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogInfo or LogDebug.
func (c *ChaincodeLogger) Info(args ...interface{}) {
	c.log.Info("Chaincode", args...)
}

// Notice logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Notice(args ...interface{}) {
	c.log.Info("Chaincode", args...)
}

// Warning logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogWarning, LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Warning(args ...interface{}) {
	c.log.Warn("Chaincode", args...)
}

// Error logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogError, LogWarning, LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Error(args ...interface{}) {
	c.log.Error("Chaincode", args...)
}

// Critical logs always appear; They can not be disabled.
func (c *ChaincodeLogger) Critical(args ...interface{}) {
	c.log.Error("Chaincode", args...)
}

// Debugf logs will only appear if the ChaincodeLogger LoggingLevel is set to
// LogDebug.
func (c *ChaincodeLogger) Debugf(format string, args ...interface{}) {
	c.log.Debugf(format, args...)
}

// Infof logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogInfo or LogDebug.
func (c *ChaincodeLogger) Infof(format string, args ...interface{}) {
	c.log.Infof(format, args...)
}

// Noticef logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Noticef(format string, args ...interface{}) {
	c.log.Infof(format, args...)
}

// Warningf logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogWarning, LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Warningf(format string, args ...interface{}) {
	c.log.Warnf(format, args...)
}

// Errorf logs will appear if the ChaincodeLogger LoggingLevel is set to
// LogError, LogWarning, LogNotice, LogInfo or LogDebug.
func (c *ChaincodeLogger) Errorf(format string, args ...interface{}) {
	c.log.Errorf(format, args...)
}

// Criticalf logs always appear; They can not be disabled.
func (c *ChaincodeLogger) Criticalf(format string, args ...interface{}) {
	c.log.Errorf(format, args...)
}
*/
//func (stub *ChaincodeStub) GetDecorations() map[string][]byte {
//	return stub.decorations
//}

// GetQueryResult documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetQueryResult(query string) (StateQueryIteratorInterface, error) {
// // Access public data by setting the collection to empty string
// collection := ""
// response, err := stub.handler.handleGetQueryResult(collection, query, stub.ChannelId, stub.TxID)
// if err != nil {
//    return nil, err
// }
// return &StateQueryIterator{CommonIterator: &CommonIterator{stub.handler, stub.ChannelId, stub.TxID, response, 0}},
// nil
//}

// CommonIterator documentation can be found in interfaces.go
//type CommonIterator struct {
// handler    *Handler
// channelId  string
// txid       string
// response   *pb.QueryResponse
// currentLoc int
//}

// StateQueryIterator documentation can be found in interfaces.go
//type StateQueryIterator struct {
// *CommonIterator
//}

// HistoryQueryIterator documentation can be found in interfaces.go
//type HistoryQueryIterator struct {
// *CommonIterator
//}

//type resultType uint8

//const (
// STATE_QUERY_RESULT resultType = iota + 1
// HISTORY_QUERY_RESULT
//)

//func (stub *ChaincodeStub) handleGetStateByRange(collection, startKey, endKey string) (StateQueryIteratorInterface, error) {
// response, err := stub.handler.handleGetStateByRange(collection, startKey, endKey, stub.ChannelId, stub.TxID)
// if err != nil {
//    return nil, err
// }
// return &StateQueryIterator{CommonIterator: &CommonIterator{stub.handler, stub.ChannelId, stub.TxID, response, 0}},
// nil
//}

// GetStateByRange documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetStateByRange(startKey, endKey string) (StateQueryIteratorInterface, error) {
// if startKey == "" {
//    startKey = emptyKeySubstitute
// }
// if err := validateSimpleKeys(startKey, endKey); err != nil {
//    return nil, err
// }
// collection := ""
// return stub.handleGetStateByRange(collection, startKey, endKey)
//}

//CreateCompositeKey documentation can be found in interfaces.go
//func (stub *ChaincodeStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
// return createCompositeKey(objectType, attributes)
//}

//SplitCompositeKey documentation can be found in interfaces.go
//func (stub *ChaincodeStub) SplitCompositeKey(compositeKey string) (string, []string, error) {
// return splitCompositeKey(compositeKey)
//}

//func createCompositeKey(objectType string, attributes []string) (string, error) {
// if err := validateCompositeKeyAttribute(objectType); err != nil {
//    return "", err
// }
// ck := compositeKeyNamespace + objectType + string(minUnicodeRuneValue)
// for _, att := range attributes {
//    if err := validateCompositeKeyAttribute(att); err != nil {
//       return "", err
//    }
//    ck += att + string(minUnicodeRuneValue)
// }
// return ck, nil
//}

//func splitCompositeKey(compositeKey string) (string, []string, error) {
// componentIndex := 1
// components := []string{}
// for i := 1; i < len(compositeKey); i++ {
//    if compositeKey[i] == minUnicodeRuneValue {
//       components = append(components, compositeKey[componentIndex:i])
//       componentIndex = i + 1
//    }
// }
// return components[0], components[1:], nil
//}

//func validateCompositeKeyAttribute(str string) error {
// if !utf8.ValidString(str) {
//    return errors.Errorf("not a valid utf8 string: [%x]", str)
// }
// for index, runeValue := range str {
//    if runeValue == minUnicodeRuneValue || runeValue == maxUnicodeRuneValue {
//       return errors.Errorf(`input contain unicode %#U starting at position [%d]. %#U and %#U are not allowed
// in the input attribute of a composite key`,
//          runeValue, index, minUnicodeRuneValue, maxUnicodeRuneValue)
//    }
// }
// return nil
//}

//To ensure that simple keys do not go into composite key namespace,
//we validate simplekey to check whether the key starts with 0x00 (which
//is the namespace for compositeKey). This helps in avoding simple/composite
//key collisions.
//func validateSimpleKeys(simpleKeys ...string) error {
// for _, key := range simpleKeys {
//    if len(key) > 0 && key[0] == compositeKeyNamespace[0] {
//       return errors.Errorf(`first character of the key [%s] contains a null character which is not allowed`, key)
//    }
// }
// return nil
//}

//GetStateByPartialCompositeKey function can be invoked by a chaincode to query the
//state based on a given partial composite key. This function returns an
//iterator which can be used to iterate over all composite keys whose prefix
//matches the given partial composite key. This function should be used only for
//a partial composite key. For a full composite key, an iter with empty response
//would be returned.
//func (stub *ChaincodeStub) GetStateByPartialCompositeKey(objectType string,
// attributes []string) (StateQueryIteratorInterface, error) {
// collection := ""
// if partialCompositeKey, err := stub.CreateCompositeKey(objectType, attributes); err == nil {
//    return stub.handleGetStateByRange(collection, partialCompositeKey, partialCompositeKey+string(maxUnicodeRuneValue))
// } else {
//    return nil, err
// }
//}

// HasNext documentation can be found in interfaces.go
//func (iter *CommonIterator) HasNext() bool {
// if iter.currentLoc < len(iter.response.Results) || iter.response.HasMore {
//    return true
// }
// return false
//}

// getResultsFromBytes deserializes QueryResult and return either a KV struct
// or KeyModification depending on the result type (i.e., state (range/execute)
// query, history query). Note that commonledger.QueryResult is an empty golang
// interface that can hold values of any type.
//func (iter *CommonIterator) getResultFromBytes(queryResultBytes *pb.QueryResultBytes,
//rType resultType) (commonledger.QueryResult, error) {
//glh
/*
   if rType == STATE_QUERY_RESULT {
      stateQueryResult := &queryresult.KV{}
      if err := proto.Unmarshal(queryResultBytes.ResultBytes, stateQueryResult); err != nil {
         return nil, errors.Wrap(err, "error unmarshaling result from bytes")
      }
      return stateQueryResult, nil

   } else if rType == HISTORY_QUERY_RESULT {
      historyQueryResult := &queryresult.KeyModification{}
      if err := proto.Unmarshal(queryResultBytes.ResultBytes, historyQueryResult); err != nil {
         return nil, err
      }
      return historyQueryResult, nil
   }
   return nil, errors.New("wrong result type")
*/
//return nil, errors.New("glh unfinished")
//}
//
//func (iter *CommonIterator) fetchNextQueryResult() error {
// if response, err := iter.handler.handleQueryStateNext(iter.response.Id, iter.channelId, iter.txid); err == nil {
//    iter.currentLoc = 0
//    iter.response = response
//    return nil
// } else {
//    return err
// }
//}

// nextResult returns the next QueryResult (i.e., either a KV struct or KeyModification)
// from the state or history query iterator. Note that commonledger.QueryResult is an
// empty golang interface that can hold values of any type.
//func (iter *CommonIterator) nextResult(rType resultType) (commonledger.QueryResult, error) {
// if iter.currentLoc < len(iter.response.Results) {
//    // On valid access of an element from cached results
//    queryResult, err := iter.getResultFromBytes(iter.response.Results[iter.currentLoc], rType)
//    if err != nil {
//       log.Errorf("Failed to decode query results: %+v", err)
//       return nil, err
//    }
//    iter.currentLoc++
//
//    if iter.currentLoc == len(iter.response.Results) && iter.response.HasMore {
//       // On access of last item, pre-fetch to update HasMore flag
//       if err = iter.fetchNextQueryResult(); err != nil {
//          log.Errorf("Failed to fetch next results: %+v", err)
//          return nil, err
//       }
//    }
//
//    return queryResult, err
// } else if !iter.response.HasMore {
//    // On call to Next() without check of HasMore
//    return nil, errors.New("no such key")
// }
//
// // should not fall through here
// // case: no cached results but HasMore is true.
// return nil, errors.New("invalid iterator state")
//}

// Close documentation can be found in interfaces.go
//func (iter *CommonIterator) Close() error {
// _, err := iter.handler.handleQueryStateClose(iter.response.Id, iter.channelId, iter.txid)
// return err
//}

// GetCreator documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetCreator() ([]byte, error) {
// return stub.creator, nil
//}

// GetTransient documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetTransient() (map[string][]byte, error) {
// return stub.transient, nil
//}

// GetBinding documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetBinding() ([]byte, error) {
// return stub.binding, nil
//}

// GetSignedProposal documentation can be found in interfaces.go
//func (stub *ChaincodeStub) GetSignedProposal() (*pb.PtnSignedProposal, error) {
// return stub.signedProposal, nil
//}
