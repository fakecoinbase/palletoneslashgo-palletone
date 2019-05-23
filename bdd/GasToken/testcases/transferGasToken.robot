*** Settings ***
Resource          setups.robot
Library           Collections

*** Test Cases ***
transferGasToken
    queryTokenHolder
    newAccounts
    transferGasTokenToNewUsers

*** Keywords ***
transferGasTokenToNewUsers
    # transfer gas token WWW to Alice
    transferPTN    ${Alice}
    wait for transaction being packaged
    ${balance}=    getBalance    ${Alice}
    Should Be Equal    ${balance}    ${amount}
    # transfer gas token WWW to Bob
    transferPTN    ${Bob}
    wait for transaction being packaged
    ${balance}=    getBalance    ${Bob}
    Should Be Equal    ${balance}    ${amount}
    # transfer gas token WWW to Carol
    transferPTN    ${Carol}
    wait for transaction being packaged
    ${balance}=    getBalance    ${Carol}
    Should Be Equal    ${balance}    ${amount}
