package cmd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/log"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/provenance-io/provenance/app"
	provenancecmd "github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/mocks"
	"github.com/provenance-io/provenance/x/exchange"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func TestAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	tests := []struct {
		name      string
		addr      string
		denom     string
		expectErr bool
	}{
		{
			name:      "invalid address",
			addr:      "",
			denom:     "1000atom",
			expectErr: true,
		},
		{
			name:      "valid address",
			addr:      addr1.String(),
			denom:     "1000atom",
			expectErr: false,
		},
		{
			name:      "multiple denoms",
			addr:      addr1.String(),
			denom:     "1000atom, 2000stake",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisAccountCmd(home)
			cmd.SetArgs([]string{
				tc.addr,
				tc.denom,
				fmt.Sprintf("--%s=home", flags.FlagHome)})

			if tc.expectErr {
				require.Error(t, cmd.ExecuteContext(ctx))
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}

func TestAddGenesisMsgFeeCmd(t *testing.T) {
	tests := []struct {
		name            string
		msgType         string
		fee             string
		msgFeeFloorCoin string
		expectErrMsg    string
	}{
		{
			name:            "invalid msg type",
			msgType:         "InvalidMsgType",
			fee:             "1000jackthecat",
			msgFeeFloorCoin: "0vspn",
			expectErrMsg:    "unable to resolve type URL /InvalidMsgType",
		},
		{
			name:            "invalid fee",
			msgType:         "/provenance.name.v1.MsgBindNameRequest",
			fee:             "not-a-fee",
			msgFeeFloorCoin: "0vspn",
			expectErrMsg:    "failed to parse coin: invalid decimal coin expression: not-a-fee",
		},
		{
			name:            "valid msg type and fee",
			msgType:         "/provenance.name.v1.MsgBindNameRequest",
			fee:             "1000jackthecat",
			msgFeeFloorCoin: "10jackthecat",
			expectErrMsg:    "",
		},
		{
			name:            "invalid fee",
			msgType:         "provenance.name.v1.MsgBindNameRequest",
			fee:             "1000jackthecat",
			msgFeeFloorCoin: "0vspn",
			expectErrMsg:    "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisCustomFloorPriceDenomCmd(home)
			cmdFee := provenancecmd.AddGenesisMsgFeeCmd(home, app.MakeEncodingConfig().InterfaceRegistry)
			cmd.SetArgs([]string{
				tc.msgFeeFloorCoin,
				fmt.Sprintf("--%s=home", flags.FlagHome)})
			cmdFee.SetArgs([]string{
				tc.msgType,
				tc.fee,
				fmt.Sprintf("--%s=home", flags.FlagHome)})

			if len(tc.expectErrMsg) > 0 {
				err = cmdFee.ExecuteContext(ctx)
				require.Error(t, err)
				require.Equal(t, tc.expectErrMsg, err.Error())
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
				require.NoError(t, cmdFee.ExecuteContext(ctx))
			}
		})
	}
}

// fixEmptiesInExchangeGenState updates all nil slices to be empty slices.
// The UnmarshalJSON function uses empty slices instead of nil, but it's cleaner and
// easier to define test cases by setting stuff to nil (or omitting the field completely).
func fixEmptiesInExchangeGenState(exGenState *exchange.GenesisState) {
	if exGenState.Params != nil && exGenState.Params.DenomSplits == nil {
		exGenState.Params.DenomSplits = make([]exchange.DenomSplit, 0)
	}

	if exGenState.Markets == nil {
		exGenState.Markets = make([]exchange.Market, 0)
	}
	for i, market := range exGenState.Markets {
		if market.FeeCreateAskFlat == nil {
			exGenState.Markets[i].FeeCreateAskFlat = make([]sdk.Coin, 0)
		}
		if market.FeeCreateBidFlat == nil {
			exGenState.Markets[i].FeeCreateBidFlat = make([]sdk.Coin, 0)
		}
		if market.FeeSellerSettlementFlat == nil {
			exGenState.Markets[i].FeeSellerSettlementFlat = make([]sdk.Coin, 0)
		}
		if market.FeeSellerSettlementRatios == nil {
			exGenState.Markets[i].FeeSellerSettlementRatios = make([]exchange.FeeRatio, 0)
		}
		if market.FeeBuyerSettlementFlat == nil {
			exGenState.Markets[i].FeeBuyerSettlementFlat = make([]sdk.Coin, 0)
		}
		if market.FeeBuyerSettlementRatios == nil {
			exGenState.Markets[i].FeeBuyerSettlementRatios = make([]exchange.FeeRatio, 0)
		}
		if market.ReqAttrCreateAsk == nil {
			exGenState.Markets[i].ReqAttrCreateAsk = make([]string, 0)
		}
		if market.ReqAttrCreateBid == nil {
			exGenState.Markets[i].ReqAttrCreateBid = make([]string, 0)
		}
		if market.AccessGrants == nil {
			exGenState.Markets[i].AccessGrants = make([]exchange.AccessGrant, 0)
		}
		for j, ag := range market.AccessGrants {
			if ag.Permissions == nil {
				exGenState.Markets[i].AccessGrants[j].Permissions = make([]exchange.Permission, 0)
			}
		}
	}

	if exGenState.Orders == nil {
		exGenState.Orders = make([]exchange.Order, 0)
	}
	for _, order := range exGenState.Orders {
		if bid := order.GetBidOrder(); bid != nil && bid.BuyerSettlementFees == nil {
			bid.BuyerSettlementFees = make(sdk.Coins, 0)
		}
	}
}

func TestAddGenesisDefaultMarketCmd(t *testing.T) {
	expDefaultMarket := func(marketID uint32, denom string, addrs ...string) exchange.Market {
		rv := provenancecmd.MakeDefaultMarket(denom, addrs)
		rv.MarketId = marketID
		return rv
	}
	addrs := []string{
		sdk.AccAddress("one_________________").String(),
		sdk.AccAddress("two_________________").String(),
		sdk.AccAddress("three_______________").String(),
	}

	tests := []struct {
		name          string
		iniAddrs      []string
		iniExGenState *exchange.GenesisState
		args          []string
		expErr        string
		expExGenState exchange.GenesisState
	}{
		{
			name:          "Incorrect usage",
			iniExGenState: exchange.DefaultGenesisState(),
			args:          []string{"--denom"},
			expErr:        "flag needs an argument: --denom",
		},
		{
			name:     "no exchange gen state in file, default denom",
			args:     []string{},
			iniAddrs: addrs[1:2],
			expExGenState: exchange.GenesisState{
				Markets: []exchange.Market{
					expDefaultMarket(1, "nhash", addrs[1]),
				},
				LastMarketId: 1,
			},
		},
		{
			name:          "default exchange gen state in file, denom provided",
			iniAddrs:      addrs,
			iniExGenState: exchange.DefaultGenesisState(),
			args:          []string{"--denom", "vspn"},
			expExGenState: exchange.GenesisState{
				Params: exchange.DefaultParams(),
				Markets: []exchange.Market{
					expDefaultMarket(1, "vspn", addrs...),
				},
				LastMarketId: 1,
			},
		},
		{
			name:     "file already has two markets",
			iniAddrs: addrs[0:2],
			iniExGenState: &exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 432},
				Markets: []exchange.Market{
					{
						MarketId:            1,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					{
						MarketId:            53,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
				},
				LastMarketId: 4,
				LastOrderId:  88,
			},
			args: []string{"--denom", "fruit"},
			expExGenState: exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 432},
				Markets: []exchange.Market{
					{
						MarketId:            1,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					{
						MarketId:            53,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					expDefaultMarket(2, "fruit", addrs[0], addrs[1]),
				},
				LastMarketId: 4,
				LastOrderId:  88,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixEmptiesInExchangeGenState(&tc.expExGenState)

			// Create a new home dir and initialize it.
			home := t.TempDir()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err, "setup: CreateDefaultTendermintConfig(%q)", home)
			cdc := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, cdc)
			require.NoError(t, err, "setup: ExecInitCmd")

			// Update the new genesis file to have the exchange genesis
			// state and just the accounts defined by this test case.
			genFile := cfg.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			require.NoError(t, err, "setup: GenesisStateFromGenFile")

			genDoc.GenesisTime = time.Unix(1618935600, 0) // 2021-04-20 16:20:00 +0000

			if tc.iniExGenState != nil {
				appState[exchange.ModuleName], err = cdc.MarshalJSON(tc.iniExGenState)
				require.NoError(t, err, "setup: MarshalJSON exchange genesis state")
			} else {
				if _, found := appState[exchange.ModuleName]; found {
					delete(appState, exchange.ModuleName)
				}
			}

			var authState authtypes.GenesisState
			if len(appState[authtypes.ModuleName]) > 0 {
				err = cdc.UnmarshalJSON(appState[authtypes.ModuleName], &authState)
				require.NoError(t, err, "setup: UnmarshalJSON auth genesis state")
			}

			oldAccs := authState.Accounts
			authState.Accounts = make([]*codectypes.Any, 0, len(oldAccs)+len(tc.iniAddrs))
			for _, acc := range oldAccs {
				_, isBase := acc.GetCachedValue().(*authtypes.BaseAccount)
				if !isBase {
					authState.Accounts = append(authState.Accounts, acc)
				}
			}

			newAccs := make([]authtypes.GenesisAccount, len(tc.iniAddrs))
			for i, addr := range tc.iniAddrs {
				accAddr, addrErr := sdk.AccAddressFromBech32(addr)
				require.NoError(t, addrErr, "setup: [%d]: AccAddressFromBech32(%q)", i, addr)
				newAccs[i] = authtypes.NewBaseAccountWithAddress(accAddr)
			}
			newAccAnys, err := authtypes.PackAccounts(newAccs)
			require.NoError(t, err, "setup: PackAccounts")
			authState.Accounts = append(authState.Accounts, newAccAnys...)

			appState[authtypes.ModuleName], err = cdc.MarshalJSON(&authState)
			require.NoError(t, err, "MarshalJSON auth genesis state")

			genDoc.AppState, err = json.Marshal(appState)
			require.NoError(t, err, "setup: Marshal app state")
			err = genutil.ExportGenesisFile(genDoc, genFile)
			require.NoError(t, err, "setup: ExportGenesisFile")

			// Create the contexts and set up the command.
			logger := log.NewNopLogger()
			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(cdc).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisDefaultMarketCmd(home)
			cmd.SetArgs(tc.args)

			// Run it!
			err = cmd.ExecuteContext(ctx)
			assertions.AssertErrorValue(t, err, tc.expErr, "%q ExecuteContext", cmd.Name())

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			appState, genDoc, err = genutiltypes.GenesisStateFromGenFile(genFile)
			require.NoError(t, err, "GenesisStateFromGenFile(%q)", genFile)
			var actExGenState exchange.GenesisState
			err = cdc.UnmarshalJSON(appState[exchange.ModuleName], &actExGenState)
			require.NoError(t, err, "UnmarshalJSON exchange genesis state")
			assert.Equal(t, tc.expExGenState, actExGenState, "genesis state after %s", cmd.Name())
		})
	}
}

