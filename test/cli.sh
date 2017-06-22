#! /bin/bash

# These global variables are required for common.sh
SERVER_EXE=tracko
CLIENT_EXE=trackocli
ACCOUNTS=(jae ethan igor rigel)
RICH=${ACCOUNTS[0]}

oneTimeSetUp() {
    quickSetup .test_tracko tracko-chain
}

oneTimeTearDown() {
    quickTearDown
}

#Sequence numbers for the four accounts tested
SEQ=(1 1 1 1)
#Increase the sequence number for the arg provided 
seqUp(){
    if [ -z "$1" ]; then 
        return 
    fi 
    SEQ[$1]=$((${SEQ[$1]}+1))
}

testOpeningProfiles(){
    NAMES=(AllInBits Bucky Satoshi Dummy)
    for i in "${!NAMES[@]}"; do 
        #send all the existing accounts some coins 
        ADDR[$i]=$(${CLIENT_EXE} keys get ${ACCOUNTS[$i]} --output=json | jq .address | tr -d '"')

        if [ "$i" != "0" ]; then
            TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=1000mycoin \
                --sequence=${SEQ[0]} --to=${ADDR[$i]} --name=${ACCOUNTS[0]})
            txSucceeded $? "$TX" 
            seqUp 0
        fi 

        #open the profile
        TX=$(echo qwertyuiop | ${CLIENT_EXE} tx profile-open ${NAMES[$i]} --cur=BTC \
            --amount=1mycoin --sequence=${SEQ[$i]} --name=${ACCOUNTS[$i]})
        txSucceeded $? "$TX" 
        seqUp $i
    done

    #check if the profiles have been opened
    PROFILES=$(${CLIENT_EXE} query profiles)
    for i in "${!NAMES[@]}"; do 
        assertTrue 'profile not created' "[[ $PROFILES == *"${NAMES[$i]}"* ]]"
    done
}

testDeletingProfile(){
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx profile-deactivate ${NAMES[3]} \
        --amount=1mycoin --sequence=${SEQ[3]} --name=${ACCOUNTS[3]})
    txSucceeded $? "$TX" 
    seqUp 3

    #test if profile is active
    ACTIVE=$(${CLIENT_EXE} query profile ${NAMES[3]} | jq .Active)
    assertEquals 'deleted profile still active' "$ACTIVE" "false"

    #verify it doesn't exist in the active list
    PROFILES=$(${CLIENT_EXE} query profiles)
    assertFalse 'profile should be removed from active' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"

    #verify it does exist in the inactive list
    PROFILES=$(${CLIENT_EXE} query profiles --inactive)
    assertTrue 'profile should exist on inactive' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"
}

testEditingProfile(){
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx profile-edit ${NAMES[0]} --cur=USD \
        --amount=1mycoin --sequence=${SEQ[0]} --name=${ACCOUNTS[0]})
    txSucceeded $? "$TX" 

    seqUp 0
    CUR=$(${CLIENT_EXE} query profile ${NAMES[0]} | jq .AcceptedCur | tr -d '"')
    assertEquals 'active profile should be editable' "$CUR" "USD"

    #make sure that we're prevented from editing an inactive profile
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx profile-edit ${NAMES[3]} --cur=USD \
        --amount=1mycoin --sequence=${SEQ[3]} --name=${ACCOUNTS[3]})
    #TODO write a test to ensure that delivery failed for TX here
    #txSucceeded $? "$TX" 

    CUR=$(${CLIENT_EXE} query profile ${NAMES[3]} | jq .AcceptedCur | tr -d '"')
    assertNotEquals 'inactive profile should not be editable' "$CUR" "USD"
}

testContractInvoice(){
    #Create the invoice
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx contract-open 1000.99USD --date=2017-01-01 --to=AllInBits --notes=thanks! \
        --amount=1mycoin --sequence=${SEQ[1]} --name=${ACCOUNTS[1]})
    txSucceeded $? "$TX" 
    seqUp 1 

    ID=$(${CLIENT_EXE} query invoices | jq .[0][1].ID | tr -d '"')
    CUR1=$(${CLIENT_EXE} query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)

    #Edit the invoice
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx contract-edit 1000.99CAD --id=0x$ID --date=2017-01-01 --to=AllInBits --notes=thanks! \
        --amount=1mycoin --sequence=${SEQ[1]} --name=${ACCOUNTS[1]})
    txSucceeded $? "$TX" 
    seqUp 1 
    CUR2=$(${CLIENT_EXE} query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)

    assertNotEquals "contract invoice currency should have been edited $CUR1 $CUR2" "$CUR1" "$CUR2"

    #pay the contract invoice
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx payment ${NAMES[1]} --ids=0x$ID --paid=0.5BTC --date=2017-01-01 --tx-id=FOOBTC-TX-01 \
        --amount=1mycoin --sequence=${SEQ[0]} --name=${ACCOUNTS[0]})
    txSucceeded $? "$TX" 
    seqUp 0 
    open=$(${CLIENT_EXE} query invoice 0x$ID | jq .data.Ctx.Open)
    assertEquals "Invoice should be open as not fully paid" "$open" "true"

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx payment ${NAMES[1]} --ids=0x$ID --paid=0.2454003323983133BTC --date=2017-01-01 --tx-id=FOOBTC-TX-02 \
        --amount=1mycoin --sequence=${SEQ[0]} --name=${ACCOUNTS[0]})
    txSucceeded $? "$TX" 
    seqUp 0 
    open=$(${CLIENT_EXE} query invoice 0x$ID | jq .data.Ctx.Open)
    assertNotEquals "invoice should nolonger be open" "$open" "true"

    #query the payments
    len=$(${CLIENT_EXE} query payments | jq '. | length')
    assertEquals "Payments should have two entries" 2 $len
}

