package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MaxEventTagLength is the maximum length that an event tag can have.
// 100 was chosen because that's what we used for the external ids.
const MaxEventTagLength = 100

// Validate returns an error if this Commitment is invalid.
func (c Commitment) Validate() error {
	if _, err := sdk.AccAddressFromBech32(c.Account); err != nil {
		return fmt.Errorf("invalid account %q: %w", c.Account, err)
	}

	if c.MarketId == 0 {
		return errors.New("invalid market id: cannot be zero")
	}

	if err := c.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount %q: %w", c.Amount, err)
	}

	return nil
}

// String returns a string representation of this AccountAmount.
func (a AccountAmount) String() string {
	return fmt.Sprintf("%s:%q", a.Account, a.Amount)
}

// Validate returns an error if this AccountAmount is invalid.
func (a AccountAmount) Validate() error {
	if _, err := sdk.AccAddressFromBech32(a.Account); err != nil {
		return fmt.Errorf("invalid account %q: %w", a.Account, err)
	}
	if err := a.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount %q: %w", a.Amount, err)
	}
	return nil
}

// SumAccountAmounts gets the total of all the amounts in the provided entries.
func SumAccountAmounts(entries []AccountAmount) sdk.Coins {
	var rv sdk.Coins
	for _, entry := range entries {
		rv = rv.Add(entry.Amount...)
	}
	return rv
}

// SimplifyAccountAmounts combines entries with the same account into a single entry.
func SimplifyAccountAmounts(entries []AccountAmount) []AccountAmount {
	if len(entries) <= 1 {
		return entries
	}

	amounts := make(map[string]sdk.Coins)
	addrs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if _, known := amounts[entry.Account]; !known {
			amounts[entry.Account] = nil
			addrs = append(addrs, entry.Account)
		}
		amounts[entry.Account] = amounts[entry.Account].Add(entry.Amount...)
	}

	rv := make([]AccountAmount, len(addrs))
	for i, addr := range addrs {
		rv[i] = AccountAmount{Account: addr, Amount: amounts[addr]}
	}

	return rv
}

// AccountAmountsToBankInputs converts some AccountAmount entries, each to a banktypes.Input.
func AccountAmountsToBankInputs(entries ...AccountAmount) []banktypes.Input {
	rv := make([]banktypes.Input, len(entries))
	for i, entry := range entries {
		rv[i] = banktypes.Input{Address: entry.Account, Coins: entry.Amount}
	}
	return rv
}

// AccountAmountsToBankOutputs converts some AccountAmount entries, each to a banktypes.Output.
func AccountAmountsToBankOutputs(entries ...AccountAmount) []banktypes.Output {
	rv := make([]banktypes.Output, len(entries))
	for i, entry := range entries {
		rv[i] = banktypes.Output{Address: entry.Account, Coins: entry.Amount}
	}
	return rv
}

// String returns a string representation of this MarketAmount.
func (m MarketAmount) String() string {
	return fmt.Sprintf("%d:%q", m.MarketId, m.Amount)
}

// String returns a string representation of this NetAssetPrice.
func (n NetAssetPrice) String() string {
	return fmt.Sprintf("%q=%q", n.Assets, n.Price)
}

// Validate returns an error if this NetAssetPrice is invalid.
func (n NetAssetPrice) Validate() error {
	if err := n.Assets.Validate(); err != nil {
		return fmt.Errorf("invalid assets %q: %w", n.Assets, err)
	}
	if n.Assets.IsZero() {
		return fmt.Errorf("invalid assets %q: cannot be zero", n.Assets)
	}

	if err := n.Price.Validate(); err != nil {
		return fmt.Errorf("invalid price %q: %w", n.Price, err)
	}
	if n.Price.IsZero() {
		return fmt.Errorf("invalid price %q: cannot be zero", n.Price)
	}

	return nil
}

// ValidateEventTag makes sure an event tag is okay.
func ValidateEventTag(eventTag string) error {
	if len(eventTag) > MaxEventTagLength {
		return fmt.Errorf("invalid event tag %q (length %d): exceeds max length %d",
			eventTag[:5]+"..."+eventTag[len(eventTag)-5:], len(eventTag), MaxEventTagLength)
	}
	return nil
}

// BuildCommitmentTransfers builds all of the transfers needed to process the provided commitment transfers.
func BuildCommitmentTransfers(marketID uint32, inputs, outputs, fees []AccountAmount) ([]*Transfer, error) {
	rv, err := buildPrimaryTransfers(inputs, outputs)
	if err != nil {
		return nil, err
	}

	if ft := buildFeesTransfer(marketID, fees); ft != nil {
		rv = append(rv, ft)
	}

	return rv, nil
}