func TestMakeDefaultMarket(t *testing.T) {
	addrs := []string{
		"one_________________",
		"two_________________",
		"three_______________",
	}
	coins := func(amount int64, denom string) []sdk.Coin {
		return []sdk.Coin{{Denom: denom, Amount: sdkmath.NewInt(amount)}}
	}
	ratios := func(denom string, priceAmt, feeAmt int64) []exchange.FeeRatio {
		return []exchange.FeeRatio{{
			Price: sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(priceAmt)},
			Fee:   sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(feeAmt)},
		}}
	}

	tests := []struct {
		name      string
		feeDenom  string
		addrs     []string
		expMarket exchange.Market
	}{
		{
			name:     "no denom, nil addrs",
			feeDenom: "",
			addrs:    nil,
			expMarket: exchange.Market{
				MarketDetails:       exchange.MarketDetails{Name: "Default Market"},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
			},
		},
		{
			name:     "no denom, one addr",
			feeDenom: "",
			addrs:    addrs[0:1],
			expMarket: exchange.Market{
				MarketDetails:       exchange.MarketDetails{Name: "Default Market"},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
				AccessGrants:        []exchange.AccessGrant{{Address: addrs[0], Permissions: exchange.AllPermissions()}},
			},
		},
		{
			name:     "no denom, three addrs",
			feeDenom: "",
			addrs:    addrs,
			expMarket: exchange.Market{
				MarketDetails:       exchange.MarketDetails{Name: "Default Market"},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
				AccessGrants: []exchange.AccessGrant{
					{Address: addrs[0], Permissions: exchange.AllPermissions()},
					{Address: addrs[1], Permissions: exchange.AllPermissions()},
					{Address: addrs[2], Permissions: exchange.AllPermissions()},
				},
			},
		},
		{
			name:     "empty addrs",
			feeDenom: "else",
			addrs:    []string{},
			expMarket: exchange.Market{
				MarketDetails:             exchange.MarketDetails{Name: "Default else Market"},
				FeeCreateAskFlat:          coins(100, "else"),
				FeeCreateBidFlat:          coins(100, "else"),
				FeeSellerSettlementFlat:   coins(500, "else"),
				FeeSellerSettlementRatios: ratios("else", 20, 1),
				FeeBuyerSettlementFlat:    coins(500, "else"),
				FeeBuyerSettlementRatios:  ratios("else", 20, 1),
				AcceptingOrders:           true,
				AllowUserSettlement:       true,
			},
		},
		{
			name:     "one addr",
			feeDenom: "vspn",
			addrs:    addrs[0:1],
			expMarket: exchange.Market{
				MarketDetails:             exchange.MarketDetails{Name: "Default vspn Market"},
				FeeCreateAskFlat:          coins(100, "vspn"),
				FeeCreateBidFlat:          coins(100, "vspn"),
				FeeSellerSettlementFlat:   coins(500, "vspn"),
				FeeSellerSettlementRatios: ratios("vspn", 20, 1),
				FeeBuyerSettlementFlat:    coins(500, "vspn"),
				FeeBuyerSettlementRatios:  ratios("vspn", 20, 1),
				AcceptingOrders:           true,
				AllowUserSettlement:       true,
				AccessGrants:              []exchange.AccessGrant{{Address: addrs[0], Permissions: exchange.AllPermissions()}},
			},
		},
		{
			name:     "three addrs",
			feeDenom: "nhash",
			addrs:    addrs,
			expMarket: exchange.Market{
				MarketDetails:             exchange.MarketDetails{Name: "Default nhash Market"},
				FeeCreateAskFlat:          coins(100, "nhash"),
				FeeCreateBidFlat:          coins(100, "nhash"),
				FeeSellerSettlementFlat:   coins(500, "nhash"),
				FeeSellerSettlementRatios: ratios("nhash", 20, 1),
				FeeBuyerSettlementFlat:    coins(500, "nhash"),
				FeeBuyerSettlementRatios:  ratios("nhash", 20, 1),
				AcceptingOrders:           true,
				AllowUserSettlement:       true,
				AccessGrants: []exchange.AccessGrant{
					{Address: addrs[0], Permissions: exchange.AllPermissions()},
					{Address: addrs[1], Permissions: exchange.AllPermissions()},
					{Address: addrs[2], Permissions: exchange.AllPermissions()},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actMarket exchange.Market
			testFunc := func() {
				actMarket = provenancecmd.MakeDefaultMarket(tc.feeDenom, tc.addrs)
			}
			require.NotPanics(t, testFunc, "MakeDefaultMarket")
			assert.Equal(t, tc.expMarket, actMarket, "MakeDefaultMarket result")
		})
	}
}

