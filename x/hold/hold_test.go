package hold

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestAccountHold_Validate(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	accountHold := func(addr string, amount sdk.Coins) AccountHold {
		return AccountHold{
			Address: addr,
			Amount:  amount,
		}
	}

	addr := sdk.AccAddress("control_addr________").String()

	tests := []struct {
		name string
		ae   AccountHold
		exp  string
	}{
		{
			name: "control",
			ae:   accountHold(addr, coins("1000nhash")),
			exp:  "",
		},
		{
			name: "control with two coins",
			ae:   accountHold(addr, coins("50atom,1000nhash")),
			exp:  "",
		},
		{
			name: "no address",
			ae:   accountHold("", coins("1000nhash")),
			exp:  "invalid address: empty address string is not allowed",
		},
		{
			name: "invalid address",
			ae:   accountHold("bad", coins("1000nhash")),
			exp:  "invalid address: decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name: "invalid amount",
			ae:   accountHold(addr, sdk.Coins{coin(-50, "badcoin")}),
			exp:  "invalid amount: coin -50badcoin amount is not positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ae.Validate()
			assertions.AssertErrorValue(t, err, tc.exp, "Validate()")
		})
	}
}