// accountInt associates an account with an sdkmath.Int.
type accountInt struct {
	account string
	int     sdkmath.Int
}

// denomSourceMap is a map that associates a denom with accounts and amounts of that denom.
type denomSourceMap map[string][]*accountInt

// newDenomSourceMap creates a new denomSourceMap from a set of AccountAmount entries.
func newDenomSourceMap(entries []AccountAmount) denomSourceMap {
	rv := make(denomSourceMap)
	for _, entry := range entries {
		for _, coin := range entry.Amount {
			rv[coin.Denom] = append(rv[coin.Denom], &accountInt{account: entry.Account, int: coin.Amount})
		}
	}
	return rv
}

// sum returns the total Coins in this denomSourceMap.
func (m denomSourceMap) sum() sdk.Coins {
	var rv sdk.Coins
	for denom, entries := range m {
		for _, entry := range entries {
			rv = rv.Add(sdk.Coin{Denom: denom, Amount: entry.int})
		}
	}
	return rv
}

// useFunds finds the coin amount among the denomSourceMap, and removes that amount from the map.
func useFunds(coin sdk.Coin, funds denomSourceMap) ([]AccountAmount, error) {
	var rv []AccountAmount
	amtLeft := coin.Amount
	for amtLeft.IsPositive() && len(funds[coin.Denom]) > 0 {
		if funds[coin.Denom][0].int.LTE(amtLeft) {
			rv = append(rv, AccountAmount{
				Account: funds[coin.Denom][0].account,
				Amount:  sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: funds[coin.Denom][0].int}},
			})
			amtLeft = amtLeft.Sub(funds[coin.Denom][0].int)
			funds[coin.Denom] = funds[coin.Denom][1:]
			continue
		}

		rv = append(rv, AccountAmount{
			Account: funds[coin.Denom][0].account,
			Amount:  sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: amtLeft}},
		})
		amtLeft = sdkmath.ZeroInt()
		funds[coin.Denom][0].int = funds[coin.Denom][0].int.Sub(amtLeft)
	}

	if len(funds[coin.Denom]) == 0 {
		delete(funds, coin.Denom)
	}

	if !amtLeft.IsZero() {
		return nil, fmt.Errorf("failed to allocate %s to outputs: %s left over", coin, amtLeft)
	}
	return rv, nil
}

// buildPrimaryTransfers builds the transfers for a set of inputs and outputs.
func buildPrimaryTransfers(inputs, outputs []AccountAmount) ([]*Transfer, error) {
	if len(inputs) == 1 || len(outputs) == 1 {
		rv := make([]*Transfer, 1, 2)
		rv[0] = &Transfer{
			Inputs:  AccountAmountsToBankInputs(inputs...),
			Outputs: AccountAmountsToBankOutputs(outputs...),
		}
		return rv, nil
	}

	splitInputs := len(inputs) > len(outputs)
	mainEntries, splittableEntries := inputs, outputs
	if splitInputs {
		mainEntries, splittableEntries = outputs, inputs
	}

	funds := newDenomSourceMap(splittableEntries)
	rv := make([]*Transfer, 0, len(mainEntries)+1)
	for _, entry := range mainEntries {
		var trSplits []AccountAmount
		for _, coin := range entry.Amount {
			coinSplits, err := useFunds(coin, funds)
			if err != nil {
				return nil, err
			}
			trSplits = append(trSplits, coinSplits...)
		}

		if !splitInputs {
			rv = append(rv, &Transfer{
				Inputs:  AccountAmountsToBankInputs(entry),
				Outputs: AccountAmountsToBankOutputs(SimplifyAccountAmounts(trSplits)...),
			})
		} else {
			rv = append(rv, &Transfer{
				Inputs:  AccountAmountsToBankInputs(SimplifyAccountAmounts(trSplits)...),
				Outputs: AccountAmountsToBankOutputs(entry),
			})
		}
	}

	unallocated := funds.sum()
	if !unallocated.IsZero() {
		source := "outputs"
		if splitInputs {
			source = "inputs"
		}
		return nil, fmt.Errorf("%s are left with %s in unallocated funds", source, unallocated)
	}

	return rv, nil
}

// buildFeesTransfer builds the transfer needed to move the provided fees to the market account.
func buildFeesTransfer(marketID uint32, fees []AccountAmount) *Transfer {
	if len(fees) == 0 {
		return nil
	}

	output := AccountAmount{
		Account: GetMarketAddress(marketID).String(),
		Amount:  SumAccountAmounts(fees),
	}
	return &Transfer{
		Inputs:  AccountAmountsToBankInputs(fees...),
		Outputs: AccountAmountsToBankOutputs(output),
	}
}
