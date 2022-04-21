package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/dbmigrate/utils"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

const (
	FlagBackupDir  = "backup-dir"
	FlagBatchSize  = "batch-size"
	FlagStagingDir = "staging-dir"
	FlagStageOnly  = "stage-only"
)

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

	// Allow a user to define the log_level and log_format of this utility through the environment variables
	// DBM_LOG_LEVEL and DBM_LOG_FORMAT. Otherwise, we want to default them to info and plain.
	// Without this, the config's log_level and log_format are used.
	// So, for example, if the config has log_level = error, this utility wouldn't output anything unless it hits an error.
	// But that setting is desired mostly for the constant running of a node, as opposed to the single-time run of this utility.
	logLevel := "info"
	logFormat := "plain"
	if v := os.Getenv("DBM_LOG_LEVEL"); v != "" {
		logLevel = v
	}
	if v := os.Getenv("DBM_LOG_FORMAT"); v != "" {
		logLevel = logFormat
	}
	// Ignoring any errors here. If we can't set an environment variable, oh well.
	// Using the values from the config file isn't the end of the world, and is preferable to not allowing execution.
	_ = os.Setenv("PIO_LOG_LEVEL", logLevel)
	_ = os.Setenv("PIO_LOG_FORMAT", logFormat)

	rv := &cobra.Command{
		Use:   "dbmigrate <target type>",
		Short: "Provenance Blockchain Database Migration Tool",
		Long: fmt.Sprintf(`Provenance Blockchain Database Migration Tool
Converts an existing Provenance Blockchain Database to a new backend type.

Valid <target type> values: %s

Migration process:
1. Copy the current data directory into a staging data directory, migrating any databases appropriately.
   The staging directory is named data-dbmigrate-tmp-{timestamp}-{target dbtype}
   and by default will be in the {home} directory.
2. Move the current data directory to the backup location.
   The backup directory is named data-dbmigrate-backup-{timestamp}-{dbtypes}
   and by default will be in the {home} directoyr.
3. Move the staging data directory into place as the current data directory.
4. Update the config file to reflect the new database backend type.
`, strings.Join(utils.GetPossibleDBTypes(), ", ")),
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
			batchSizeMB, err := command.Flags().GetUint(FlagBatchSize)
			if err != nil {
				return fmt.Errorf("could not parse --%s option: %w", FlagBatchSize, err)
			}
			migrator := &utils.Migrator{
				TargetDBType: strings.ToLower(args[0]),
				HomePath:     client.GetClientContextFromCmd(command).HomeDir,
				BatchSize:    batchSizeMB * utils.BytesPerMB,
			}

			migrator.StageOnly, err = command.Flags().GetBool(FlagStageOnly)
			if err != nil {
				return fmt.Errorf("could not parse --%s flag: %w", FlagStageOnly, err)
			}

			migrator.BackupDir, err = command.Flags().GetString(FlagBackupDir)
			if err != nil {
				return fmt.Errorf("could not parse --%s option: %w", FlagBackupDir, err)
			}

			migrator.StagingDir, err = command.Flags().GetString(FlagStagingDir)
			if err != nil {
				return fmt.Errorf("could not parse --%s option: %w", FlagStagingDir, err)
			}

			err = DoMigrateCmd(command, migrator)
			if err != nil {
				server.GetServerContextFromCmd(command).Logger.Error(err.Error())
				// If this returns an error, the help is printed. But that isn't wanted here.
				// But since we got an error, it shouldn't exit with code 0 either.
				// So we exit 1 here instead of returning an error and letting the caller handle the exit.
				os.Exit(1)
			}
			return nil
		},
	}
	rv.Flags().String(FlagBackupDir, "", "directory to hold the backup directory (default {home})")
	rv.Flags().String(FlagStagingDir, "", "directory to hold the staging directory (default {home})")
	rv.Flags().Uint(FlagBatchSize, 2_048, "(in megabytes) after a batch reaches this size it is written and a new one is started (0 = unlimited)")
	rv.Flags().Bool(FlagStageOnly, false, "only migrate/copy the data (do not backup and replace the data directory and do not update the config)")
	return rv
}

// Execute sets up and executes the provided command.
func Execute(command *cobra.Command) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	command.PersistentFlags().String(tmcli.HomeFlag, app.DefaultNodeHome, "directory for config and data")

	return command.ExecuteContext(ctx)
}

// DoMigrateCmd does all the work associated with the dbmigrate command (assuming that inputs have been validated).
func DoMigrateCmd(command *cobra.Command, migrator *utils.Migrator) error {
	logger := server.GetServerContextFromCmd(command).Logger
	logger.Info("Setting up database migrations.")
	err := migrator.Initialize()
	if err != nil {
		return err
	}
	logger.Info("Starting migrations.")
	err = migrator.Migrate(logger)
	if err != nil {
		return err
	}
	if !migrator.StageOnly {
		logger.Info("Updating config.")
		var oldValue string
		oldValue, err = UpdateDBBackendConfigValue(command, migrator.TargetDBType)
		if err != nil {
			return err
		}
		logger.Info("Config Updated.", "key", "db_backend", "was", oldValue, "is now", migrator.TargetDBType)
	}
	logger.Info("Done migrating databases.")
	return nil
}

// UpdateDBBackendConfigValue updates the db backend value in the config file and returns the value it used to be.
func UpdateDBBackendConfigValue(command *cobra.Command, newValue string) (string, error) {
	// Warning: This wipes out all the viper setup stuff up to this point.
	// It needs to be done so that just the file values or defaults are loaded
	// without considering environment variables.
	// This is needed, at least, so that the log_level and log_format entries aren't changed.
	// It can't be undone because viper.New() overwrites the global Viper instance, and there is no way to set it back to what it was.
	// The contexts could get the original viper instance, but there's no guarantee that nothing uses the global functions.
	// So I figure it's best to at least keep them all in sync.
	// Ideally, it doesn't matter, though, since everything *should* be reloaded the same way (but who really knows).
	clientCtx := client.GetClientContextFromCmd(command)
	clientCtx.Viper = viper.New()
	server.GetServerContextFromCmd(command).Viper = clientCtx.Viper
	if err := client.SetCmdClientContext(command, clientCtx); err != nil {
		return "", fmt.Errorf("could not set client context: %w", err)
	}

	// Now that we have a clean viper, load the config from files again.
	if err := config.LoadConfigFromFiles(command); err != nil {
		return "", fmt.Errorf("could not load config from files: %w", err)
	}

	tmConfig, err := config.ExtractTmConfig(command)
	if err != nil {
		return "", fmt.Errorf("could not extract Tendermint config: %w", err)
	}
	oldValue := tmConfig.DBBackend
	tmConfig.DBBackend = newValue
	config.SaveConfigs(command, nil, tmConfig, nil, false)
	return oldValue, nil
}
