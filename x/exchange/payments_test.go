package exchange

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

var ValidPayment = Payment{
	Source:       sdk.AccAddress("Source______________").String(),
	SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("strawberry", 7)),
	Target:       sdk.AccAddress("Target______________").String(),
	TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("tangerine", 5)),
	ExternalId:   "41D83560-8AC7-43FE-9B74-4D2BF090CB92",
}

func TestPayment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payment Payment
		expErr  []string
	}{
		{
			name:    "valid payment",
			payment: ValidPayment,
			expErr:  nil,
		},
		{
			name: "no source",
			payment: Payment{
				Source:       "",
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"invalid source \"\": empty address string is not allowed"},
		},
		{
			name: "invalid source",
			payment: Payment{
				Source:       "notgonnahappen",
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"invalid source \"notgonnahappen\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "no target",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: nil,
		},
		{
			name: "invalid target",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       "notgoodeither",
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"invalid target \"notgoodeither\": decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "no source amount",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: nil,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: nil,
		},
		{
			name: "invalid source amount",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: sdk.Coins{sdk.Coin{Denom: "strawberry", Amount: sdkmath.NewInt(-1)}},
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"invalid source amount \"-1strawberry\": coin -1strawberry amount is not positive"},
		},
		{
			name: "no target amount",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: nil,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: nil,
		},
		{
			name: "invalid target amount",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: sdk.Coins{sdk.Coin{Denom: "tangerine", Amount: sdkmath.NewInt(0)}},
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"invalid target amount \"0tangerine\": coin 0tangerine amount is not positive"},
		},
		{
			name: "no amounts",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: nil,
				Target:       ValidPayment.Target,
				TargetAmount: nil,
				ExternalId:   ValidPayment.ExternalId,
			},
			expErr: []string{"source amount and target amount cannot both be zero"},
		},
		{
			name: "no external id",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   "",
			},
			expErr: nil,
		},
		{
			name: "invalid external id",
			payment: Payment{
				Source:       ValidPayment.Source,
				SourceAmount: ValidPayment.SourceAmount,
				Target:       ValidPayment.Target,
				TargetAmount: ValidPayment.TargetAmount,
				ExternalId:   "p" + strings.Repeat("i", MaxExternalIDLength) + "o",
			},
			expErr: []string{fmt.Sprintf("invalid external id %q (length %d): max length %d",
				"piiii...iiiio", MaxExternalIDLength+2, MaxExternalIDLength)},
		},
		{
			name: "multiple errors",
			payment: Payment{
				Source:       "",
				SourceAmount: sdk.Coins{sdk.Coin{Denom: "strawberry", Amount: sdkmath.NewInt(-1)}},
				Target:       "notgoodeither",
				TargetAmount: sdk.Coins{sdk.Coin{Denom: "tangerine", Amount: sdkmath.NewInt(0)}},
				ExternalId:   "p" + strings.Repeat("i", MaxExternalIDLength) + "o",
			},
			expErr: []string{
				"invalid source \"\": empty address string is not allowed",
				"invalid source amount \"-1strawberry\": coin -1strawberry amount is not positive",
				"invalid target \"notgoodeither\": decoding bech32 failed: invalid separator index -1",
				"invalid target amount \"0tangerine\": coin 0tangerine amount is not positive",
				fmt.Sprintf("invalid external id %q (length %d): max length %d",
					"piiii...iiiio", MaxExternalIDLength+2, MaxExternalIDLength),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.payment.Validate()
			}
			require.NotPanics(t, testFunc, "Validate()")
			assertions.AssertErrorContents(t, err, tc.expErr, "Validate() error")
		})
	}
}

func TestPayment_String(t *testing.T) {
	tests := []struct {
		name    string
		payment Payment
		exp     string
	}{
		{
			name:    "zero value",
			payment: Payment{},
			exp:     "?+\"\"-x-?",
		},
		{
			name: "with all fields",
			payment: Payment{
				Source:       "sam",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("apple", 5), sdk.NewInt64Coin("banana", 99)),
				Target:       "taylor",
				TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("pear", 12), sdk.NewInt64Coin("raisin", 34)),
				ExternalId:   "abc123",
			},
			exp: "sam+\"abc123\":5apple,99banana<->taylor:12pear,34raisin",
		},
		{
			name: "no target",
			payment: Payment{
				Source:       "shannon",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 43)),
				Target:       "",
				TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("date", 65)),
				ExternalId:   "a1b2",
			},
			exp: "shannon+\"a1b2\":43cherry<->?:65date",
		},
		{
			name: "no target amount",
			payment: Payment{
				Source:       "shae",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("blueberry", 51)),
				Target:       "tanner",
				TargetAmount: nil,
				ExternalId:   "myid",
			},
			exp: "shae+\"myid\":51blueberry-->tanner",
		},
		{
			name: "no target or target amount",
			payment: Payment{
				Source:       "spencer",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("eggplant", 1)),
				Target:       "",
				TargetAmount: nil,
				ExternalId:   "abbadaba",
			},
			exp: "spencer+\"abbadaba\":1eggplant-->?",
		},
		{
			name: "no source amount",
			payment: Payment{
				Source:       "sydney",
				SourceAmount: nil,
				Target:       "terry",
				TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("apricot", 38)),
				ExternalId:   "gimmiegimmie",
			},
			exp: "sydney+\"gimmiegimmie\"<--terry:38apricot",
		},
		{
			name: "no external id",
			payment: Payment{
				Source:       "stevie",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("strawberry", 5)),
				Target:       "tatum",
				TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("tangerine", 19)),
				ExternalId:   "",
			},
			exp: "stevie+\"\":5strawberry<->tatum:19tangerine",
		},
		{
			name: "external id with control chars",
			payment: Payment{
				Source:       "sawyer",
				SourceAmount: sdk.NewCoins(sdk.NewInt64Coin("starfruit", 6)),
				Target:       "tobin",
				TargetAmount: sdk.NewCoins(sdk.NewInt64Coin("tomato", 1)),
				ExternalId:   string([]byte{0, 'x', 'y', '\a', 'z', '\r', 254, ' '}),
			},
			exp: "sawyer+\"\\x00xy\\az\\r\\xfe \":6starfruit<->tobin:1tomato",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.payment.String()
			}
			require.NotPanics(t, testFunc, "Payment.String()")
			assert.Equal(t, tc.exp, act, "Payment.String() result")
		})
	}
}
