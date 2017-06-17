#! /bin/bash

oneTimeSetUp(){
    # start tendermint
    TR=~/.tracko
    rm -rf $TR
    tracko init
    tracko unsafe_reset_all

    LC=~/.trackocli
    TMPDIR=(/tmp/tracko)
    rm -rf $LC
    rm -rf $TMPDIR
    export BCHOME=$LC
    
    ACCOUNTS=(jae ethan craig rigel)
    for i in "${!ACCOUNTS[@]}"; do 
        newKey ${ACCOUNTS[$i]}  > /dev/null
    done
    
    GENKEY=$(trackocli keys get ${ACCOUNTS[0]} -o json | jq .pubkey.data)

    #change the genesis to the first account
    GENJSON=$(cat $TR/genesis.json)
    echo $GENJSON | jq '.app_options.accounts[0].pub_key.data='$GENKEY > $TR/genesis.json 

    #start the node with the new genesis
    export BCHOME=$TR
    tracko start  > /dev/null &
    sleep 5
    pid_basecoin=$!

    #start the light-client server
    export BCHOME=$LC
    initLightCli >/dev/null
}

oneTimeTearDown() {
    # close tendermint/clean temp files
    rm -rf $LC
    rm -rf $TMPDIR
    kill -9 $pid_basecoin
    echo "cleaning up bash test"
}

newKey(){
  assertNotNull "keyname required" "$1"
  KEYPASS=${2:-qwertyuiop}
  (echo $KEYPASS; echo $KEYPASS) | trackocli keys new $1 >/dev/null 2>&1
  assertTrue "created $1" $?
}

initLightCli(){
  SEND=${2:-y}
  (echo $SEND) | trackocli init --chainid=test_chain_id \
      --node=tcp://localhost:46657 >/dev/null 2>&1
}


tx(){
    if [ -z "$3" ]; then 
        return 
    fi 
    txAmount "$1" "$2" "$3" "1"
}

txAmount(){
    if [ -z "$4" ]; then 
        return 
    fi 
    KEYPASS=${2:-qwertyuiop}
    (echo $KEYPASS) | trackocli tx $1 --name $2 \
        --amount $4mycoin --fee 0mycoin --sequence $3 >/dev/null 2>&1
}

SEQ=(1 1 1 1)
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
        ADDR[$i]=$(trackocli keys get ${ACCOUNTS[$i]} --output=json | jq .address | tr -d '"')
        
        if [ "$i" != "0" ]; then
            err=$((txAmount "send --to ${ADDR[$i]}" "${ACCOUNTS[0]}" ${SEQ[0]} "100")  2>&1 > /dev/null)
            assertNull "Error Non-Null Line $LINENO $err" "$err"
            seqUp 0
        fi 
    
        #open the profile
        err=$((tx "profile-open --profile-name=${NAMES[$i]} --cur BTC" ${ACCOUNTS[$i]} ${SEQ[$i]}) 2>&1 > /dev/null)
        seqUp $i
        assertNull "Error Non-Null Line $LINENO $err" "$err"
    done
    
    #check if the profiles have been opened
    PROFILES=$(trackocli query profiles)
    for i in "${!NAMES[@]}"; do 
       assertTrue 'profile not created' "[[ $PROFILES == *"${NAMES[$i]}"* ]]"
    done
}

testDeletingProfile(){
    err=$((tx "profile-deactivate" ${ACCOUNTS[3]} ${SEQ[3]}) 2>&1 > /dev/null)
    seqUp 3
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    
    #test if profile is active
    ACTIVE=$(trackocli query profile ${NAMES[3]} | jq .Active)
    assertEquals 'deleted profile still active' "$ACTIVE" "false"

    #verify it doesn't exist in the active list
    PROFILES=$(trackocli query profiles)
    assertFalse 'profile should be removed from active' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"
    
    #verify it does exist in the inactive list
    PROFILES=$(trackocli query profiles --inactive)
    assertTrue 'profile should exist on inactive' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"
}

testEditingProfile(){
    err=$((tx "profile-edit --cur=USD" ${ACCOUNTS[0]} ${SEQ[0]}) 2>&1 > /dev/null)
    seqUp 0
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    CUR=$(trackocli query profile ${NAMES[0]} | jq .AcceptedCur | tr -d '"')
    assertEquals 'active profile should be editable' "$CUR" "USD"

    #make sure that we're prevented from editing an inactive profile
    err=$((tx "profile-edit --cur=USD" ${ACCOUNTS[3]} ${SEQ[3]}) 2>&1 > /dev/null)
    #TODO fix this check, need lightcli to output errors to the stderr
    #assertNotNull "Non-Null Error expected at Line $LINENO" "$err"
    CUR=$(trackocli query profile ${NAMES[3]} | jq .AcceptedCur | tr -d '"')
    assertNotEquals 'inactive profile should not be editable' "$CUR" "USD"
}