func TestAddGenesisCustomMarketCmd(t *testing.T) {
	tests := []struct {
		name          string
		iniExGenState *exchange.GenesisState
		args          []string
		expErr        string
		expExGenState exchange.GenesisState
	}{
		{
			name:          "Incorrect usage",
			iniExGenState: exchange.DefaultGenesisState(),
			args:          []string{"--create-ask", "nhash"},
			expErr:        "invalid coin expression: \"nhash\"",
		},
		{
			name: "no exchange gen state in file",
			args: []string{"--name", "My New Market", "--accepting-orders"},
			expExGenState: exchange.GenesisState{
				Markets: []exchange.Market{
					{
						MarketId:        1,
						MarketDetails:   exchange.MarketDetails{Name: "My New Market"},
						AcceptingOrders: true,
					},
				},
				LastMarketId: 1,
			},
		},
		{
			name:          "default exchange gen state in file",
			iniExGenState: exchange.DefaultGenesisState(),
			args:          []string{"--name", "THE market", "--market", "6"},
			expExGenState: exchange.GenesisState{
				Params: exchange.DefaultParams(),
				Markets: []exchange.Market{
					{
						MarketId:      6,
						MarketDetails: exchange.MarketDetails{Name: "THE market"},
					},
				},
			},
		},
		{
			name: "file already has two markets",
			iniExGenState: &exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 432},
				Markets: []exchange.Market{
					{
						MarketId:            1,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					{
						MarketId:            53,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
				},
				LastMarketId: 4,
				LastOrderId:  88,
			},
			args: []string{"--name", "A Better Market", "--create-bid", "50nhash", "--create-ask", "30nhash",
				"--accepting-orders", "--access-grants", sdk.AccAddress("addr1_______________").String() + ":all"},
			expExGenState: exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 432},
				Markets: []exchange.Market{
					{
						MarketId:            1,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					{
						MarketId:            53,
						MarketDetails:       exchange.MarketDetails{Name: "Market One"},
						AcceptingOrders:     true,
						AllowUserSettlement: true,
					},
					{
						MarketId:         2,
						MarketDetails:    exchange.MarketDetails{Name: "A Better Market"},
						FeeCreateAskFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 30)},
						FeeCreateBidFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 50)},
						AcceptingOrders:  true,
						AccessGrants: []exchange.AccessGrant{
							{
								Address:     sdk.AccAddress("addr1_______________").String(),
								Permissions: exchange.AllPermissions(),
							},
						},
					},
				},
				LastMarketId: 4,
				LastOrderId:  88,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixEmptiesInExchangeGenState(&tc.expExGenState)

			// Create a new home dir and initialize it.
			home := t.TempDir()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err, "setup: CreateDefaultTendermintConfig(%q)", home)
			cdc := sdksim.MakeTestEncodingConfig().Codec
			err = genutiltest.ExecInitCmd(testMbm, home, cdc)
			require.NoError(t, err, "setup: ExecInitCmd")

			// Update the new genesis file to have the exchange genesis state defined by this test case.
			genFile := cfg.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			require.NoError(t, err, "setup: GenesisStateFromGenFile")

			genDoc.GenesisTime = time.Unix(1618935600, 0) // 2021-04-20 16:20:00 +0000

			if tc.iniExGenState != nil {
				appState[exchange.ModuleName], err = cdc.MarshalJSON(tc.iniExGenState)
				require.NoError(t, err, "setup: MarshalJSON exchange genesis state")
			} else {
				if _, found := appState[exchange.ModuleName]; found {
					delete(appState, exchange.ModuleName)
				}
			}

			genDoc.AppState, err = json.Marshal(appState)
			require.NoError(t, err, "setup: Marshal app state")
			err = genutil.ExportGenesisFile(genDoc, genFile)
			require.NoError(t, err, "setup: ExportGenesisFile")

			// Create the contexts and set up the command.
			logger := log.NewNopLogger()
			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(cdc).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.AddGenesisCustomMarketCmd(home)
			cmd.SetArgs(tc.args)

			// Run it!
			err = cmd.ExecuteContext(ctx)
			assertions.AssertErrorValue(t, err, tc.expErr, "%q ExecuteContext", cmd.Name())

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			appState, genDoc, err = genutiltypes.GenesisStateFromGenFile(genFile)
			require.NoError(t, err, "GenesisStateFromGenFile(%q)", genFile)
			var actExGenState exchange.GenesisState
			err = cdc.UnmarshalJSON(appState[exchange.ModuleName], &actExGenState)
			require.NoError(t, err, "UnmarshalJSON exchange genesis state")
			assert.Equal(t, tc.expExGenState, actExGenState, "genesis state after %s", cmd.Name())
		})
	}
}

