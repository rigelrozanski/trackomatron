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
First generate a couple demonstration keys using the light-client:
```
trackocli keys new bobby
trackocli keys new foobar
```

Open a second window, perform the initialization commands:

```
tracko init
tracko unsafe_reset_all
```

Modify the genesis file in `~/.tracko/genesis.json` to have the address and public key of 
the `bobby` key you created in the first window. This information can be accessed with 
`trackcli keys get bobby --output=json`. Finally in this second window start the full node
server with `tracko start`.

Great! Now we can use commands from the light client to send invoices between bobby and foobar 

...to be continued

### Commands

Transaction commands can be executed from either a full operation node
(heavy-client) or from a light-client brethren. 

| Client | Query | Transaction |
|-----|-----|-----|
| heavy  | `tracko query [command]` | `tracko tx invoicer [command]` |
| light  | `trackocli proof state [command]` | `trackocli tx [command]` |

The `--help` flag can be used from any command to list available flags/args and
their usage 

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

### Testing
Comprehensive testing is performed in bash scripts found in `test/` check them
out!  These files can give you a pretty good idea of to used some of the nuance
capabilities of trackomatron. In addition the golang test suite is utilized
throughout for unit testing, this testing is presented in the "\*\_test.go"
files throughout the repository.

### Future Development
 - Store dumb-contracts on the blockchain
 - Reference documents to be stored in IPFS
 - State encryption to prevent external query of sensitive information 
