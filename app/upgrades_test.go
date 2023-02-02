package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestIncreaseMaxCommissions(t *testing.T) {
	validators := make([]*tmtypes.Validator, 5)
	for i := range validators {
		key := mock.NewPV()
		pubKey, err := key.GetPubKey()
		require.NoError(t, err, "[%d] GetPubKey", i)
		validators[i] = tmtypes.NewValidator(pubKey, 1)
	}
	valSet := tmtypes.NewValidatorSet(validators)

	iniBal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000)))
	accs := make([]authtypes.GenesisAccount, 1)
	balances := make([]banktypes.Balance, len(accs))
	for i := range accs {
		key := secp256k1.GenPrivKey()
		accs[i] = authtypes.NewBaseAccount(key.PubKey().Address().Bytes(), key.PubKey(), uint64(i), 0)
		balances[i] = banktypes.Balance{
			Address: accs[i].GetAddress().String(),
			Coins:   iniBal,
		}
	}

	app := SetupWithGenesisValSet(t, "", valSet, accs, balances...)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	// make sure at least one doesn't have the new value so that there's something that's actually being tested.
	expectedMaxRate := sdk.OneDec()
	canTest := false
	for i, validator := range app.StakingKeeper.GetAllValidators(ctx) {
		t.Logf("Before: validator[%d] has Commission.MaxRate = %s", i, validator.Commission.MaxRate.String())
		if !expectedMaxRate.Equal(validator.Commission.MaxRate) {
			canTest = true
		}
	}
	require.True(t, canTest, "all validators already have Commission.MaxRate = %s", expectedMaxRate.String())

	IncreaseMaxCommissions(ctx, app)

	for i, validator := range app.StakingKeeper.GetAllValidators(ctx) {
		t.Logf("After: validator[%d] has Commission.MaxRate = %s", i, validator.Commission.MaxRate.String())
		assert.Equal(t, expectedMaxRate, validator.Commission.MaxRate, "validator[%d].Commission.MaxRate", i)
	}
}
