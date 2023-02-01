package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestIncreaseMaxCommissions(t *testing.T) {
	iniBal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000)))
	validatorAccs := make([]authtypes.GenesisAccount, 5)
	balances := make([]banktypes.Balance, len(validatorAccs))
	for i := range validatorAccs {
		privKey := secp256k1.GenPrivKey()
		validatorAccs[i] = authtypes.NewBaseAccount(privKey.PubKey().Address().Bytes(), privKey.PubKey(), uint64(i), 0)
		balances[i] = banktypes.Balance{
			Address: validatorAccs[i].GetAddress().String(),
			Coins:   iniBal,
		}
	}

	app := SetupWithGenesisAccounts(t, "", validatorAccs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// make sure at least one doesn't have the new value so that there's something that's actually being tested.
	expectedMaxRate := sdk.OneDec()
	canTest := false
	for _, validator := range app.StakingKeeper.GetAllValidators(ctx) {
		if !expectedMaxRate.Equal(validator.Commission.MaxRate) {
			canTest = true
			break
		}
	}
	require.True(t, canTest, "all validators already have Commission.MaxRate = %s", expectedMaxRate.String())

	IncreaseMaxCommissions(ctx, app)

	for i, validator := range app.StakingKeeper.GetAllValidators(ctx) {
		assert.Equal(t, expectedMaxRate, validator.Commission.MaxRate, "validator[%d].Commission.MaxRate", i)
	}
}
