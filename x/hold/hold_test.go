package hold

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAccountHold_Validate(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	c := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ah := func(addr string, amount sdk.Coins) AccountHold {
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
			ae:   ah(addr, cz("1000nhash")),
			exp:  "",
		},
		{
			name: "control with two coins",
			ae:   ah(addr, cz("50atom,1000nhash")),
			exp:  "",
		},
		{
			name: "no address",
			ae:   ah("", cz("1000nhash")),
			exp:  "invalid address: empty address string is not allowed",
		},
		{
			name: "invalid address",
			ae:   ah("bad", cz("1000nhash")),
			exp:  "invalid address: decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name: "invalid amount",
			ae:   ah(addr, sdk.Coins{c(-50, "badcoin")}),
			exp:  "invalid amount: coin -50badcoin amount is not positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ae.Validate()
			if len(tc.exp) > 0 {
				assert.EqualError(t, err, tc.exp, "Validate()")
			} else {
				assert.NoError(t, err, tc.exp, "Validate()")
			}
		})
	}
}