testContractInvoice(){
    #Create the invoice
    err=$((tx "contract-open --invoice-amount=1000.99USD --date=2017-01-01 --to=AllInBits --notes=thanks!" \
        ${ACCOUNTS[1]} ${SEQ[1]}) 2>&1 > /dev/null)
    seqUp 1 
    assertNull "Error Non-Null Line $LINENO $err" "$err"

    ID=$(trackocli query invoices | jq .[0][1].ID | tr -d '"')
    CUR1=$(trackocli query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)
    
    #Edit the invoice
    err=$((tx "contract-edit --invoice-amount=1000.99CAD --id=0x$ID --date=2017-01-01 --to=AllInBits --notes=thanks!" \
        ${ACCOUNTS[1]} ${SEQ[1]}) 2>&1 > /dev/null)
    seqUp 1 
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    CUR2=$(trackocli query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)
    
    assertNotEquals "contract invoice currency should have been edited $CUR1 $CUR2" "$CUR1" "$CUR2"

    #pay the contract invoice
    err=$((tx "payment --receiver-name=${NAMES[1]} --ids=0x$ID --paid=0.5BTC --date=2017-01-01 --tx-id=FOOBTC-TX-01" \
        ${ACCOUNTS[0]} ${SEQ[0]}) 2>&1 > /dev/null)
    seqUp 0 
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    open=$(trackocli query invoice 0x$ID | jq .data.Ctx.Open)
    assertEquals "Invoice should be open as not fully paid" "$open" "true"
    
    err=$((tx "payment --receiver-name=${NAMES[1]} --ids=0x$ID --paid=0.2454003323983133BTC --date=2017-01-01 --tx-id=FOOBTC-TX-02" \
        ${ACCOUNTS[0]} ${SEQ[0]}) 2>&1 > /dev/null)
    seqUp 0 
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    open=$(trackocli query invoice 0x$ID | jq .data.Ctx.Open)
    assertNotEquals "invoice should nolonger be open" "$open" "true"
    
    #query the payments
    len=$(trackocli query payments | jq '. | length')
    assertEquals "Payments should have two entries" 2 $len
}

testContractExpense(){
    #generate working directories for this test
    DIR=($TMPDIR/retrieved)
    mkdir $TMPDIR ; mkdir $DIR

    #download an image, we'll pretend this is a receipt
    wget -q -O $TMPDIR/invoicerDoc.png \
        https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png

    #Open receipt
    err=$((tx "expense-open --invoice-amount=99.99USD --date=2017-01-01 --receipt=$TMPDIR/invoicerDoc.png --taxes=3.00USD --to=AllInBits --notes=transportation" \
        ${ACCOUNTS[1]} ${SEQ[1]}) 2>&1 > /dev/null)
    seqUp 1 
    assertNull "Error Non-Null Line $LINENO $err" "$err"
   
    #Download receipt from query
    ID2=$(trackocli query invoices | jq .[1][1].ID | tr -d '"')
    err=$(trackocli query invoice 0x$ID2 --download-expense=$DIR 2>&1 > /dev/null)
    assertNull "Error Non-Null Line $LINENO $err" "$err"
    
    assertTrue "Receipt didn't download from query" "[ -f $DIR/invoicerDoc.png ]"
}

testSums(){
    #Opening four invoices of the same USD amount for various dates
    DATES=(2017-01-02 2017-01-15 2017-02-01 2017-03-15 )
    for i in "${!NAMES[@]}"; do 
        err=$((tx "contract-open --invoice-amount=1000USD --date=${DATES[$i]} --to=AllInBits --notes=thanks!" \
            ${ACCOUNTS[1]} ${SEQ[1]}) 2>&1 > /dev/null)
        seqUp 1 
        assertNull "Error Non-Null Line $LINENO $err" "$err"
    done
    
    #count the number of new invoices
    len=$(trackocli query invoices --date-range=2017-01-02: | jq '. | length')
    assertEquals "Invoices should have four entries" 4 $len
    
    #get the sum of the invoice amount due
    SUM1=$(trackocli query invoices --sum --date-range=2017-01-02: | jq .SumDue.Amount | tr -d '"')
    
    #pay a bit of the invoices off
    err=$((tx "payment --receiver-name=${NAMES[1]} --date-range=2017-01-02: --paid=2BTC --date=2017-03-15 --tx-id=FOOBTC-TX-03" \
        ${ACCOUNTS[0]} ${SEQ[0]}) 2>&1 > /dev/null)
    seqUp 0 
    assertNull "Error Non-Null Line $LINENO $err" "$err"
   
    #check that some of the invoices are closed from that last payment
    len=$(trackocli query invoices --date-range=2017-01-02: --type open | jq '. | length')
    assertEquals "Invoices should have two entries" 2 $len
    
    #check the remainded that the new sum is 2 than old sum
    SUM2=$(trackocli query invoices --sum --date-range=2017-01-02: | jq .SumDue.Amount | tr -d '"')
    SUM2Plus2=$(echo "$SUM2 + 2" | bc)
    assertTrue "Sums are not consistent" "[ "$SUM1" == "$SUM2Plus2" ]"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
#BCDIR="$GOPATH/src/github.com/tendermint/basecoin/tests/cli" #get this files directory
# load common helpers
#. $BCDIR/../basecoin/tests.common.sh
. $DIR/shunit2
