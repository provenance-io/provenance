package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AccountData is JSON of a list of JSONAccountEntry.
// Only populated if running with the embed_account_data build tag.
var AccountData []byte

func TestAccountDataTestSuite(t *testing.T) {
	suite.Run(t, new(AccountDataTestSuite))
}

type AccountDataTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context

	logBuffer bytes.Buffer

	startTime time.Time
}

func (s *AccountDataTestSuite) SetupTest() {
	pioconfig.SetProvenanceConfig("nhash", 1905)
	SetConfig(true, false)
	// Bah. maybe it'd be easier to just grab a make-run export, import it, then add all the stuff and export it.
	defer SetLoggerMaker(SetLoggerMaker(BufferedInfoLoggerMaker(&s.logBuffer)))
	s.app = Setup(s.T())
	s.logBuffer.Reset()
	s.startTime = time.Now()
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.startTime})
}

// loadAccountData will add all the entries in the AccountData file to current state.
func loadAccountData(t *testing.T, ctx sdk.Context, app *App) {
	t.Helper()
	if len(AccountData) == 0 {
		t.Logf("No account data. Skipping account data load.")
		return
	}

	pbCodec := address.Bech32Codec{Bech32Prefix: "pb"}
	curCodec := address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()}
	convertAddr := func(addr string) (sdk.AccAddress, string) {
		if len(addr) == 0 {
			return nil, ""
		}
		bz, err := pbCodec.StringToBytes(addr)
		require.NoError(t, err, "pbCodec.StringToBytes(%q)", addr)
		str, err := curCodec.BytesToString(bz)
		require.NoError(t, err, "curCodec.BytesToString(bz) bz from %q", addr)
		return bz, str
	}
	parseJSON := func(data []byte, v proto.Message, msg string, args ...interface{}) {
		err := app.appCodec.UnmarshalJSON(data, v)
		require.NoErrorf(t, err, "appCodec.UnmarshalJSON: "+msg, args...)
	}
	// I mostly just didn't want an err variable defined yet.
	validators := func() []stakingtypes.Validator {
		validators, err := app.StakingKeeper.GetAllValidators(ctx)
		require.NoError(t, err, "StakingKeeper.GetAllValidators()")
		return validators
	}()
	require.NotEmpty(t, validators, "validators")
	getValidator := func(i int) stakingtypes.Validator {
		if len(validators) == 1 {
			return validators[0]
		}
		return validators[i%len(validators)]
	}

	lastAcctNum := uint64(0)
	initialAcctNums := func() map[string]uint64 {
		rv := make(map[string]uint64)
		err := app.AccountKeeper.Accounts.Walk(ctx, nil, func(addr sdk.AccAddress, acct sdk.AccountI) (stop bool, err error) {
			key, err := pbCodec.BytesToString(addr)
			require.NoError(t, err, "initial account: pbCodec.BytesToString(...) %s", addr)
			val := acct.GetAccountNumber()
			rv[key] = val
			if val > lastAcctNum {
				lastAcctNum = val
			}
			return false, nil
		})
		require.NoError(t, err, "initial accounts: AccountKeeper.Accounts.Walk")
		return rv
	}()
	require.NotEmpty(t, initialAcctNums, "initialAcctNums")
	getAcctNum := func(pbAddr string) uint64 {
		rv, ok := initialAcctNums[pbAddr]
		if ok {
			return rv
		}
		lastAcctNum++
		return lastAcctNum
	}
	setAcctNum := func(i int, pbAddr string, acct sdk.AccountI) {
		acctNum := getAcctNum(pbAddr)
		err := acct.SetAccountNumber(acctNum)
		require.NoError(t, err, "[%d]: %q: SetAccountNumber(%d)", i, pbAddr, acctNum)
	}
	setAccount := func(i int, pbAddr string, acct sdk.AccountI) {
		setAcctNum(i, pbAddr, acct)
		app.AccountKeeper.SetAccount(ctx, acct)
	}

	accountData := func() []JSONAccountEntry {
		var rv []JSONAccountEntry
		err := json.Unmarshal(AccountData, &rv)
		require.NoErrorf(t, err, "json.Unmarshal: account data file contents")
		return rv
	}()

	type allAccts struct {
		Base              []*authtypes.BaseAccount
		Module            []*authtypes.ModuleAccount
		BaseVesting       []*vesting.BaseVestingAccount
		ContinuousVesting []*vesting.ContinuousVestingAccount
		DelayedVesting    []*vesting.DelayedVestingAccount
		PeriodicVesting   []*vesting.PeriodicVestingAccount
		PermanentLocked   []*vesting.PermanentLockedAccount
		Market            []*exchange.MarketAccount
		Marker            []*markertypes.MarkerAccount
		Interchain        []*ica.InterchainAccount
	}
	accts := allAccts{}

	totalAccounts := 0
	totalWASMAccounts := 0
	totalNhash := sdkmath.ZeroInt()
	totalDel := sdkmath.ZeroInt()
	totalDelDec := sdkmath.LegacyZeroDec()
	totalFunds := make(map[string]sdkmath.Int)
	totalRew := sdkmath.LegacyZeroDec()

	addToCoinMap := func(m map[string]sdkmath.Int, coins sdk.Coins) {
		for _, coin := range coins {
			if amt, ok := m[coin.Denom]; ok {
				m[coin.Denom] = amt.Add(coin.Amount)
			} else {
				m[coin.Denom] = coin.Amount
			}
		}
	}
	coinMapToCoins := func(m map[string]sdkmath.Int) sdk.Coins {
		rv := make(sdk.Coins, 0, len(m))
		for denom, amount := range m {
			rv = append(rv, sdk.NewCoin(denom, amount))
		}
		rv.Sort()
		return rv
	}

	proccessEntryFunc := func(i int, entry JSONAccountEntry) func() {
		return func() {
			var addr sdk.AccAddress
			entryType := strings.TrimSuffix(entry.Type, "-wasm")
			switch entryType {
			case "Base":
				var acct authtypes.BaseAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.Base = append(accts.Base, &acct)
			case "Module":
				var acct authtypes.ModuleAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.Module = append(accts.Module, &acct)
			case "BaseVesting":
				var acct vesting.BaseVestingAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.BaseVesting = append(accts.BaseVesting, &acct)
			case "ContinuousVesting":
				var acct vesting.ContinuousVestingAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.ContinuousVesting = append(accts.ContinuousVesting, &acct)
			case "DelayedVesting":
				var acct vesting.DelayedVestingAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.DelayedVesting = append(accts.DelayedVesting, &acct)
			case "PeriodicVesting":
				var acct vesting.PeriodicVestingAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.PeriodicVesting = append(accts.PeriodicVesting, &acct)
			case "PermanentLocked":
				var acct vesting.PermanentLockedAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.PermanentLocked = append(accts.PermanentLocked, &acct)
			case "exchange.Market":
				var acct exchange.MarketAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)

				setAccount(i, entry.Addr, &acct)
				accts.Market = append(accts.Market, &acct)
			case "Marker":
				var acct markertypes.MarkerAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)
				_, acct.Manager = convertAddr(acct.Manager)
				for g, grant := range acct.AccessControl {
					_, acct.AccessControl[g].Address = convertAddr(grant.Address)
				}

				setAcctNum(i, entry.Addr, &acct)
				app.MarkerKeeper.SetMarker(ctx, &acct)
				accts.Marker = append(accts.Marker, &acct)
			case "Interchain":
				var acct ica.InterchainAccount
				parseJSON(entry.Account, &acct, "[%d]: account %q", i, entry.Addr)
				addr, acct.Address = convertAddr(acct.Address)
				_, acct.AccountOwner = convertAddr(acct.AccountOwner)

				setAccount(i, entry.Addr, &acct)
				accts.Interchain = append(accts.Interchain, &acct)
			}

			var acctFund sdk.Coins
			if len(entry.Nhash) > 0 {
				nhash, err := sdk.ParseCoinNormalized(entry.Nhash + "nhash")
				require.NoError(t, err, "[%d]: Nhash: ParseCoinNormalized(%q) nhash for %s", i, entry.Nhash+"nhash", entry.Addr)
				if !nhash.IsNil() && !nhash.IsZero() {
					acctFund = acctFund.Add(nhash)
					totalNhash = totalNhash.Add(nhash.Amount)
				}
			}

			if !acctFund.IsZero() {
				err := fundAccount(ctx, app, addr, acctFund)
				require.NoError(t, err, "[%d]: FundAccount(%s, %s) for %s", i, addr, acctFund, entry.Addr)
				addToCoinMap(totalFunds, acctFund)
			}

			if len(entry.Delegated) > 0 {
				amtDec := parseDec(t, i, entry.Delegated, "Delegated")
				amt := decToInt(amtDec)
				if !amt.IsNil() && !amt.IsZero() {
					val := getValidator(i)
					_, err := app.StakingKeeper.Delegate(ctx, addr, amt, stakingtypes.Bonded, val, false)
					require.NoError(t, err, "[%d]: Delegate(%s, %s, %s)", addr, amt, val.OperatorAddress)
					totalDel = totalDel.Add(amt)
				}
				totalDelDec = totalDelDec.Add(amtDec)
			}

			if len(entry.UnclaimedRewards) > 0 {
				totalRew = totalRew.Add(parseDec(t, i, entry.UnclaimedRewards, "UnclaimedRewards"))
			}

			if entry.IsWASM {
				totalWASMAccounts++
			}
			totalAccounts++
		}
	}

	for i, entry := range accountData {
		require.NotPanics(t, proccessEntryFunc(i, entry), "[%d]: %s %s %s", i, entry.Type, entry.Name, entry.Addr)
	}

	for _, acct := range accts.Marker {
		if !acct.SupplyFixed {
			continue
		}
		amt, ok := totalFunds[acct.Denom]
		if !ok {
			amt = sdkmath.ZeroInt()
		}
		if !amt.Equal(acct.Supply) {
			t.Logf("Adjusting supply of %s to %s (from %s)", acct.Denom, amt, acct.Supply)
			acct.Supply = amt
			app.AccountKeeper.SetAccount(ctx, acct)
		}
	}

	al := AccountLister{App: app}
	t.Logf("Number of accounts  ---------------> %27s", al.AddCommas(strconv.Itoa(totalAccounts)))
	t.Logf("Number of WASM accounts  ----------> %27s", al.AddCommas(strconv.Itoa(totalWASMAccounts)))
	t.Logf("Total nhash in accounts  ----------> %27s", al.AddCommas(totalNhash.String()))
	t.Logf("Total to delegate accounts  -------> %49s", al.AddCommas(totalDelDec.String()))
	t.Logf("Total delegated from accounts  ----> %27s", al.AddCommas(totalDel.String()))
	t.Logf("Total rewards for accounts  -------> %49s", al.AddCommas(totalRew.String()))
	t.Logf("Total of all funds  ---------------> %27s", al.AddCommas(coinMapToCoins(totalFunds).String()))
}