testContractExpense(){
    #generate working directories for this test
    TMPDIR=(/tmp/tracko)
    DIR=($TMPDIR/retrieved)
    mkdir $TMPDIR 2> /dev/null
    mkdir $DIR 2> /dev/null

    #download an image, we'll pretend this is a receipt
    wget -q -O $TMPDIR/invoicerDoc.png \
        https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png

    #Open receipt
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx expense-open 99.99USD --date=2017-01-01 \
        --receipt=$TMPDIR/invoicerDoc.png --taxes=3.00USD --to=AllInBits --notes=transportation \
        --amount=1mycoin --sequence=${SEQ[1]} --name=${ACCOUNTS[1]})
    txSucceeded $? "$TX" 
    seqUp 1 

    #Download receipt from query
    ID2=$(${CLIENT_EXE} query invoices | jq .[1][1].ID | tr -d '"')
    err=$(${CLIENT_EXE} query invoice 0x$ID2 --download-expense=$DIR 2>&1 > /dev/null)
    assertNull "Error Non-Null Line $LINENO $err" "$err"

    assertTrue "Receipt didn't download from query" "[ -f $DIR/invoicerDoc.png ]"
}

testSums(){
    #Opening four invoices of the same USD amount for various dates
    DATES=(2017-01-02 2017-01-15 2017-02-01 2017-03-15 )
    for i in "${!NAMES[@]}"; do 
        TX=$(echo qwertyuiop | ${CLIENT_EXE} tx contract-open 1000USD --date=${DATES[$i]} --to=AllInBits --notes=thanks! \
            --amount=1mycoin --sequence=${SEQ[1]} --name=${ACCOUNTS[1]})
        txSucceeded $? "$TX" 
        #err=$((tx "contract-open --invoice-amount=1000USD --date=${DATES[$i]} --to=AllInBits --notes=thanks!" \
            #${ACCOUNTS[1]} ${SEQ[1]}) 2>&1 > /dev/null)
        #assertNull "Error Non-Null Line $LINENO $err" "$err"
        seqUp 1 
    done

    #count the number of new invoices
    len=$(${CLIENT_EXE} query invoices --date-range=2017-01-02: | jq '. | length')
    assertEquals "Invoices should have four entries" 4 $len

    #get the sum of the invoice amount due
    SUM1=$(${CLIENT_EXE} query invoices --sum --date-range=2017-01-02: | jq .SumDue.Amount | tr -d '"')

    #pay a bit of the invoices off
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx payment ${NAMES[1]} --date-range=2017-01-02: --paid=2BTC --date=2017-03-15 --tx-id=FOOBTC-TX-03 \
        --amount=1mycoin --sequence=${SEQ[0]} --name=${ACCOUNTS[0]})
    txSucceeded $? "$TX" 
    #err=$((tx "payment --receiver-name=${NAMES[1]} --date-range=2017-01-02: --paid=2BTC --date=2017-03-15 --tx-id=FOOBTC-TX-03" \
        #${ACCOUNTS[0]} ${SEQ[0]}) 2>&1 > /dev/null)
    #assertNull "Error Non-Null Line $LINENO $err" "$err"
    seqUp 0 

    #check that some of the invoices are closed from that last payment
    len=$(${CLIENT_EXE} query invoices --date-range=2017-01-02: --type open | jq '. | length')
    assertEquals "Invoices should have two entries" 2 $len

    #check the remainded that the new sum is 2 than old sum
    SUM2=$(${CLIENT_EXE} query invoices --sum --date-range=2017-01-02:2017-01-30 | jq .SumDue.Amount | tr -d '"')
    SUM2Plus2=$(echo "$SUM2 + 2" | bc)
    assertTrue "Sums are not consistent" "[ "$SUM1" == "$SUM2Plus2" ]"
}

# load common and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
BCDIR="$GOPATH/src/github.com/tendermint/basecoin/tests/cli"
. $BCDIR/common.sh
. $DIR/shunit2
