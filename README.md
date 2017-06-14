# Trackomatron

Track invoices on the blockchain!

### Overview
This software is intended to create a space to easily send invoices between and within
institutions. Firstly, the commands of trackmatron are separated into two broad
categories: submitting information to the blockchain (transactions), and
retrieving information from the blockchain (query).  The transaction commands
are separated into three main categories: 
 - Profile management
  - Open/edit/deactivate profiles 
 - Sending Invoices
  - Sending and Editing capabilities
  - Send invoices intended to fulfill contracts
  - Send expense invoices 
 - Paying invoices
  - Bulk payment of invoices from a single receiver

The querying information from the blockchain allows users so sort and retrieve profiles, 
invoices by payments in bulk, additional utility is provided to allow to generate quick totals 
of amount due between parties.  

For a list of commands please see the section below `Commands`.

### Install
```
go get github.com/tendermint/trackomatron
cd $GOPATH/src/github.com/tendermint/trackomatron
make all
```
These commands will generate the binaries in `$GOPATH/bin` named `tracko` and
`trackocli` representing the full and light client respectively.  

### Example

#### Initialize/reset trackomatron, but don't start the chain
```
tracko init
tracko unsafe_reset_all
trackocli init --force-reset
```

#### Set up your trackocli with some keys
```
trackocli keys new bobby
trackocli keys new buddy
```

#### Update genesis so you are rich
```
trackocli keys get bobby -o json
vi ~/.tracko/genesis.json
-> cut/paste your pubkey from the results above
```
or alternatively:  
```
GENKEY=`trackocli keys get bobby -o json | jq .pubkey.data`
GENJSON=`cat ~/.tracko/genesis.json`
echo $GENJSON | jq '.app_options.accounts[0].pub_key.data='$GENKEY > ~/.tracko/genesis.json 
```

#### Start the tracko node
```
tracko start
```

#### In a second terminal connect your trackocli the first time
```
trackocli init --chainid=test_chain_id --node=tcp://localhost:46657
```

#### Send some mycoins over to your pal so they can open a profile
```
ME=`trackocli keys get bobby -o json | jq .address | tr -d '"'`
YOU=`trackocli keys get buddy -o json | jq .address | tr -d '"'`
trackocli tx send --name=bobby --amount=1000mycoin --fee=0mycoin --sequence 1 --to $YOU
```

#### Open up some profiles
```
trackocli tx profile-open --profile-name=b0b --name=bobby --amount=1mycoin --fee=0mycoin --sequence=2
trackocli tx profile-open --profile-name=bud --name=buddy --amount=1mycoin --fee=0mycoin --sequence=1
```


#### Send an invoice to your pal! Then list the open invoices
```
trackocli tx contract-open --invoice-amount=99.99USD --date=2017-01-01 --to=bud --notes=thanks! --name=bobby --amount=1mycoin --fee=0mycoin --sequence=3
trackocli proof state invoices | jq
```

Great! Now you're ready to start using trackomatron to start invoicing all your friends!

#### Close up shop
In the first terminal window hit `ctrl-c`  

### Commands

Transaction commands can be executed from either a full operation node
(heavy-client) or from a light-client brethren. 

| Client | Query | Transaction |
|-----|-----|-----|
| heavy  | `tracko query [command]` | `tracko tx invoicer [command]` |
| light  | `trackocli query app [command]` | `trackocli tx [command]` |

The `--help` flag can be used from any command to list available flags/args and
their full usage. An overview of the commands available are as follows: 

Query
 - invoice     Query an invoice by ID
 - invoices    Query all invoice
 - payment     List historical payment
 - payments    List historical payments
 - profile     Query a profile
 - profiles    List all open profiles

Transaction
 - contract-edit      Edit an open contract invoice to amount <value><currency>
 - contract-open      Send a contract invoice of amount <value><currency>
 - expense-edit       Edit an open expense invoice to amount <value><currency>
 - expense-open       Send an expense invoice of amount <value><currency>
 - payment            pay invoices and expenses with transaction information
 - profile-deactivate Deactivate and existing profile
 - profile-edit       Edit an existing profile
 - profile-open       Open a profile for sending/receiving invoices

One cool flag I will mention to check out is the `--sum` flag used for querying
invoices.  This flag allows you to generate a total of all the invoice amounts
due between two parties.

### Testing
Comprehensive testing is performed in bash scripts found in `test/` check them
out!  These files can give you a pretty good idea of to used some of the nuance
capabilities of trackomatron. Note that these tests require the `expect` package
which can be installed with
```
sudo apt-get update; sudo apt-get install expect
```
In addition the golang test suite is utilized
throughout for unit testing, this testing is presented in the "\*\_test.go"
files throughout the repository.

### Future Development
 - Store dumb-contracts on the blockchain
 - Reference documents to be stored in IPFS
 - State encryption to prevent external query of sensitive information 