// fundAccount is like testutil.FundAccount but doesn't check for blocked addresses.
func fundAccount(ctx sdk.Context, app *App, addr sdk.AccAddress, amounts sdk.Coins) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error funding account %q: %v", addr.String(), r)
		}
	}()

	if err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return fmt.Errorf("error minting coins: %w", err)
	}

	return app.BankKeeper.SendCoins(markertypes.WithBypass(ctx), app.AccountKeeper.GetModuleAddress(minttypes.ModuleName), addr, amounts)
}

// parseDec converts the provided str into a LegacyDec.
func parseDec(t *testing.T, i int, str, name string) sdkmath.LegacyDec {
	t.Helper()
	parts := strings.Split(str, ".")
	switch len(parts) {
	case 1:
		parts = append(parts, "000000000000000000")
	case 2:
		parts[1] += strings.Repeat("0", 18-len(parts[1]))
	default:
		t.Fatalf("[%d]: %s: strings.Split(%q, \".\") = %q must have exactly 1 or 2 entries", i, name, str, parts)
	}
	if len(parts[0]) == 0 {
		parts[0] = "0"
	}
	val := parts[0] + "." + parts[1]
	amt, err := sdkmath.LegacyNewDecFromStr(val)
	require.NoError(t, err, "[%d]: %s: sdkmath.LegacyNewDecFromStr(%q)", i, name, val)
	return amt
}