func TestAddMarketsToAppState(t *testing.T) {
	askOrder := *exchange.NewOrder(1).WithAsk(&exchange.AskOrder{
		MarketId: 1,
		Seller:   sdk.AccAddress("seller______________").String(),
		Assets:   sdk.NewInt64Coin("apple", 15),
		Price:    sdk.NewInt64Coin("plum", 100),
	})
	bidOrder := *exchange.NewOrder(2).WithBid(&exchange.BidOrder{
		MarketId:            1,
		Buyer:               sdk.AccAddress("buyer_______________").String(),
		Assets:              sdk.NewInt64Coin("apple", 30),
		Price:               sdk.NewInt64Coin("plum", 200),
		BuyerSettlementFees: make([]sdk.Coin, 0),
	})

	tests := []struct {
		name          string
		codec         codec.Codec // Defaults to the test encoding config codec.
		exGenState    exchange.GenesisState
		markets       []exchange.Market
		expExGenState exchange.GenesisState
		expErr        string
	}{
		{
			name:   "error unmarshalling exchange gen state",
			codec:  mocks.NewMockCodec().WithUnmarshalJSONErrs("injected error message"),
			expErr: "could not extract exchange genesis state: injected error message",
		},
		{
			name:   "error marshalling exchange gen state",
			codec:  mocks.NewMockCodec().WithMarshalJSONErrs("another injected error message"),
			expErr: "failed to marshal exchange genesis state: another injected error message",
		},
		{
			name: "no markets: none added",
			exGenState: exchange.GenesisState{
				Params:       &exchange.Params{DefaultSplit: 123},
				Markets:      nil,
				Orders:       []exchange.Order{askOrder, bidOrder},
				LastMarketId: 42,
				LastOrderId:  300,
			},
			expExGenState: exchange.GenesisState{
				Params:       &exchange.Params{DefaultSplit: 123},
				Markets:      nil,
				Orders:       []exchange.Order{askOrder, bidOrder},
				LastMarketId: 42,
				LastOrderId:  300,
			},
		},
		{
			name:       "no markets: one added, id 0",
			exGenState: exchange.GenesisState{},
			markets:    []exchange.Market{{MarketId: 0, MarketDetails: exchange.MarketDetails{Name: "some test"}}},
			expExGenState: exchange.GenesisState{
				Markets:      []exchange.Market{{MarketId: 1, MarketDetails: exchange.MarketDetails{Name: "some test"}}},
				LastMarketId: 1,
			},
		},
		{
			name:       "no markets: one added, id 3",
			exGenState: exchange.GenesisState{LastMarketId: 2},
			markets:    []exchange.Market{{MarketId: 3, MarketDetails: exchange.MarketDetails{Name: "some test"}}},
			expExGenState: exchange.GenesisState{
				Markets:      []exchange.Market{{MarketId: 3, MarketDetails: exchange.MarketDetails{Name: "some test"}}},
				LastMarketId: 2,
			},
		},
		{
			name: "two markets: two added",
			exGenState: exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 444, DenomSplits: []exchange.DenomSplit{{Denom: "nhash", Split: 555}}},
				Markets: []exchange.Market{
					{MarketId: 1, MarketDetails: exchange.MarketDetails{Name: "Market One"}},
					{MarketId: 8, MarketDetails: exchange.MarketDetails{Name: "Market Eight", Description: "Dude!"}},
				},
				Orders:       []exchange.Order{bidOrder, askOrder},
				LastMarketId: 3,
				LastOrderId:  76,
			},
			markets: []exchange.Market{
				{MarketId: 12, MarketDetails: exchange.MarketDetails{Name: "Market Twelve"}},
				{MarketId: 0, MarketDetails: exchange.MarketDetails{Name: "Market Four"}},
			},
			expExGenState: exchange.GenesisState{
				Params: &exchange.Params{DefaultSplit: 444, DenomSplits: []exchange.DenomSplit{{Denom: "nhash", Split: 555}}},
				Markets: []exchange.Market{
					{MarketId: 1, MarketDetails: exchange.MarketDetails{Name: "Market One"}},
					{MarketId: 8, MarketDetails: exchange.MarketDetails{Name: "Market Eight", Description: "Dude!"}},
					{MarketId: 12, MarketDetails: exchange.MarketDetails{Name: "Market Twelve"}},
					{MarketId: 2, MarketDetails: exchange.MarketDetails{Name: "Market Four"}},
				},
				Orders:       []exchange.Order{bidOrder, askOrder},
				LastMarketId: 3,
				LastOrderId:  76,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixEmptiesInExchangeGenState(&tc.expExGenState)

			appCdc := sdksim.MakeTestEncodingConfig().Codec
			egsBz, err := appCdc.MarshalJSON(&tc.exGenState)
			require.NoError(t, err, "MarshalJSON initial exchange genesis state")
			appState := map[string]json.RawMessage{exchange.ModuleName: egsBz}

			if tc.codec == nil {
				tc.codec = appCdc
			}
			home := t.TempDir()
			clientCtx := client.Context{}.WithCodec(tc.codec).WithHomeDir(home)

			testFunc := func() {
				err = provenancecmd.AddMarketsToAppState(clientCtx, appState, tc.markets...)
			}
			require.NotPanics(t, testFunc, "AddMarketsToAppState")
			assertions.AssertErrorValue(t, err, tc.expErr)

			var actExpGenState exchange.GenesisState
			err = appCdc.UnmarshalJSON(appState[exchange.ModuleName], &actExpGenState)
			assert.Equal(t, tc.expExGenState, actExpGenState, "exchange genesis state after AddMarketsToAppState")
		})
	}
}

