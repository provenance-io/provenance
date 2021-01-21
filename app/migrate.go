package app

// This file implements a genesis migration from sdk 0.39 to 0.40 based Provenance Blockchains. It migrates state from
//   the modules as well including marker, metadata, name, and account (attribute)
// This file also implements setting an initial height from an upgrade.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	captypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	evtypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibcxfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibccoretypes "github.com/cosmos/cosmos-sdk/x/ibc/core/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagGenesisTime   = "genesis-time"
	flagInitialHeight = "initial-height"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [genesis-file]",
		Short: "Migrate genesis to a specified target version",
		Long: fmt.Sprintf(`Migrate the source genesis into the target version and print to STDOUT.

Example:
$ %s migrate /path/to/genesis.json --chain-id=provenance-2 --genesis-time=2020-01-1T00:00:00Z --initial-height=5000
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			var err error

			importGenesis := args[0]

			jsonBlob, err := ioutil.ReadFile(importGenesis)

			if err != nil {
				return errors.Wrap(err, "failed to read provided genesis file")
			}

			genDoc, err := tmtypes.GenesisDocFromJSON(jsonBlob)
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis document from file %s", importGenesis)
			}

			var initialState types.AppMap
			if err = json.Unmarshal(genDoc.AppState, &initialState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			firstMigration := "v0.40"

			migrationFunc := cli.GetMigrationCallback(firstMigration)
			if migrationFunc == nil {
				return fmt.Errorf("unknown migration function for version: %s", firstMigration)
			}

			// TODO: handler error from migrationFunc call
			newGenState := migrationFunc(initialState, clientCtx)

			var stakingGenesis staking.GenesisState

			clientCtx.JSONMarshaler.MustUnmarshalJSON(newGenState[staking.ModuleName], &stakingGenesis)

			ibcTransferGenesis := ibcxfertypes.DefaultGenesisState()
			ibcCoreGenesis := ibccoretypes.DefaultGenesisState()
			capGenesis := captypes.DefaultGenesis()
			evGenesis := evtypes.DefaultGenesisState()

			ibcTransferGenesis.Params.ReceiveEnabled = false
			ibcTransferGenesis.Params.SendEnabled = false

			ibcCoreGenesis.ClientGenesis.Params.AllowedClients = []string{exported.Tendermint}
			stakingGenesis.Params.HistoricalEntries = 10000

			newGenState[ibcxfertypes.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(ibcTransferGenesis)
			newGenState[host.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(ibcCoreGenesis)
			newGenState[captypes.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(capGenesis)
			newGenState[evtypes.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(evGenesis)
			newGenState[staking.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(&stakingGenesis)

			genDoc.AppState, err = json.Marshal(newGenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

			genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				err = t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				genDoc.GenesisTime = t
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			initialHeight, _ := cmd.Flags().GetInt(flagInitialHeight)

			genDoc.InitialHeight = int64(initialHeight)

			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().Int(flagInitialHeight, 0, "Set the starting height for the chain")
	cmd.Flags().String(flags.FlagChainID, "", "override chain_id with this flag")

	return cmd
}
