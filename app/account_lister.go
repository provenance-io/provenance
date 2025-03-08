package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	"github.com/provenance-io/provenance/internal/provutils"
	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type AccountLister struct {
	App       *App
	BondDenom string
}

// ListAccounts creates a JSON file with information about all the current accounts.
// To make the binary create an accounts file, call ListAccounts in the app.PreBlocker method.
// You'll want to make a global bool to indicate whether it's been written yet, and only do it
// the one time, otherwise you'll dump on every block, and a dump is about 35M.
// The file will end up in your current working directory when you start the node.
func ListAccounts(ctx sdk.Context, app *App) {
	al := AccountLister{App: app}
	var err error
	al.BondDenom, err = app.StakingKeeper.BondDenom(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error getting bond denom: %v", err))
		return
	}
	al.CreateAccountsFile(ctx)
}

// CreateAccountsFile looks up all the account info and creates a JSON file with all the data.
func (al *AccountLister) CreateAccountsFile(ctx sdk.Context) {
	filename := fmt.Sprintf("./accounts_%d_%s.json", ctx.BlockHeight(), ctx.BlockTime().Format("2006-01-02_03-04-05"))
	ctx.Logger().Info(fmt.Sprintf("Writing account info to %q.", filename))

	file, err := os.Create(filename)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error creating file %q: %v", filename, err))
		return
	}

	var buf *bufio.Writer
	defer func() {
		if buf != nil && buf.Buffered() > 0 {
			ctx.Logger().Debug(fmt.Sprintf("Flushing final buffer with %d bytes.", buf.Buffered()))
			if err = buf.Flush(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error flushing final buffer %q: %v.", filename, err))
			}
		}
		buf = nil

		if file != nil {
			if err = file.Sync(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error syncing file %q: %v.", filename, err))
			}
			if err = file.Close(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error closing file %q: %v.", filename, err))
			}
			file = nil
		}
		ctx.Logger().Info(fmt.Sprintf("Done writing account info to %q.", filename))
	}()

	buf = bufio.NewWriterSize(file, 32768)
	ctx.Logger().Debug(fmt.Sprintf("Buffer size: %d", buf.Size()))

	totalAccounts := 0
	totalWasmAccounts := 0
	totalNhash := sdkmath.ZeroInt()
	totalUnvested := sdk.Coins{}
	totalDelegated := sdk.DecCoins{}
	totalRewards := sdk.DecCoins{}
	totalOnHold := sdkmath.ZeroInt()

	first := true
	err = al.App.AccountKeeper.Accounts.Walk(ctx, nil, func(_ sdk.AccAddress, acct sdk.AccountI) (stop bool, err error) {
		entry, err := al.NewAccountEntry(ctx, acct)
		if err != nil {
			return true, err
		}
		totalAccounts++
		totalNhash = totalNhash.Add(entry.Nhash)
		totalUnvested = totalUnvested.Add(entry.Unvested...)
		totalDelegated = totalDelegated.Add(entry.Delegated...)
		totalRewards = totalRewards.Add(entry.UnclaimedRewards...)
		totalOnHold = totalOnHold.Add(entry.OnHold)
		if entry.IsWASM {
			totalWasmAccounts++
		}

		data, err := al.AsJSON(entry)
		if err != nil {
			return true, err
		}

		lead := provutils.Ternary(first, "[\n  ", ",\n  ")
		_, err = buf.WriteString(lead)
		if err != nil {
			return true, fmt.Errorf("could not write line lead %q to buffer: %w", lead, err)
		}

		if buf.Available() < len(data) {
			ctx.Logger().Debug(fmt.Sprintf("Flushing buffer with %d bytes.", buf.Buffered()))
			if err = buf.Flush(); err != nil {
				return true, fmt.Errorf("could not flush buffer: %w", err)
			}
		}

		_, err = buf.Write(data)
		if err != nil {
			return true, fmt.Errorf("could not write account %d %q to buffer: %w", acct.GetAccountNumber(), acct.GetAddress().String(), err)
		}
		first = false
		return false, nil
	})
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error walking accounts: %v", err))
	}

	_, err = buf.WriteString("\n]\n")
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error writing ending: %v", err))
	}

	ctx.Logger().Info(fmt.Sprintf("Number of accounts  ---------------> %27s", al.AddCommas(strconv.Itoa(totalAccounts))))
	ctx.Logger().Info(fmt.Sprintf("Number of WASM accounts  ----------> %27s", al.AddCommas(strconv.Itoa(totalWasmAccounts))))
	ctx.Logger().Info(fmt.Sprintf("Total nhash in accounts  ----------> %27s", al.AddCommas(totalNhash.String())))
	ctx.Logger().Info(fmt.Sprintf("Total unvested in accounts  -------> %27s", al.AddCommas(totalUnvested.String())))
	ctx.Logger().Info(fmt.Sprintf("Total delegated from accounts  ----> %49s", al.AddCommas(totalDelegated.String())))
	ctx.Logger().Info(fmt.Sprintf("Total rewards for accounts  -------> %49s", al.AddCommas(totalRewards.String())))
	ctx.Logger().Info(fmt.Sprintf("Total nhash on hold in accounts  --> %27s", al.AddCommas(totalOnHold.String())))
}

