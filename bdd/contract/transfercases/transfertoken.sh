#!/usr/bin/expect
#!/bin/bash
set timeout 30
set tokenHolder [lindex $argv 0]
set another [lindex $argv 1]
spawn ../../node/gptn --exec "wallet.transferPTN($tokenHolder,$another,500,1)" attach ../../node/palletone/gptn.ipc
expect "Passphrase:"
send "1\n"
interact