// pointZeroOne is a LegacyDec equal to 0.01.
var pointZeroOne = sdkmath.LegacyMustNewDecFromStr("0.01")

// decToInt converts a LegacyDec to an int by truncating it (but rounding .99 up).
func decToInt(amt sdkmath.LegacyDec) sdkmath.Int {
	// Some look like float errors (e.g. 4320002600351.999999999999999999).
	// So we essentially round .99 up and truncate everything else.
	return amt.Add(pointZeroOne).TruncateInt()
}

// exportGenesis will export the current state as a genesis file.
func exportGenesis(t *testing.T, ctx sdk.Context, app *App) {
	t.Helper()
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err, "StakingKeeper.GetAllValidators() error")
	require.NotEmpty(t, validators, "StakingKeeper.GetAllValidators() result")

	finalizeBlockReq := &abci.RequestFinalizeBlock{
		Hash:   app.LastCommitID().Hash,
		Height: app.LastBlockHeight() + 1,
		// DecidedLastCommit: abci.CommitInfo{},
		// NextValidatorsHash: nil,
		// ProposerAddress:    nil,
		// Misbehavior:       nil,
		// Time:               time.Now(),
		// Txs:               nil,
	}
	_, err = app.FinalizeBlock(finalizeBlockReq)
	require.NoError(t, err, "FinalizeBlock")

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err, "ExportAppStateAndValidators(false, nil, nil)")
	appGen := &genutiltypes.AppGenesis{
		AppName:       version.AppName,
		AppVersion:    version.Version,
		GenesisTime:   time.Now(),
		ChainID:       app.ChainID(),
		InitialHeight: exported.Height,
		AppHash:       nil, // Not sure where to get this. It's null in the `make run` gen file.
		AppState:      exported.AppState,
		Consensus:     genutiltypes.NewConsensusGenesis(exported.ConsensusParams, exported.Validators),
	}

	outDir := "/Users/dannywedul/git/provenance/app/testdata"
	if _, err = os.Stat(outDir); err != nil {
		t.Logf("Output directory does not exist. Not writing resulting genesis file.")
		return
	}

	outFile := outDir + "/genesis.json"
	err = appGen.SaveAs(outFile)
	require.NoError(t, err, "appGen.SaveAs(%q)", outFile)
	t.Logf("Genesis file written to: %s", outFile)
}

