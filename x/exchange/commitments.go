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

// ValidateWithOptionalAmount returns an error if this AccountAmount is invalid. The amount is allowed to be empty.
func (a AccountAmount) ValidateWithOptionalAmount() error {
	if _, err := sdk.AccAddressFromBech32(a.Account); err != nil {
		return fmt.Errorf("invalid account %q: %w", a.Account, err)
	}
	if err := a.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount %q: %w", a.Amount, err)
	}
	return nil
}

// Validate returns an error if this AccountAmount is invalid or has a zero amount.
func (a AccountAmount) Validate() error {
	if err := a.ValidateWithOptionalAmount(); err != nil {
		return err
	}
	if a.Amount.IsZero() {
		return fmt.Errorf("invalid amount %q: cannot be zero", a.Amount)
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
// Contract: The inputs, outputs, and fees must be simplified using SimplifyAccountAmounts.
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

// useCoin finds the coin amount among this denomSourceMap, removes that amount from the map,
// and returns AccountAmount entries for the funds used.
func (m denomSourceMap) useCoin(coin sdk.Coin, splittableSource string) ([]AccountAmount, error) {
	var rv []AccountAmount
	amtLeft := coin.Amount
	for amtLeft.IsPositive() && len(m[coin.Denom]) > 0 {
		if m[coin.Denom][0].int.LTE(amtLeft) {
			rv = append(rv, AccountAmount{
				Account: m[coin.Denom][0].account,
				Amount:  sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: m[coin.Denom][0].int}},
			})
			amtLeft = amtLeft.Sub(m[coin.Denom][0].int)
			m[coin.Denom] = m[coin.Denom][1:]
			continue
		}

		rv = append(rv, AccountAmount{
			Account: m[coin.Denom][0].account,
			Amount:  sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: amtLeft}},
		})
		m[coin.Denom][0].int = m[coin.Denom][0].int.Sub(amtLeft)
		amtLeft = sdkmath.ZeroInt()
	}

	if len(m[coin.Denom]) == 0 {
		delete(m, coin.Denom)
	}

	if !amtLeft.IsZero() {
		return nil, fmt.Errorf("failed to allocate %s to %s: %s left over", coin, splittableSource, amtLeft)
	}
	return rv, nil
}

// useCoins calls useCoin on each of the provided coins.
func (m denomSourceMap) useCoins(coins sdk.Coins, splittableSource string) ([]AccountAmount, error) {
	var rv []AccountAmount
	for _, coin := range coins {
		splits, err := m.useCoin(coin, splittableSource)
		if err != nil {
			return nil, err
		}
		rv = append(rv, splits...)
	}
	return SimplifyAccountAmounts(rv), nil
}

// buildPrimaryTransfers builds the transfers for a set of inputs and outputs.
func buildPrimaryTransfers(inputs, outputs []AccountAmount) ([]*Transfer, error) {
	if len(inputs) == 1 || len(outputs) == 1 {
		rv := make([]*Transfer, 1, 2) // 1 extra cap to maybe hold the fees transfer.
		rv[0] = &Transfer{
			Inputs:  AccountAmountsToBankInputs(inputs...),
			Outputs: AccountAmountsToBankOutputs(outputs...),
		}
		return rv, nil
	}

	splitInputs := len(inputs) > len(outputs)
	mainEntries, splittableEntries := inputs, outputs
	mainSource, splittableSource := "inputs", "outputs"
	if splitInputs {
		mainEntries, splittableEntries = outputs, inputs
		mainSource, splittableSource = "outputs", "inputs"
	}

	funds := newDenomSourceMap(splittableEntries)
	rv := make([]*Transfer, 0, len(mainEntries)+1) // 1 extra cap to maybe hold the fees transfer.
	for _, entry := range mainEntries {
		splits, err := funds.useCoins(entry.Amount, splittableSource)
		if err != nil {
			return nil, err
		}

		if !splitInputs {
			rv = append(rv, &Transfer{
				Inputs:  AccountAmountsToBankInputs(entry),
				Outputs: AccountAmountsToBankOutputs(SimplifyAccountAmounts(splits)...),
			})
		} else {
			rv = append(rv, &Transfer{
				Inputs:  AccountAmountsToBankInputs(SimplifyAccountAmounts(splits)...),
				Outputs: AccountAmountsToBankOutputs(entry),
			})
		}
	}

	unallocated := funds.sum()
	if !unallocated.IsZero() {
		return nil, fmt.Errorf("%s are left with %s in unallocated funds", mainSource, unallocated)
	}

	return rv, nil
}

// buildFeesTransfer builds the transfer needed to move the provided fees to the market account.
func buildFeesTransfer(marketID uint32, fees []AccountAmount) *Transfer {
	if len(fees) == 0 {
		return nil
	}

	return &Transfer{
		Inputs: AccountAmountsToBankInputs(fees...),
		Outputs: AccountAmountsToBankOutputs(AccountAmount{
			Account: GetMarketAddress(marketID).String(),
			Amount:  SumAccountAmounts(fees),
		}),
	}
}
