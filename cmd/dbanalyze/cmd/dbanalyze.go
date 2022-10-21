package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/dbanalyze/utils"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

// NewDBAnalyzeCmd creates a command for analyzing the keys/values stored in the application database.
func NewDBAnalyzeCmd() *cobra.Command {
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
	if v := os.Getenv("DBA_LOG_LEVEL"); v != "" {
		logLevel = v
	}
	if v := os.Getenv("DBA_LOG_FORMAT"); v != "" {
		logLevel = logFormat
	}
	// Ignoring any errors here. If we can't set an environment variable, oh well.
	// Using the values from the config file isn't the end of the world, and is preferable to not allowing execution.
	_ = os.Setenv("PIO_LOG_LEVEL", logLevel)
	_ = os.Setenv("PIO_LOG_FORMAT", logFormat)

	rv := &cobra.Command{
		Use:   "dbanalyze",
		Short: "Provenance Blockchain Database Analysis Tool",
		Long: `Provenance Blockchain Database Analysis Tool
Provides an overview of the application database broken down by modules and subkeys.
`,
		Args: cobra.NoArgs,
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
			migrator := &utils.Analyzer{
				HomePath: client.GetClientContextFromCmd(command).HomeDir,
			}

			err := DoAnalyzeCommand(command, migrator)
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
func DoAnalyzeCommand(command *cobra.Command, analyzer *utils.Analyzer) error {
	logger := server.GetServerContextFromCmd(command).Logger
	logger.Info("Initializing analyzer.")
	err := analyzer.Initialize()
	if err != nil {
		return err
	}
	logger.Info("Starting analysis...")
	err = analyzer.Analyze(logger)
	if err != nil {
		return err
	}

	logger.Info("Finished analysis.")
	return nil
}
