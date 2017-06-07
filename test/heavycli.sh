#! /bin/bash

oneTimeSetUp(){
    # start tendermint
    WORKDIR=(~/.tracko)
    rm -rf $WORKDIR
    tracko init
    tracko unsafe_reset_all
    tracko start  > /dev/null &
    sleep 5
    pid_basecoin=$!
}

oneTimeTearDown() {
    # close tendermint/clean temp files
    rm -rf /tmp/tracko/
    kill -9 $pid_basecoin
    echo "cleaning up bash test"
}

testOpeningProfiles(){
    
    NAMES=(AllInBits Bucky Satoshi Dummy)
    
    for i in "${!NAMES[@]}"; do 
        #make some keys and send them some mycoin 
        TESTKEY[$i]=testkey$i.json
        tracko key new > $WORKDIR/${TESTKEY[$i]}
        ADDR[$i]=$(cat $WORKDIR/${TESTKEY[$i]} | jq .address | tr -d '"')
        
        err=$(tracko tx send --from key.json --to ${ADDR[$i]} --amount 1000mycoin > /dev/null 2>&1)
        assertNull "Error Non-Null Line "$LINENO "$err"
    
        #open the profile
        err=$(tracko tx invoicer profile-open ${NAMES[$i]} --cur BTC \
            --from ${TESTKEY[$i]} --amount 1mycoin > /dev/null 2>&1)
        assertNull "Error Non-Null Line"$LINENO "$err"
    done
    
    #check if the profiles have been opened
    PROFILES=$(tracko query profiles)
    for i in "${!NAMES[@]}"; do 
       assertTrue 'profile not created' "[[ $PROFILES == *"${NAMES[$i]}"* ]]"
    done
}
    
testDeletingProfile(){
    err=$(tracko tx invoicer profile-deactivate --from ${TESTKEY[3]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line "$LINENO "$err"
    
    #test if profile is active
    ACTIVE=$(tracko query profile ${NAMES[3]} | jq .Active)
    assertEquals 'deleted profile still active' "$ACTIVE" "false"

    #verify it doesn't exist in the active list
    PROFILES=$(tracko query profiles)
    assertFalse 'profile should be removed from active' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"
    
    #verify it does exist in the inactive list
    PROFILES=$(tracko query profiles --inactive)
    assertTrue 'profile should exist on inactive' "[[ "${PROFILES}" == *"${NAMES[3]}"* ]]"
}

testEditingProfile(){
    err=$(tracko tx invoicer profile-edit --from ${TESTKEY[0]} --cur USD --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
    CUR=$(tracko query profile ${NAMES[0]} | jq .AcceptedCur | tr -d '"')
    assertEquals 'active profile should be editable' "$CUR" "USD"

    #make sure that we're prevented from editing an inactive profile
    err=$(tracko tx invoicer profile-edit --from ${TESTKEY[3]} --cur USD --amount 1mycoin > /dev/null 2>&1)
    #TODO change to NotNULL this next line should actually be generating an error
    assertNull "Error Non-Null Line"$LINENO "$err"
    CUR=$(tracko query profile ${NAMES[3]} | jq .AcceptedCur | tr -d '"')
    assertNotEquals 'inactive profile should not be editable' "$CUR" "USD"
}

testContractInvoice(){
    #Create the invoice
    err=$(tracko tx invoicer contract-open 1000.99USD --date 2017-01-01 \
        --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line "$LINENO "$err"

    ID=$(tracko query invoices | jq .[0][1].ID | tr -d '"')
    CUR1=$(tracko query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)
    
    #Edit the invoice
    err=$(tracko tx invoicer contract-edit 1000.99CAD --id 0x$ID --date 2017-01-01 \
        --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line "$LINENO "$err"
    CUR2=$(tracko query invoice 0x$ID | jq .data.Ctx.Invoiced.CurTime.Cur)
    
    assertNotEquals 'contract invoice currency should have been edited' "$CUR1" "$CUR2"

    #pay the contract invoice
    err=$(tracko tx invoicer payment Bucky --ids 0x$ID --paid 0.5BTC --date 2017-01-01 \
        --tx-id "FOOBTC-TX-01" --from ${TESTKEY[0]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
    open=$(tracko query invoice 0x$ID | jq .data.Ctx.Open)
    assertEquals "Invoice should be open as not fully paid" "$open" "true"
    
    err=$(tracko tx invoicer payment Bucky --ids 0x$ID --paid 0.2454003323983133BTC \
        --date 2017-01-01 --tx-id "FOOBTC-TX-02" --from ${TESTKEY[0]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
    open=$(tracko query invoice 0x$ID | jq .data.Ctx.Open)
    assertNotEquals "invoice should nolonger be open" "$open" "true"
    
    #query the payments
    len=$(tracko query payments | jq '. | length')
    assertEquals "Payments should have two entries" 2 $len
}

testContractExpense(){
    #generate working directories for this test
    DIR1=(/tmp/tracko)
    DIR2=($DIR1/retrieved)
    mkdir $DIR1 ; mkdir $DIR2

    #download an image, we'll pretend this is a receipt
    wget -q -O $DIR1/invoicerDoc.png \
        https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png

    #Open receipt
    err=$(tracko tx invoicer expense-open 99.99USD --date 2017-01-01 --receipt $DIR1/invoicerDoc.png \
        --taxes 3.00USD --to AllInBits --notes transportation --from ${TESTKEY[1]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
   
    #Download receipt from query
    ID2=$(tracko query invoices | jq .[1][1].ID | tr -d '"')
    err=$(tracko query invoice 0x$ID2 --download-expense $DIR2 > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
    
    assertTrue "Receipt didn't download from query" "[ -f $DIR2/invoicerDoc.png ]"
}

testSums(){
    #Opening four invoices of the same USD amount for various dates
    DATES=(2017-01-02 2017-01-15 2017-02-01 2017-03-15 )
    for i in "${!NAMES[@]}"; do 
        err=$(tracko tx invoicer contract-open 1000USD --date ${DATES[$i]} \
            --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null 2>&1)
        assertNull "Error Non-Null Line"$LINENO "$err"
    done
    
    #count the number of new invoices
    len=$(tracko query invoices --date-range 2017-01-02: | jq '. | length')
    assertEquals "Invoices should have four entries" 4 $len
    
    #get the sum of the invoice amount due
    SUM1=$(tracko query invoices --sum --date-range 2017-01-02: | jq .SumDue.Amount | tr -d '"')
    
    #pay a bit of the invoices off
    err=$(tracko tx invoicer payment Bucky --date-range 2017-01-02: --paid 2BTC --date 2017-03-15 \
        --tx-id "FOOBTC-TX-03" --from ${TESTKEY[0]} --amount 1mycoin > /dev/null 2>&1)
    assertNull "Error Non-Null Line"$LINENO "$err"
   
    #check that some of the invoices are closed from that last payment
    len=$(tracko query invoices --date-range 2017-01-02: --type open | jq '. | length')
    assertEquals "Invoices should have two entries" 2 $len
    
    #check the remainded that the new sum is 2 than old sum
    SUM2=$(tracko query invoices --sum --date-range 2017-01-02: | jq .SumDue.Amount | tr -d '"')
    SUM2Plus2=$(echo "$SUM2 + 2" | bc)
    assertTrue "Sums are not consistent" "[ "$SUM1" == "$SUM2Plus2" ]"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
