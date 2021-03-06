// Copyright 2016 The go-gptn Authors
// This file is part of the go-gptn library.
//
// The go-gptn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-gptn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-gptn library. If not, see <http://www.gnu.org/licenses/>.

package console

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	//"github.com/palletone/go-palletone/core/accounts/usbwallet"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rpc"
	"github.com/robertkrimen/otto"
)

// bridge is a collection of JavaScript utility methods to bride the .js runtime
// environment and the Go RPC connection backing the remote method calls.
type bridge struct {
	client   *rpc.Client  // RPC client to execute PalletOne requests through
	prompter UserPrompter // Input prompter to allow interactive user feedback
	printer  io.Writer    // Output writer to serialize any display strings to
}

// newBridge creates a new JavaScript wrapper around an RPC client.
func newBridge(client *rpc.Client, prompter UserPrompter, printer io.Writer) *bridge {
	return &bridge{
		client:   client,
		prompter: prompter,
		printer:  printer,
	}
}

// NewAccount is a wrapper around the personal.newAccount RPC method that uses a
// non-echoing password prompt to acquire the passphrase and executes the original
// RPC method (saved in jptn.newAccount) with it to actually execute the RPC call.
func (b *bridge) NewAccount(call otto.FunctionCall) (response otto.Value) {
	var (
		password string
		confirm  string
		err      error
	)
	switch {
	// No password was specified, prompt the user for it
	case len(call.ArgumentList) == 0:
		if password, err = b.prompter.PromptPassword("Passphrase: "); err != nil {
			throwJSException(err.Error())
		}
		if confirm, err = b.prompter.PromptPassword("Repeat passphrase: "); err != nil {
			throwJSException(err.Error())
		}
		if password != confirm {
			throwJSException("passphrases don't match!")
		}

		// A single string password was specified, use that
	case len(call.ArgumentList) == 1 && call.Argument(0).IsString():
		password, _ = call.Argument(0).ToString()

		// Otherwise fail with some error
	default:
		throwJSException("expected 0 or 1 string argument")
	}
	// Password acquired, execute the call and return
	ret, err := call.Otto.Call("jptn.newAccount", nil, password)
	if err != nil {
		throwJSException(err.Error())
	}
	return ret
}

// OpenWallet is a wrapper around personal.openWallet which can interpret and
// react to certain error messages, such as the Trezor PIN matrix request.
func (b *bridge) OpenWallet(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have a wallet specified to open
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the wallet URL to open")
	}
	wallet := call.Argument(0)

	var passwd otto.Value
	if call.Argument(1).IsUndefined() || call.Argument(1).IsNull() {
		passwd, _ = otto.ToValue("")
	} else {
		passwd = call.Argument(1)
	}
	// Open the wallet and return if successful in itself
	val, err := call.Otto.Call("jptn.openWallet", nil, wallet, passwd)
	if err == nil {
		return val
	}

	// Trezor PIN matrix input requested, display the matrix to the user and fetch the data
	fmt.Fprintf(b.printer, "Look at the device for number positions\n\n")
	fmt.Fprintf(b.printer, "7 | 8 | 9\n")
	fmt.Fprintf(b.printer, "--+---+--\n")
	fmt.Fprintf(b.printer, "4 | 5 | 6\n")
	fmt.Fprintf(b.printer, "--+---+--\n")
	fmt.Fprintf(b.printer, "1 | 2 | 3\n\n")

	if input, err := b.prompter.PromptPassword("Please enter current PIN: "); err != nil {
		throwJSException(err.Error())
	} else {
		passwd, _ = otto.ToValue(input)
	}
	if val, err = call.Otto.Call("jptn.openWallet", nil, wallet, passwd); err != nil {
		throwJSException(err.Error())
	}
	return val
}

