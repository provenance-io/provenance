package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

// Options passed to compare command after initial processing.
type CompareOptions struct {
	SourceDir    string
	TargetDir    string
	DatabaseName string
}

// NewDBCompareCmd creates a command for comparing contents of a single database between a source and target folder.
func NewDBCompareCmd() *cobra.Command {
	// Creating the client context early because the WithViper function
	// creates a new Viper instance which wipes out the existing global one.
	// Technically, it's not needed for the dbmigrate stuff, but having it
	// makes loading all the rest of the config stuff easier.
	clientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("PIO")

	rv := &cobra.Command{
		Use:   "dbcompare <source dir> <target dir> <db name>",
		Short: "Provenance Blockchain Database Comparison Tool",
		Long:  "Provenance Blockchain Database Comparison Tool",
		Args:  cobra.ExactArgs(3),
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
			sourceDir, err := resolvePath(args[0])
			if err != nil {
				return fmt.Errorf("source directory could not be resolved: %w", err)
			}
			targetDir, err := resolvePath(args[1])
			if err != nil {
				return fmt.Errorf("target directory could not be resolved: %w", err)
			}
			opts := CompareOptions{
				SourceDir:    sourceDir,
				TargetDir:    targetDir,
				DatabaseName: args[2],
			}
			err = DoCompare(command, opts)
			if err != nil {
				return fmt.Errorf("unable to compare databases: %w", err)
			}
			return nil
		},
	}
	return rv
}

// Resolves path string as either absolute or falls back to relative and verifies path exists.
func resolvePath(path string) (string, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		return "", err
	}
	return resolved, nil
}

// Execute sets up and executes the provided command.
func Execute(command *cobra.Command) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	return command.ExecuteContext(ctx)
}

// Log settings before executing comparison.
func logSettings(opts CompareOptions, logger tmlog.Logger) error {
	bytes, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Comparing database using options:\n%s", string(bytes)))
	return nil
}

// Pretty prints data by removing special characters and trimming length.
func prettyPrintData(data string, escaper *regexp.Regexp, maxlen int) string {
	escaped := escaper.ReplaceAllString(string(data), "*")
	if len(escaped) > maxlen {
		return escaped[:maxlen]
	}
	return escaped
}

// Perform compare operation.
func DoCompare(command *cobra.Command, opts CompareOptions) error {
	logger := server.GetServerContextFromCmd(command).Logger
	logSettings(opts, logger)

	// Open the source database.
	sourcedb, err := tmdb.NewDB(opts.DatabaseName, tmdb.GoLevelDBBackend, opts.SourceDir)
	if err != nil {
		return fmt.Errorf("could not open %q source db: %w", opts.DatabaseName, err)
	}
	defer sourcedb.Close()

	// Open the target database.
	targetdb, err := tmdb.NewDB(opts.DatabaseName, tmdb.GoLevelDBBackend, opts.TargetDir)
	if err != nil {
		return fmt.Errorf("could not open %q target db: %w", opts.DatabaseName, err)
	}
	defer targetdb.Close()

	// Create an iterator for all keys in the source database.
	srcit, err := sourcedb.Iterator(nil, nil)
	if err != nil {
		return err
	}

	// Iterate through all keys.
	count := 0
	nospecial := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	for ; srcit.Valid(); srcit.Next() {
		key := srcit.Key()
		srcval := srcit.Value()
		tgtval, err := targetdb.Get(key)
		if err != nil {
			return err
		}
		keyenc := prettyPrintData(string(key), nospecial, 30)
		srcenc := prettyPrintData(string(srcval), nospecial, 30)
		if tgtval == nil {
			logger.Error(fmt.Sprintf("target does not contain key: %s (value %s)", keyenc, srcenc))
		} else if !bytes.Equal(srcval, tgtval) {
			tgtenc := prettyPrintData(string(tgtval), nospecial, 30)
			logger.Error(fmt.Sprintf("target value %s != %s for key: %s", srcenc, tgtenc, keyenc))
		}
		if count%1000000 == 0 {
			logger.Info(fmt.Sprintf("processed %d records", count))
		}
		count++
	}

	return nil
}
