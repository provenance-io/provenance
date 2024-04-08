package queries

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetModuleAccountByName will execute a gprc query to get a module account by name.
func GetModuleAccountByName(val *network.Validator, moduleName string) (sdk.ModuleAccountI, error) {
	url := fmt.Sprintf("%s/cosmos/auth/v1beta1/module_accounts/%s", val.APIAddress, moduleName)
	resp, err := GetRequest(val, url, &authtypes.QueryModuleAccountByNameResponse{})
	if err != nil {
		return nil, err
	}
	if resp.Account == nil {
		return nil, fmt.Errorf("no account found for %q module", moduleName)
	}

	acct := resp.Account.GetCachedValue()
	rv, ok := acct.(sdk.ModuleAccountI)
	if !ok {
		return nil, fmt.Errorf("failed to cast response account %T as %T", acct, sdk.ModuleAccountI(nil))
	}

	return rv, nil
}