// AccountEntry holds all the info that we want about an account.
type AccountEntry struct {
	Addr             sdk.AccAddress
	Type             string
	Name             string
	Nhash            sdkmath.Int
	VestingStart     int64
	VestingEnd       int64
	Unvested         sdk.Coins
	Delegated        sdk.DecCoins
	UnclaimedRewards sdk.DecCoins
	OnHold           sdkmath.Int
	IsWASM           bool
	AccountI         sdk.AccountI
}

// NewAccountEntry looks up all the extra info about an account, creates an AccountEntry, and coverts it to JSON.
func (al *AccountLister) NewAccountEntry(ctx sdk.Context, acctI sdk.AccountI) (*AccountEntry, error) {
	addr := acctI.GetAddress()
	entry := &AccountEntry{Addr: addr, AccountI: acctI}

	switch acct := acctI.(type) {
	case *authtypes.BaseAccount:
		entry.Type = "Base"
		// Nothing extra to do here.
	case *authtypes.ModuleAccount:
		entry.Type = "Module"
		entry.Name = acct.Name
	case *vesting.BaseVestingAccount:
		entry.Type = "BaseVesting"
		entry.VestingEnd = acct.EndTime
		entry.Unvested = acct.LockedCoinsFromVesting(acct.OriginalVesting)
	case *vesting.ContinuousVestingAccount:
		entry.Type = "ContinuousVesting"
		entry.VestingStart = acct.StartTime
		entry.VestingEnd = acct.EndTime
		entry.Unvested = acct.LockedCoins(ctx.BlockTime())
	case *vesting.DelayedVestingAccount:
		entry.Type = "DelayedVesting"
		entry.VestingEnd = acct.EndTime
		entry.Unvested = acct.LockedCoins(ctx.BlockTime())
	case *vesting.PeriodicVestingAccount:
		entry.Type = "PeriodicVesting"
		entry.VestingStart = acct.StartTime
		entry.VestingEnd = acct.EndTime
		entry.Unvested = acct.LockedCoins(ctx.BlockTime())
	case *vesting.PermanentLockedAccount:
		entry.Type = "PermanentLocked"
		entry.VestingEnd = acct.EndTime
		entry.Unvested = acct.LockedCoins(ctx.BlockTime())
	case *exchange.MarketAccount:
		entry.Type = "exchange.Market"
		if len(acct.MarketDetails.Name) > 0 {
			entry.Name = fmt.Sprintf("%s (market_id %d)", acct.MarketDetails.Name, acct.MarketId)
		} else {
			entry.Name = fmt.Sprintf("market_id %d", acct.MarketId)
		}
	case *markertypes.MarkerAccount:
		entry.Type = "Marker"
		entry.Name = acct.Denom
	case *ica.InterchainAccount:
		entry.Type = "Interchain"
		// Nothing extra to do here.
	default:
		return nil, fmt.Errorf("Unknown account type %T", acctI)
	}

	entry.Nhash = al.GetNhash(ctx, addr)

	var err error
	entry.Delegated, entry.UnclaimedRewards, err = al.GetDelegationInfo(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("could not get delgation info for %q: %w", addr.String(), err)
	}

	nhashOnHold, err := al.App.HoldKeeper.GetHoldCoin(ctx, addr, "nhash")
	if err != nil {
		return nil, fmt.Errorf("could not get nhash on hold for %q: %w", addr.String(), err)
	}
	if !nhashOnHold.IsNil() {
		entry.OnHold = nhashOnHold.Amount
	} else {
		entry.OnHold = sdkmath.ZeroInt()
	}

	entry.IsWASM = al.IsWASM(ctx, entry)
	if entry.IsWASM {
		entry.Type += "-wasm"
	}
	return entry, nil
}

