package commands

import (
	"github.com/tendermint/basecoin-examples/invoicer/plugins/invoicer"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	btypes "github.com/tendermint/basecoin/types"
)

func init() {

	//Register invoicer with basecoin
	bcmd.RegisterStartPlugin(invoicer.Name, func() btypes.Plugin { return invoicer.New() })

	//Change the working directory
	bcmd.DefaultHome = "." + invoicer.Name

	//Change the GenesisJSON
	bcmd.GenesisJSON = `{
  "app_hash": "",
  "chain_id": "test_chain_id",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      }
    }
  ],
  "app_options": {
    "accounts": [{
      "pub_key": {
        "type": "ed25519",
        "data": "619D3678599971ED29C7529DDD4DA537B97129893598A17C82E3AC9A8BA95279"
      },
      "coins": [
        {
          "denom": "mycoin",
          "amount": 9007199254740992
        }
      ]
    }]
  }
}`

}
