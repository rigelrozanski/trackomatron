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
    
    newKey craig > /dev/null
    GENKEY=$(trackocli keys get craig -o json | jq .pubkey.data)

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

openProfile(){
    if [ -z "$3" ]; then 
        return 
    fi 
    expect <<- DONE
      spawn trackocli tx profile-open --profile-name $1 --name $2 \
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
    openProfile satoshi craig 1

    #read the profile just written
    echo "using state get"
    trackocli proof state get --app=profile --key=satoshi
    echo "using state profile"
    trackocli proof state profile satoshi --trace
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