// JSONAccountEntry holds all the info that we want about an account. It is marshaled as JSON for the file.
type JSONAccountEntry struct {
	Addr             string          `json:"addr"`
	Type             string          `json:"type"`
	Name             string          `json:"name,omitempty"`
	Nhash            string          `json:"nhash,omitempty"`
	VestingStart     string          `json:"vesting_start,omitempty"`
	VestingEnd       string          `json:"vesting_end,omitempty"`
	Unvested         string          `json:"unvested,omitempty"`
	Delegated        string          `json:"delegated,omitempty"`
	UnclaimedRewards string          `json:"unclaimed_rewards,omitempty"`
	OnHold           string          `json:"on_hold,omitempty"`
	IsWASM           bool            `json:"is_wasm,omitempty"`
	Account          json.RawMessage `json:"account"`
}

// AsJSON converts the provided entry to a json object.
func (al *AccountLister) AsJSON(entry *AccountEntry) ([]byte, error) {
	je := JSONAccountEntry{
		Addr:             entry.Addr.String(),
		Type:             entry.Type,
		Name:             entry.Name,
		Nhash:            al.CleanNhash(entry.Nhash.String()),
		VestingStart:     al.TimeString(entry.VestingStart),
		VestingEnd:       al.TimeString(entry.VestingEnd),
		Unvested:         al.CleanNhash(entry.Unvested.String()),
		Delegated:        al.CleanNhash(entry.Delegated.String()),
		UnclaimedRewards: al.CleanNhash(entry.UnclaimedRewards.String()),
		OnHold:           al.CleanNhash(entry.OnHold.String()),
		IsWASM:           entry.IsWASM,
	}

	// Using the app codec here so it can actually expand the account.
	var err error
	je.Account, err = al.App.appCodec.MarshalJSON(entry.AccountI)
	if err != nil {
		return nil, fmt.Errorf("could not marshal account %d %q: %w", entry.AccountI.GetAccountNumber(), entry.Addr.String(), err)
	}

	// Using the json library here because an AccountEntry isn't a Message.
	rv, err := json.Marshal(je)
	if err != nil {
		return nil, fmt.Errorf("could not marshal entry for %d %q: %w", entry.AccountI.GetAccountNumber(), entry.Addr.String(), err)
	}
	return rv, nil
}

// TimeString returns a string of the provided epoch.
func (al *AccountLister) TimeString(epoch int64) string {
	if epoch == 0 {
		return ""
	}
	return time.Unix(epoch, 0).UTC().Format(time.RFC3339)
}

// GetNhash gets just the amount of nhash in the account.
func (al *AccountLister) GetNhash(ctx sdk.Context, addr sdk.AccAddress) sdkmath.Int {
	coin := al.App.BankKeeper.GetBalance(ctx, addr, "nhash")
	if coin.IsNil() {
		return sdkmath.ZeroInt()
	}
	return coin.Amount
}

// GetAllBalances returns the amount of nhash, other coins, and nfts in an account.
func (al *AccountLister) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, sdk.Coins, int) {
	var nhash sdk.Coin
	var balances sdk.Coins
	nfts := 0
	al.App.BankKeeper.IterateAccountBalances(ctx, addr, func(coin sdk.Coin) bool {
		switch {
		case coin.Denom == "nhash":
			nhash = coin
		case strings.HasPrefix(coin.Denom, metadatatypes.DenomPrefix):
			nfts++
		default:
			// The SDK uses balances = balances.Add(balance) here, but that's crazy slow with lots of coins.
			balances = append(balances, coin)
		}
		return false
	})

	balances.Sort()
	return nhash, balances.Sort(), nfts
}