// UnlockAccount is a wrapper around the personal.unlockAccount RPC method that
// uses a non-echoing password prompt to acquire the passphrase and executes the
// original RPC method (saved in jptn.unlockAccount) with it to actually execute
// the RPC call.
func (b *bridge) UnlockAccount(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the account to unlock")
	}
	account := call.Argument(0)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value

	if call.Argument(1).IsUndefined() || call.Argument(1).IsNull() {
		fmt.Fprintf(b.printer, "Unlock account %s\n", account)
		if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
			throwJSException(err.Error())
		} else {
			passwd, _ = otto.ToValue(input)
		}
	} else {
		if !call.Argument(1).IsString() {
			throwJSException("password must be a string")
		}
		passwd = call.Argument(1)
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(2).IsDefined() && !call.Argument(2).IsNull() {
		if !call.Argument(2).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(2)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.unlockAccount", nil, account, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}
func isUnlock(call otto.FunctionCall, account otto.Value) (bool, error) {
	val, err := call.Otto.Call("personal.isUnlock", nil, account)
	if err != nil {
		return false, err
	}
	return val.ToBoolean()
}
func (b *bridge) GetUnitsByIndex(call otto.FunctionCall) (response otto.Value) {
	if !call.Argument(0).IsNumber() || !call.Argument(1).IsNumber() {
		throwJSException("the argument must be number.")
	}
	if !call.Argument(2).IsString() {
		throwJSException("the argument must be string.")
	}
	start := call.Argument(0)
	end := call.Argument(1)
	asset := call.Argument(2)
	val, err := call.Otto.Call("jptn.getUnitsByIndex", nil, start, end, asset)
	if err != nil {
		throwJSException("jay++++" + err.Error())
	}
	// Send the request to the backend and return
	return val
}

//add by wzhyuan
func (b *bridge) SignRawTransaction(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the rawtx to sign")
	}
	rawtx := call.Argument(0)

	if !call.Argument(1).IsString() {
		throwJSException("second argument must be the hashtype ")
	}
	hashtype := call.Argument(1)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value

	if call.Argument(2).IsUndefined() || call.Argument(2).IsNull() {
		fmt.Fprintf(b.printer, "Sign rawtx %s\n", rawtx)
		if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
			throwJSException(err.Error())
		} else {
			passwd, _ = otto.ToValue(input)
		}
	} else {
		if !call.Argument(2).IsString() {
			throwJSException("password must be a string")
		}
		passwd = call.Argument(2)
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(3).IsDefined() && !call.Argument(3).IsNull() {
		if !call.Argument(3).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(3)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.signRawTransaction", nil, rawtx, hashtype, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

func (b *bridge) CreateTxWithOutFee(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the tokenId")
	}
	tokenId := call.Argument(0)

	if !call.Argument(1).IsString() || !call.Argument(2).IsString() {
		throwJSException("second argument must be the tokenfrom ")
	}
	tokenfrom := call.Argument(1)
	tokento := call.Argument(2)
	amount := call.Argument(3)
	// If password is not given or is the null value, prompt the user for it
	var password otto.Value
	unlock, _ := isUnlock(call, tokenfrom)
	if !unlock {
		if call.Argument(4).IsUndefined() || call.Argument(4).IsNull() {
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				password, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(4).IsString() {
				throwJSException("password must be a string")
			}
			password = call.Argument(4)
		}
	}

	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(5).IsDefined() && !call.Argument(5).IsNull() {
		if !call.Argument(5).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(5)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.createTxWithOutFee", nil, tokenId, tokenfrom, tokento, amount, password, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

func (b *bridge) SignAndFeeTransaction(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the rawtx to sign")
	}
	rawtx := call.Argument(0)

	if !call.Argument(1).IsString() {
		throwJSException("second argument must be the gas from address ")
	}
	gasfrom := call.Argument(1)
	//to := call.Argument(3)
	gasfee := call.Argument(2)
	extra := call.Argument(3)
	// If password is not given or is the null value, prompt the user for it
	var password otto.Value
	unlock, _ := isUnlock(call, gasfrom)
	if !unlock {
		// if the password is not given or null ask the user and ensure password is a string
		if call.Argument(4).IsUndefined() || call.Argument(4).IsNull() {
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				password, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(4).IsString() {
				throwJSException("password must be a string")
			}
			password = call.Argument(4)
		}
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(5).IsDefined() && !call.Argument(5).IsNull() {
		if !call.Argument(5).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(5)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.signAndFeeTransaction", nil, rawtx, gasfrom, gasfee, extra, password, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

//add by wzhyuan
func (b *bridge) MultiSignRawTransaction(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the rawtx to sign")
	}
	rawtx := call.Argument(0)
	//lockscript := call.Argument(1)
	redeemscript := call.Argument(1)
	addr := call.Argument(2)

	if !call.Argument(3).IsString() {
		throwJSException("second argument must be the hashtype ")
	}
	hashtype := call.Argument(3)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value
	unlock, _ := isUnlock(call, addr)
	if !unlock {
		if call.Argument(4).IsUndefined() || call.Argument(4).IsNull() {
			fmt.Fprintf(b.printer, "Sign rawtx %s\n", rawtx)
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				passwd, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(4).IsString() {
				throwJSException("password must be a string")
			}
			passwd = call.Argument(4)
		}
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(5).IsDefined() && !call.Argument(3).IsNull() {
		if !call.Argument(5).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(5)
	}
	// Send the request to the backend and return
	// sencond CHAR must upper
	val, err := call.Otto.Call("jptn.multiSignRawTransaction", nil, rawtx, redeemscript, addr, hashtype, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

//add by wzhyuan
func (b *bridge) GetPtnTestCoin(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be the address string ")
	}

	if !call.Argument(1).IsString() {
		throwJSException("sencond argument must be address string to receive token")
	}
	if !call.Argument(2).IsString() {
		throwJSException("third argument must be limit of receive token")
	}

	from := call.Argument(0)
	to := call.Argument(1)
	limit := call.Argument(2)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value
	unlock, _ := isUnlock(call, from)
	if !unlock {
		if call.Argument(3).IsUndefined() || call.Argument(3).IsNull() {
			//fmt.Fprintf(b.printer, "Sign rawtx %s\n", rawtx)
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				passwd, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(3).IsString() {
				throwJSException("password must be a string")
			}
			passwd = call.Argument(3)
		}
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(4).IsDefined() && !call.Argument(4).IsNull() {
		if !call.Argument(4).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(4)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.getPtnTestCoin", nil, from, to, limit, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

//zxl add
func (b *bridge) TransferToken(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock
	if !call.Argument(0).IsString() {
		throwJSException("first argument must be asset string of transfer token")
	}
	if !call.Argument(1).IsString() {
		throwJSException("sencond argument must be account address string to unlock")
	}
	if !call.Argument(2).IsString() {
		throwJSException("third argument must be account address string to receive token")
	}
	asset := call.Argument(0)
	from := call.Argument(1)
	to := call.Argument(2)

	//3 index, amount
	//4 index, fee
	amount := call.Argument(3)
	fee := call.Argument(4)
	extra := call.Argument(5)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value
	unlock, _ := isUnlock(call, from)
	if !unlock {
		if call.Argument(6).IsUndefined() || call.Argument(6).IsNull() {
			fmt.Fprintf(b.printer, "asset: %s\n", asset)
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				passwd, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(6).IsString() {
				throwJSException("password must be a string")
			}
			passwd = call.Argument(6)
		}
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(7).IsDefined() && !call.Argument(7).IsNull() {
		if !call.Argument(7).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(7)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.transferToken", nil, asset, from, to, amount, fee, extra, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

func (b *bridge) TransferGasToken(call otto.FunctionCall) (response otto.Value) {
	// Make sure we have an account specified to unlock

	if !call.Argument(0).IsString() {
		throwJSException("sencond argument must be account address string to unlock")
	}
	if !call.Argument(1).IsString() {
		throwJSException("third argument must be account address string to receive token")
	}
	// asset := call.Argument(0)
	from := call.Argument(0)
	to := call.Argument(1)

	//3 index, amount
	//4 index, fee
	amount := call.Argument(2)
	fee := call.Argument(3)
	extra := call.Argument(4)

	// If password is not given or is the null value, prompt the user for it
	var passwd otto.Value
	unlock, _ := isUnlock(call, from)
	if !unlock {
		if call.Argument(5).IsUndefined() || call.Argument(5).IsNull() {
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				passwd, _ = otto.ToValue(input)
			}
		} else {
			if !call.Argument(5).IsString() {
				throwJSException("password must be a string")
			}
			passwd = call.Argument(5)
		}
	}
	// Third argument is the duration how long the account must be unlocked.
	duration := otto.NullValue()
	if call.Argument(6).IsDefined() && !call.Argument(6).IsNull() {
		if !call.Argument(6).IsNumber() {
			throwJSException("unlock duration must be a number")
		}
		duration = call.Argument(6)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.transferPTN", nil, from, to, amount, fee, extra, passwd, duration)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

func (b *bridge) Ccinvoketx(call otto.FunctionCall) (response otto.Value) {
	if !call.Argument(0).IsString() {
		throwJSException("1 argument must be account address string to unlock")
	}
	if !call.Argument(1).IsString() {
		throwJSException("2 argument must be account address string to receive token")
	}
	from := call.Argument(0)
	to := call.Argument(1)

	//2 index, amount
	//3 index, fee
	amount := call.Argument(2)
	fee := call.Argument(3)
	deployId := call.Argument(4)

	params := call.Argument(5)

	password := call.Argument(6)
	unlock, _ := isUnlock(call, from)
	if !unlock {
		// if the password is not given or null ask the user and ensure password is a string
		if password.IsUndefined() || password.IsNull() {
			fmt.Fprintf(b.printer, "Give password for account %s\n", from)
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				password, _ = otto.ToValue(input)
			}
		}
		if !password.IsString() {
			throwJSException("the password must be a string")
		}
	}
	timeout := otto.NullValue()
	if call.Argument(7).IsDefined() && !call.Argument(7).IsNull() {
		timeout = call.Argument(7)
	}
	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.ccinvoketx", nil, from, to, amount, fee, deployId, params, password, timeout)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

// TransferPtn is a wrapper around the personal.TransferPtn RPC method that
// uses a non-echoing password prompt to acquire the passphrase and executes the
// original RPC method (saved in jptn.TransferPtn) with it to actually execute
// the RPC call.
// append by albert·gou
func (b *bridge) TransferPtn(call otto.FunctionCall) (response otto.Value) {
	var (
		from     = call.Argument(0)
		to       = call.Argument(1)
		amount   = call.Argument(2)
		text     = call.Argument(3)
		password = call.Argument(4)
	)

	if !from.IsString() {
		throwJSException("first argument must be the account")
	}
	if !to.IsString() {
		throwJSException("second argument must be the account")
	}
	if !amount.IsNumber() {
		throwJSException("third argument must be the amount")
	}

	if text.IsDefined() && !text.IsNull() {
		if !text.IsString() {
			throwJSException("text must be a string")
		}
	}

	// if the password is not given or null ask the user and ensure password is a string
	if password.IsUndefined() || password.IsNull() {
		fmt.Fprintf(b.printer, "Give password for account %s\n", from)
		if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
			throwJSException(err.Error())
		} else {
			password, _ = otto.ToValue(input)
		}
	}
	if !password.IsString() {
		throwJSException("third argument must be the password to unlock the account")
	}

	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.transferPtn", nil, from, to, amount, text, password)
	if err != nil {
		throwJSException(err.Error())
	}

	return val
}
func (b *bridge) GetPublicKey(call otto.FunctionCall) (response otto.Value) {
	var (
		addr     = call.Argument(0)
		password = call.Argument(1)
	)

	if !addr.IsString() {
		throwJSException("first argument must be the account")
	}
	unlock, _ := isUnlock(call, addr)
	if !unlock {
		// if the password is not given or null ask the user and ensure password is a string
		if password.IsUndefined() || password.IsNull() {
			fmt.Fprintf(b.printer, "Give password for account %s\n", addr)
			if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
				throwJSException(err.Error())
			} else {
				password, _ = otto.ToValue(input)
			}
		}
		if !password.IsString() {
			throwJSException("the password must be a string")
		}
	} // Send the request to the backend and return
	val, err := call.Otto.Call("jptn.getPublicKey", nil, addr, password)
	if err != nil {
		throwJSException(err.Error())
	}

	return val
}

// Sign is a wrapper around the personal.sign RPC method that uses a non-echoing password
// prompt to acquire the passphrase and executes the original RPC method (saved in
// jptn.sign) with it to actually execute the RPC call.
func (b *bridge) Sign(call otto.FunctionCall) (response otto.Value) {
	var (
		message = call.Argument(0)
		account = call.Argument(1)
		passwd  = call.Argument(2)
	)

	if !message.IsString() {
		throwJSException("first argument must be the message to sign")
	}
	if !account.IsString() {
		throwJSException("second argument must be the account to sign with")
	}

	// if the password is not given or null ask the user and ensure password is a string
	if passwd.IsUndefined() || passwd.IsNull() {
		fmt.Fprintf(b.printer, "Give password for account %s\n", account)
		if input, err := b.prompter.PromptPassword("Passphrase: "); err != nil {
			throwJSException(err.Error())
		} else {
			passwd, _ = otto.ToValue(input)
		}
	}
	if !passwd.IsString() {
		throwJSException("third argument must be the password to unlock the account")
	}

	// Send the request to the backend and return
	val, err := call.Otto.Call("jptn.sign", nil, message, account, passwd)
	if err != nil {
		throwJSException(err.Error())
	}
	return val
}

// Sleep will block the console for the specified number of seconds.
func (b *bridge) Sleep(call otto.FunctionCall) (response otto.Value) {
	if call.Argument(0).IsNumber() {
		sleep, _ := call.Argument(0).ToInteger()
		time.Sleep(time.Duration(sleep) * time.Second)
		return otto.TrueValue()
	}
	return throwJSException("usage: sleep(<number of seconds>)")
}

// SleepBlocks will block the console for a specified number of new blocks optionally
// until the given timeout is reached.
func (b *bridge) SleepBlocks(call otto.FunctionCall) (response otto.Value) {
	var (
		blocks = int64(0)
		sleep  = int64(9999999999999999) // indefinitely
	)
	// Parse the input parameters for the sleep
	nArgs := len(call.ArgumentList)
	if nArgs == 0 {
		throwJSException("usage: sleepBlocks(<n blocks>[, max sleep in seconds])")
	}
	if nArgs >= 1 {
		if call.Argument(0).IsNumber() {
			blocks, _ = call.Argument(0).ToInteger()
		} else {
			throwJSException("expected number as first argument")
		}
	}
	if nArgs >= 2 {
		if call.Argument(1).IsNumber() {
			sleep, _ = call.Argument(1).ToInteger()
		} else {
			throwJSException("expected number as second argument")
		}
	}
	// go through the console, this will allow web3 to call the appropriate
	// callbacks if a delayed response or notification is received.
	blockNumber := func() int64 {
		result, err := call.Otto.Run("ptn.blockNumber")
		if err != nil {
			throwJSException(err.Error())
		}
		block, err := result.ToInteger()
		if err != nil {
			throwJSException(err.Error())
		}
		return block
	}
	// Poll the current block number until either it ot a timeout is reached
	targetBlockNr := blockNumber() + blocks
	deadline := time.Now().Add(time.Duration(sleep) * time.Second)

	for time.Now().Before(deadline) {
		if blockNumber() >= targetBlockNr {
			return otto.TrueValue()
		}
		time.Sleep(time.Second)
	}
	return otto.FalseValue()
}

type jsonrpcCall struct {
	Id     int64
	Method string
	Params []interface{}
}

// Send implements the web3 provider "send" method.
func (b *bridge) Send(call otto.FunctionCall) (response otto.Value) {
	// Remarshal the request into a Go value.
	JSON, _ := call.Otto.Object("JSON")
	reqVal, err := JSON.Call("stringify", call.Argument(0))
	if err != nil {
		throwJSException(err.Error())
	}
	var (
		rawReq = reqVal.String()
		dec    = json.NewDecoder(strings.NewReader(rawReq))
		reqs   []jsonrpcCall
		batch  bool
	)
	dec.UseNumber() // avoid float64s
	if rawReq[0] == '[' {
		batch = true
		dec.Decode(&reqs)
	} else {
		batch = false
		reqs = make([]jsonrpcCall, 1)
		dec.Decode(&reqs[0])
	}

	// Execute the requests.
	resps, _ := call.Otto.Object("new Array()")
	for _, req := range reqs {
		resp, _ := call.Otto.Object(`({"jsonrpc":"2.0"})`)
		resp.Set("id", req.Id)
		var result json.RawMessage
		err = b.client.Call(&result, req.Method, req.Params...)
		switch err := err.(type) {
		case nil:
			if result == nil {
				// Special case null because it is decoded as an empty
				// raw message for some reason.
				resp.Set("result", otto.NullValue())
			} else {
				resultVal, err := JSON.Call("parse", string(result))
				if err != nil {
					setError(resp, -32603, err.Error())
				} else {
					resp.Set("result", resultVal)
				}
			}
		case rpc.Error:
			setError(resp, err.ErrorCode(), err.Error())
		default:
			setError(resp, -32603, err.Error())
		}
		resps.Call("push", resp)
	}

	// Return the responses either to the callback (if supplied)
	// or directly as the return value.
	if batch {
		response = resps.Value()
	} else {
		response, _ = resps.Get("0")
	}
	if fn := call.Argument(1); fn.Class() == "Function" {
		fn.Call(otto.NullValue(), otto.NullValue(), response)
		return otto.UndefinedValue()
	}
	return response
}

func setError(resp *otto.Object, code int, msg string) {
	resp.Set("error", map[string]interface{}{"code": code, "message": msg})
}

// throwJSException panics on an otto.Value. The Otto VM will recover from the
// Go panic and throw msg as a JavaScript error.
func throwJSException(msg interface{}) otto.Value {
	val, err := otto.ToValue(msg)
	if err != nil {
		log.Error("Failed to serialize JavaScript exception", "exception", msg, "err", err)
	}
	panic(val)
}