func (s *AccountDataTestSuite) TestLoadDataAndExportGenesis() {
	loadAccountData(s.T(), s.ctx, s.app)
	exportGenesis(s.T(), s.ctx, s.app)
}

func TestSumAmounts(t *testing.T) {
	if len(AccountData) == 0 {
		t.Skip("No account data to load.")
	}
	var accountData []JSONAccountEntry
	err := json.Unmarshal(AccountData, &accountData)
	require.NoErrorf(t, err, "json.Unmarshal: AccountData")

	totalNash := sdkmath.ZeroInt()
	totalDel := sdkmath.LegacyZeroDec()
	totalRew := sdkmath.LegacyZeroDec()
	for i, entry := range accountData {
		if len(entry.Nhash) > 0 {
			amt, ok := sdkmath.NewIntFromString(entry.Nhash)
			require.True(t, ok, "[%d]: Nhash: NewIntFromString(%q)", i, entry.Nhash)
			totalNash = totalNash.Add(amt)
		}

		if len(entry.Delegated) > 0 {
			amt := parseDec(t, i, entry.Delegated, "Delegated")
			totalDel = totalDel.Add(amt)
		}

		if len(entry.UnclaimedRewards) > 0 {
			amt := parseDec(t, i, entry.UnclaimedRewards, "UnclaimedRewards")
			totalRew = totalRew.Add(amt)
		}
	}

	t.Logf("Total nhash: %s", totalNash)
	t.Logf("Total Delegated: %s", totalDel)
	t.Logf("Total Rewards: %s", totalRew)
}
