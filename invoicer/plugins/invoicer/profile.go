package invoicer

import (
	abci "github.com/tendermint/abci/types"
	btypes "github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin-examples/invoicer/types"
)

func validateProfile(profile *types.Profile) abci.Result {
	switch {
	case len(profile.Name) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have a name")
	case len(profile.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have an accepted currency")
	case profile.DueDurationDays < 0:
		return abci.ErrInternalError.AppendLog("new profile due duration must be non-negative")
	default:
		return abci.OK
	}
}

func writeProfile(store btypes.KVStore, active []string, profile *types.Profile) {

	store.Set(ProfileKey(profile.Name), wire.BinaryBytes(*profile))

	//also add it to the list of open profiles
	active = append(active, profile.Name)
	store.Set(ListProfileKey(), wire.BinaryBytes(active))
}

func removeProfile(store btypes.KVStore, active []string, profile *types.Profile) {

	//TODO remove profile, can't delete store entry on current KVstore implementation
	store.Set(ProfileKey(profile.Name), nil)

	//remove from the active profile list
	for i, v := range active {
		if v == profile.Name {
			active = append(active[:i], active[i+1:]...)
			break
		}
	}
	store.Set(ListProfileKey(), wire.BinaryBytes(active))
}

//TODO remove this once replaced KVStore functionality
func profileIsActive(active []string, name string) bool {
	for _, p := range active {
		if p == name {
			return true
		}
	}
	return false
}

func runTxProfile(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte, shouldExist bool,
	action func(store btypes.KVStore, active []string, profile *types.Profile)) (res abci.Result) {

	// Decode tx
	var profile = new(types.Profile)
	err := wire.ReadBinaryBytes(txBytes, profile)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	//Check existence
	active, err := getListProfile(store)
	if err != nil {
		return abciErrGetProfiles
	}
	if shouldExist && !profileIsActive(active, profile.Name) {
		return abciErrProfileNonExistent
	}
	if !shouldExist && profileIsActive(active, profile.Name) {
		return abciErrProfileExists
	}

	//Validate Tx
	res = validateProfile(profile)
	if res.IsErr() {
		return res
	}

	action(store, active, profile)
	return abci.OK
}
