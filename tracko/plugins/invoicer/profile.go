package invoicer

import (
	abci "github.com/tendermint/abci/types"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	btypes "github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin-examples/tracko/types"
)

func validateProfile(profile *types.Profile) abci.Result {
	switch {
	case len(profile.Name) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have a name")
	case len(profile.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have an accepted currency")
	case profile.DueDurationDays < 0:
		return abci.ErrInternalError.AppendLog("new profile due duration must be non-negative")
	case !profile.Active:
		return abciErrProfileInactive
	default:
		return abci.OK
	}
}

func writeProfile(store btypes.KVStore, active []string, profile *types.Profile) abci.Result {

	//Validate Tx
	res := validateProfile(profile)
	if res.IsErr() {
		return res
	}

	//write the profile to the profile key
	store.Set(ProfileKey(profile.Name), wire.BinaryBytes(*profile))

	//add the profile name to the list of active profiles
	active = append(active, profile.Name)
	store.Set(ListProfileActiveKey(), wire.BinaryBytes(active))

	return abci.OK
}

func deactivateProfile(store btypes.KVStore, active []string, profile *types.Profile) abci.Result {

	name := profile.Name

	//get the original profile that's saved from the store, set that one to inactive
	storeProfile, err := getProfile(store, name)
	if err != nil {
		return abciErrNoProfile
	}

	storeProfile.Active = false
	store.Set(ProfileKey(name), wire.BinaryBytes(storeProfile))

	//remove profile from the list of active profiles
	active = removeElemStringArray(active, name)
	store.Set(ListProfileActiveKey(), wire.BinaryBytes(active))

	//Add the profile name to the list of inactive profiles
	all, err := getListString(store, ListProfileActiveKey())
	if err != nil {
		return abciErrGetAllProfiles
	}
	all = append(all, name)
	store.Set(ListProfileInactiveKey(), wire.BinaryBytes(all))

	return abci.OK
}

//TODO move to tmlibs/common
func removeElemStringArray(a []string, remove string) []string {
	for i, el := range a {
		if el == remove {
			a = append(a[:i], a[i+1:]...)
		}
	}
	return a
}

//TODO remove this once replaced KVStore functionality
func profileRegistered(active []string, name string) bool {
	for _, p := range active {
		if p == name {
			return true
		}
	}
	return false
}

func nameFromAddress(store btypes.KVStore, active []string, address bcmd.Address) string {
	for _, name := range active {
		profile, _ := getProfile(store, name)
		if profile.Address == address {
			return profile.Name
		}
	}
	return ""
}

func runTxProfile(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte, shouldExist bool,
	action func(store btypes.KVStore, active []string, profile *types.Profile) abci.Result) abci.Result {

	// Decode tx
	var profile = new(types.Profile)
	err := wire.ReadBinaryBytes(txBytes, profile)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	//get the name from address, if not opening a new profile
	active, err := getListString(store, ListProfileActiveKey())
	if err != nil {
		return abciErrGetProfiles
	}
	if len(profile.Name) == 0 {
		profile.Name = nameFromAddress(store, active, profile.Address)
	}

	//Check existence
	if shouldExist && !profileRegistered(active, profile.Name) {
		return abciErrProfileNonExistent
	}
	if !shouldExist && profileRegistered(active, profile.Name) {
		return abciErrProfileExists
	}

	return action(store, active, profile)
}