func TestGetNextAvailableMarketID(t *testing.T) {
	tests := []struct {
		name       string
		exGenState exchange.GenesisState
		exp        uint32
	}{
		{
			name:       "nil markets",
			exGenState: exchange.GenesisState{Markets: nil},
			exp:        1,
		},
		{
			name:       "empty markets",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{}},
			exp:        1,
		},
		{
			name:       "no markets: last market id 100",
			exGenState: exchange.GenesisState{LastMarketId: 100},
			exp:        1,
		},
		{
			name:       "one market: 1",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 1}}},
			exp:        2,
		},
		{
			name:       "one market: 2",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 2}}},
			exp:        1,
		},
		{
			name:       "three markets: 1 2 3",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 1}, {MarketId: 2}, {MarketId: 3}}},
			exp:        4,
		},
		{
			name:       "three markets: 3 2 1",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 3}, {MarketId: 2}, {MarketId: 1}}},
			exp:        4,
		},
		{
			name:       "three markets: 1 4 1",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 1}, {MarketId: 4}, {MarketId: 1}}},
			exp:        2,
		},
		{
			name:       "three markets: 1 2 4",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 1}, {MarketId: 2}, {MarketId: 4}}},
			exp:        3,
		},
		{
			name:       "three markets: 1 3 4",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 1}, {MarketId: 3}, {MarketId: 4}}},
			exp:        2,
		},
		{
			name:       "three markets: 2 3 4",
			exGenState: exchange.GenesisState{Markets: []exchange.Market{{MarketId: 2}, {MarketId: 3}, {MarketId: 4}}},
			exp:        1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act uint32
			testFunc := func() {
				act = provenancecmd.GetNextAvailableMarketID(tc.exGenState)
			}
			require.NotPanics(t, testFunc, "GetNextAvailableMarketID")
			assert.Equal(t, tc.exp, act, "GetNextAvailableMarketID result")
		})
	}
}
