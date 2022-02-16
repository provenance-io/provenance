package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmdb "github.com/tendermint/tm-db"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/dbmigrate/utils"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

const FlagBackupDir = "backup-dir"

var PossibleDBTypes = []string{
	string(tmdb.RocksDBBackend), string(tmdb.BadgerDBBackend),
	string(tmdb.GoLevelDBBackend), string(tmdb.CLevelDBBackend),
}

// NewDBMigrateCmd creates a command for migrating the provenanced database from one underlying type to another.
func NewDBMigrateCmd() *cobra.Command {
	// Creating the client context early because the WithViper function
	// creates a new Viper instance which wipes out the existing global one.
	// Technically, it's not needed for the dbmigrate stuff, but having it
	// makes loading all the rest of the config stuff easier.
	clientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("PIO")

	rv := &cobra.Command{
		Use:   "dbmigrate <target type>",
		Short: "Provenanced Blockchain Database Migration Tool",
		Long: fmt.Sprintf(`Provenanced Blockchain Database Migration Tool
Converts an existing Provenance Blockchain Database to a new backend type.

Valid <target type> values: %s

Default Backup directory: {PIO_HOME}/data-{timestamp}-{old type}`, strings.Join(PossibleDBTypes, ", ")),
		Args: cobra.ExactArgs(1),
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			command.SetOut(command.OutOrStdout())
			command.SetErr(command.ErrOrStderr())

			if command.Flags().Changed(flags.FlagHome) {
				homeDir, _ := command.Flags().GetString(flags.FlagHome)
				clientCtx = clientCtx.WithHomeDir(homeDir)
			}

			if err := client.SetCmdClientContext(command, clientCtx); err != nil {
				return err
			}
			if err := config.InterceptConfigsPreRunHandler(command); err != nil {
				return err
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			targetDBType := strings.ToLower(args[0])
			if !IsPossibleDBType(targetDBType) {
				return fmt.Errorf("invalid target type: %q - must be one of: %s", targetDBType, strings.Join(PossibleDBTypes, ", "))
			}
			tmConfig, err := config.ExtractTmConfig(command)
			if err != nil {
				return fmt.Errorf("could not read Tendermint Config: %w", err)
			}
			sourceDBType := strings.ToLower(tmConfig.DBBackend)
			if !IsPossibleDBType(sourceDBType) {
				return fmt.Errorf("cannot convert source database of type: %q", sourceDBType)
			}
			logger := server.GetServerContextFromCmd(command).Logger
			if sourceDBType == targetDBType {
				logger.Info(fmt.Sprintf("Database already has type %q. Nothing to do.", targetDBType))
				return nil
			}
			// If this is just an empty string, the default ends up being used.
			backupDir, _ := command.Flags().GetString(FlagBackupDir)
			err = DoMigrateCmd(command, tmConfig.RootDir, sourceDBType, targetDBType, backupDir)
			if err != nil {
				logger.Error(err.Error())
				// If this returns an error, the help is printed. But that isn't wanted here.
				// But since we got an error, it shouldn't exit with code 0 either.
				// So we exit 1 here instead of returning an error and letting the caller handle the exit.
				os.Exit(1)
			}
			return nil
		},
	}
	rv.Flags().String(FlagBackupDir, "", "directory to back up the current data directory to (default is {home}/data-{timestamp}-{dbtype})")
	return rv
}

// Execute sets up and executes the provided command.
func Execute(command *cobra.Command) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	command.PersistentFlags().String(tmcli.HomeFlag, app.DefaultNodeHome, "directory for config and data")
	command.PersistentFlags().String(flags.FlagLogLevel, zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	command.PersistentFlags().String(flags.FlagLogFormat, tmcfg.LogFormatPlain, "The logging format (json|plain)")

	return command.ExecuteContext(ctx)
}

// IsPossibleDBType checks if the given dbType string is one that this migrator can handle.
func IsPossibleDBType(dbType string) bool {
	for _, p := range PossibleDBTypes {
		if dbType == p {
			return true
		}
	}
	return false
}

// DoMigrateCmd does all the work associated with the dbmigrate command (assuming that inputs have been validated).
func DoMigrateCmd(command *cobra.Command, homePath, sourceDBType, targetDBType, backupDir string) error {
	logger := server.GetServerContextFromCmd(command).Logger
	logger.Info("Setting up database migrations.", "home", homePath, "source type", sourceDBType, "target type", targetDBType)
	migrator, err := utils.SetUpMigrator(homePath, sourceDBType, targetDBType, backupDir)
	if err != nil {
		return err
	}
	logger.Info("Starting migrations.")
	err = migrator.Migrate(logger)
	if err != nil {
		return err
	}
	err = migrator.UpdateConfig(logger, command)
	if err != nil {
		return err
	}
	logger.Info("Done migrating databases.")
	return nil
}
