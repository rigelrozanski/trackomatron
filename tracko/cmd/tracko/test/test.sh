#! /bin/bash

echo "starting tendermint network"

# start tendermint
tracko init
tracko unsafe_reset_all
tracko start  > /dev/null &

sleep 5

pid_basecoin=$!

function cleanup {
    echo "cleaning up"
    rm -rf /tmp/tracko/
    kill -9 $pid_basecoin
}
trap cleanup EXIT


echo "opening profiles"

WORKDIR=(~/.tracko)
NAMES=(AllInBits Bucky Satoshi Dummy)

for i in "${!NAMES[@]}"; do 
    #make some keys and send them some mycoin 
    TESTKEY[$i]=testkey$i.json
    tracko key new > $WORKDIR/${TESTKEY[$i]}
    ADDR[$i]=$(cat $WORKDIR/${TESTKEY[$i]} | jq .address | tr -d '"')
    tracko tx send --from key.json --to ${ADDR[$i]} --amount 1000mycoin > /dev/null 

    #open the profile
    tracko tx invoicer profile-open ${NAMES[$i]} --cur BTC --from ${TESTKEY[$i]} --amount 1mycoin > /dev/null
done

#check if the profiles have been opened
PROFILES=$(tracko query profiles)
for i in "${!NAMES[@]}"; do 
   if [[ $PROFILES != *"${NAMES[$i]}"* ]]; then
         echo "Error Missing Profile ${NAMES[$i]}"
         echo $PROFILES
         exit 1
     fi
done


echo "deleting a profile"
echo "tracko tx invoicer profile-deactivate --from ${TESTKEY[3]} --amount 1mycoin"
tracko tx invoicer profile-deactivate --from ${TESTKEY[3]} --amount 1mycoin > /dev/null

#test if profile is active
ACTIVE=$(tracko query profile ${NAMES[3]} | jq .Active)
if [ "$ACTIVE" != "false" ]; then 
    echo "Error profile should be inactive: ${NAMES[3]}"
    echo $ACTIVE
    exit 1
fi

#verify it doesn't exist in the active list
PROFILES=$(tracko query profiles)
if [[ "${PROFILES}" == *"${NAMES[3]}"* ]]; then
    echo "Error profile should be removed: ${NAMES[3]}"
    echo $PROFILES
    exit 1
fi

#verify it does exist in the inactive list
PROFILES=$(tracko query profiles --inactive)
if [[ "${PROFILES}" != *"${NAMES[3]}"* ]]; then
    echo "Error profile should be removed: ${NAMES[3]}"
    echo $PROFILES
    exit 1
fi

#make sure that we're prevented from editing an inactive profile
tracko tx invoicer profile-edit --from ${TESTKEY[3]} --cur USD --amount 1mycoin &> /dev/null
CUR=$(tracko query profile ${NAMES[3]} | jq .AcceptedCur | tr -d '"')
if [ "$CUR" == "USD" ]; then 
    echo "Error inactive profile should not be editable"
    exit 1
fi

echo "editing an existing active profile"
tracko tx invoicer profile-edit --from ${TESTKEY[0]} --cur USD --amount 1mycoin > /dev/null
CUR=$(tracko query profile ${NAMES[0]} | jq .AcceptedCur | tr -d '"')
if [ "$CUR" != "USD" ]; then 
    printf 'Error active profile should be editable: want USD, have %s\n' $CUR
    exit 1
fi


echo "sending a contract invoice"
tracko tx invoicer contract-open 1000.99USD --date 2017-01-01 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null

echo "Here is the open invoice:"
ID=$(tracko query invoices | jq .[0][1].ID | tr -d '"')
tracko query invoice 0x$ID

echo "editing the already open invoice"
tracko tx invoicer contract-edit 1000.99CAD --id 0x$ID --date 2017-01-01 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin --debug
echo "Here is the edited invoice:"
tracko query invoice 0x$ID


echo "query all invoices"
tracko query invoices | jq

echo "pay the opened invoice with some cash!"
tracko tx invoicer payment Bucky --ids 0x$ID --paid 0.5BTC --date 2017-01-01 --tx-id "FOOBTC-TX-01" --from ${TESTKEY[0]} --amount 1mycoin 
tracko query invoice 0x$ID | jq
tracko tx invoicer payment Bucky --ids 0x$ID --paid 0.2454003323983133BTC --date 2017-01-01 --tx-id "FOOBTC-TX-02" --from ${TESTKEY[0]} --amount 1mycoin --debug 
tracko query invoice 0x$ID | jq #TODO test if open or not right here
tracko query payments | jq
 
echo "open a receipt"
DIR1=(/tmp/tracko)
DIR2=($DIR1/retrieved)
mkdir $DIR1 ; mkdir $DIR2
wget -O $DIR1/invoicerDoc.png https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png
tracko tx invoicer expense-open 99.99USD --date 2017-01-01 --receipt $DIR1/invoicerDoc.png --taxes 3.00USD --to AllInBits --notes transportation --from ${TESTKEY[1]} --amount 1mycoin --debug > /dev/null

echo "Download the receipt"
ID2=$(tracko query invoices | jq .[1][1].ID | tr -d '"')
echo "tracko query invoice 0x$ID2 --download-expense $DIR2"
tracko query invoice 0x$ID2 --download-expense $DIR2 > /dev/null

if [ ! -f $DIR2/invoicerDoc.png ]; then
    echo "ERROR: receipt didn't download from query"
fi

echo "opening a bunch of invoices for various dates and attempting to get the sum of amount owing"
tracko tx invoicer contract-open 1000USD --date 2017-01-02 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null
tracko tx invoicer contract-open 1000USD --date 2017-01-15 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null
tracko tx invoicer contract-open 1000USD --date 2017-02-01 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null
tracko tx invoicer contract-open 1000USD --date 2017-03-15 --to AllInBits --notes thanks! --from ${TESTKEY[1]} --amount 1mycoin > /dev/null

echo "invoice list"
echo "tracko query invoices | jq"
tracko query invoices | jq
echo "tracko query invoices --date-range 2017-01-02: | jq"
tracko query invoices --date-range 2017-01-02: | jq

echo "sum of invoices due"
tracko query invoices --sum --date-range 2017-01-02: | jq

echo "pay a bit of the invoices off"
tracko tx invoicer payment Bucky --date-range 2017-01-02: --paid 2BTC --date 2017-03-15 --tx-id "FOOBTC-TX-03" --from ${TESTKEY[0]} --amount 1mycoin --debug
tracko query invoices --date-range 2017-01-02: | jq
tracko query invoices --sum --date-range 2017-01-02: | jq

echo "pay off the rest of the invoices"
REMAINDER=$(tracko query invoices --sum | jq .SumDue.Amount)
echo $REMAINDER

