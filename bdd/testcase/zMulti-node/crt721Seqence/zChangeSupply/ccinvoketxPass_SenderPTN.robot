*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA114

*** Test Cases ***
Scenario: 721 Contract - Change token then supply token
    [Documentation]    Verify Sender's PTN and token
    Given Send the new address PTN
    And CcinvokePass normal
    ${ret2}    When Supply token of 721 contract before change supply
    ${PTN2}    And Request getbalance after supply token
    And Change supply address to new address
    ${PTN1}    And Request getbalance before supply token
    ${PTNGAIN}    And Calculate gain
    And Supply token of 721 contract after change supply
    ${PTN3}    And Request getbalance after change supply
    Then Assert gain    ${PTN1}    ${PTN3}    ${PTNGAIN}
    #And Genesis address supply token of 721 contract
    #And Request getbalance after genesis supply token

*** Keywords ***
Send the new address PTN
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${jsonRes}    newAccount
    Set Suite Variable    ${reciever}    ${jsonRes['result']}
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${reciever}    100000    ${PTNPoundage}    ${locktime}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    sleep    4

CcinvokePass normal
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    sleep    4
    [Return]    ${resp}

Supply token of 721 contract before change supply
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Request getbalance after supply token
    #normalCcqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${reciever}    ${mutiHost1}
    ${key}    getTokenId    ${preTokenId}    ${result2['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})+1
    Should Not Contain    ${result2['result']}    ${tokenCommonId}-6
    [Return]    ${PTN2}

Change supply address to new address
    ${ccList}    Create List    ${changeSupplyMethod}    ${preTokenId}    ${reciever}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${geneAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}

Request getbalance before supply token
    sleep    4
    ${result1}    getBalance    ${reciever}    ${mutiHost1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    [Return]    ${PTN1}

Calculate gain
    #${GAIN}    Evaluate    ${PTNAmount}-${PTNPoundage}
    ${PTNGAIN}    countRecieverPTN    ${PTNPoundage}
    #${PTNGAIN}    Evaluate    decimal.Decimal('${PTNAmount}')-decimal.Decimal('${PTNPoundage}')    decimal
    [Return]    ${PTNGAIN}

Supply token of 721 contract after change supply
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Request getbalance after change supply
    sleep    4
    ${PTN3}    ${result3}    normalGetBalance    ${reciever}    ${mutiHost1}
    ${key}    getTokenId    ${preTokenId}    ${result3['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})
    : FOR    ${num}    IN RANGE    6    ${len}    1
    \    ${voteToken}    Get From Dictionary    ${result3['result']}    ${tokenCommonId}-${num}
    \    log    ${tokenCommonId}-${num}
    \    Should Be Equal As Numbers    ${voteToken}    1
    [Return]    ${PTN3}

Assert gain
    [Arguments]    ${PTN1}    ${PTN3}    ${PTNGAIN}
    ${GAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNGAIN}')    decimal
    Should Be Equal As Strings    ${PTN3}    ${GAIN}

Genesis address supply token of 721 contract
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${geneAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Request getbalance after genesis supply token
    sleep    4
    ${PTN4}    ${result4}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    ${key}    getTokenId    ${preTokenId}    ${result4['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})+1
    Should Not Contain    ${result4['result']}    ${tokenCommonId}-11
    [Return]    ${PTN4}
