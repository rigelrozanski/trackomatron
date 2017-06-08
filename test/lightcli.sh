#! /bin/bash

oneTimeSetUp(){
    # start tendermint
    TR=~/.tracko
    rm -rf $TR
    tracko init
    tracko unsafe_reset_all

    LC=~/.trackolightcli
    rm -rf $LC
    export BCHOME=$LC
    
    ACCOUNTS=(jae ethan craig rigel)
    for i in "${!ACCOUNTS[@]}"; do 
        newKey ${ACCOUNTS[$i]}  > /dev/null
    done
    
    GENKEY=$(trackocli keys get ${ACCOUNTS[0]} -o json | jq .pubkey.data)

    #change the genesis to satoshi
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
    rm -rf /tmp/tracko/
    kill -9 $pid_basecoin
    echo "cleaning up bash test"
}

newKey(){
    if [ -z "$1" ]; then 
        return 
    fi 
    expect <<- DONE
      spawn trackocli keys new $1
      expect "Enter a passphrase:" 
      send -- "passweirdo\r"
      expect "Repeat the passphrase:"
      send -- "passweirdo\r"
      expect eof
DONE
}

initLightCli(){
    expect <<- DONE
      spawn basecli init --chainid test_chain_id --node tcp://localhost:46657
      expect "Is this valid (y/n)?" 
      send -- "y\r"
      expect eof
DONE
}


tx100(){
    if [ -z "$3" ]; then 
        return 
    fi 
    expect <<- DONE
      spawn trackocli tx $1 --name $2 \
          --amount 100mycoin --fee 0mycoin --sequence $3
      expect "Please enter passphrase for $2:" 
      send -- "passweirdo\r"
      expect eof
DONE
}

tx(){
    if [ -z "$3" ]; then 
        return 
    fi 
    expect <<- DONE
      spawn trackocli tx $1 --name $2 \
          --amount 1mycoin --fee 0mycoin --sequence $3
      expect "Please enter passphrase for $2:" 
      send -- "passweirdo\r"
      expect eof
DONE
}

#testSetupKeys(){
#}

testOpeningProfiles(){
    #open the profile of satoshi with craig key and sequence of 1
    #tx "profile-open --profile-name satoshi" "jae" 1

    #read the profile just written
    #echo "using state get"
    #trackocli proof state get --app=profile --key=satoshi
    #echo "using state profile"
    #trackocli proof state profile satoshi --trace
    
    
    NAMES=(AllInBits Bucky Satoshi Dummy)
    SEQ=(1 1 1 1)
    for i in "${!NAMES[@]}"; do 
        #send all the existing accounts some coins 
        ADDR[$i]=$(trackocli keys get ${ACCOUNTS[$i]} --output=json | jq .address | tr -d '"')
        
        if [ "$i" != "0" ]; then
            err=$((tx100 "send --to ${ADDR[$i]}" "${ACCOUNTS[0]}" ${SEQ[0]})  2>&1 > /dev/null)
            assertNull "Error Non-Null Line $LINENO $err" "$err"
            SEQ[0]=$((${SEQ[0]}+1))
        fi 
    
        #open the profile
        err=$((tx "profile-open --profile-name=${NAMES[$i]} --cur BTC" ${ACCOUNTS[$i]} ${SEQ[$i]}) 2>&1 > /dev/null)
        SEQ[$i]=$((${SEQ[$i]}+1))
        assertNull "Error Non-Null Line $LINENO $err" "$err"
    done
    
    #check if the profiles have been opened
    PROFILES=$(trackocli proof state profiles)
    for i in "${!NAMES[@]}"; do 
       assertTrue 'profile not created' "[[ $PROFILES == *"${NAMES[$i]}"* ]]"
    done
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