// GetDelegationInfo returns the amount delegated and the amount of rewards waiting.
func (al *AccountLister) GetDelegationInfo(ctx sdk.Context, addr sdk.AccAddress) (delegated sdk.DecCoins, rewards sdk.DecCoins, err error) {
	// Need to identify the amount delegated and the amount of rewards they've got waiting.
	dels, err := al.App.StakingKeeper.GetAllDelegatorDelegations(ctx, addr)
	if err != nil {
		return nil, nil, err
	}
	if len(dels) == 0 {
		return nil, nil, nil
	}

	// The distribution query needs a state update that we don't want persisted.
	ctx, _ = ctx.CacheContext()

	for _, del := range dels {
		if del.Shares.IsZero() {
			continue
		}

		valAddr := del.GetValidatorAddr()
		var valBz []byte
		valBz, err = al.App.appCodec.InterfaceRegistry().SigningContext().ValidatorAddressCodec().StringToBytes(valAddr)
		if err != nil {
			return nil, nil, fmt.Errorf("could not decode validator %q: %w", valAddr, err)
		}

		var val stakingtypes.Validator
		val, err = al.App.StakingKeeper.GetValidator(ctx, valBz)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get validator %q: %w", valAddr, err)
		}
		delCoin := sdk.NewDecCoinFromDec(al.BondDenom, val.TokensFromShares(del.Shares))
		delegated = delegated.Add(delCoin)

		var endingPeriod uint64
		endingPeriod, err = al.App.DistrKeeper.IncrementValidatorPeriod(ctx, val)
		if err != nil {
			return nil, nil, fmt.Errorf("could not increment validator period for %q: %w", valAddr, err)
		}

		var rewardCoins sdk.DecCoins
		rewardCoins, err = al.App.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get delegation rewards for %q: %w", valAddr, err)
		}
		rewards = rewards.Add(rewardCoins...)
	}

	return delegated, rewards, nil
}

func (al *AccountLister) IsWASM(ctx sdk.Context, entry *AccountEntry) bool {
	// if entry.Type != "Base" || entry.AccountI.GetPubKey() != nil || entry.AccountI.GetSequence() != uint64(0) {
	// 	return false
	// }
	return al.App.WasmKeeper.HasContractInfo(ctx, entry.Addr)
}

// CleanNhash will remove the "nhash" from a coin string that ONLY has nhash in it.
// Otherwise, it returns the provided string.
func (al *AccountLister) CleanNhash(amt string) string {
	amt = strings.TrimLeft(amt, "0")
	if len(amt) == 0 || strings.Contains(amt, ",") || !strings.HasSuffix(amt, "nhash") {
		return amt
	}
	rv := strings.TrimSuffix(amt, "nhash")
	if len(rv) == 0 {
		return ""
	}
	if rv[0] == '.' {
		return "0" + rv
	}
	return rv
}

// AddCommas will add commas to the number string provided. If it has a fractional portion, spaces will be added to that.
// If it ends in "nhash", that will be stripped away. If it has a comma already, nothing will change.
func (al *AccountLister) AddCommas(amt string) string {
	if strings.Contains(amt, ",") {
		return amt
	}
	orig := amt
	amt = strings.TrimSuffix(amt, "nhash")
	if len(amt) == 0 {
		return amt
	}

	parts := strings.Split(amt, ".")
	if len(parts) < 1 || len(parts) > 2 {
		panic(fmt.Errorf("cannot add commas to invalid amount %q", orig))
	}

	lenP0 := len(parts[0])
	lhs := parts[0]
	if len(lhs) > 3 {
		lhz := make([]rune, 0, lenP0+(lenP0-1)/3)
		for i, digit := range lhs {
			if i > 0 && (lenP0-i)%3 == 0 {
				lhz = append(lhz, ',')
			}
			lhz = append(lhz, digit)
		}
		lhs = string(lhz)
	}

	if len(parts) < 2 || len(parts[1]) == 0 {
		return lhs
	}

	lenP1 := len(parts[1])
	rhs := parts[1]
	if len(rhs) > 5 {
		rhz := make([]rune, 0, lenP1+(lenP1-1)/5)
		for i, digit := range rhs {
			if i > 0 && i%5 == 0 {
				rhz = append(rhz, ' ')
			}
			rhz = append(rhz, digit)
		}
		rhs = string(rhz)
	}

	return lhs + "." + rhs
}
