package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

var PossibleDBTypes = []string{
	string(tmdb.RocksDBBackend), string(tmdb.BadgerDBBackend),
	string(tmdb.GoLevelDBBackend), string(tmdb.CLevelDBBackend),
}

// TODO: Create a converter params struct for all the args and pass that around instead of all the args.
// TODO: Add flag for defining a new backup directory.
// TODO: Add flag for defining the staging directory?

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
			err = convertDB(command, server.GetServerContextFromCmd(command).Logger, tmConfig.RootDir, sourceDBType, targetDBType)
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

func IsPossibleDBType(dbType string) bool {
	for _, p := range PossibleDBTypes {
		if dbType == p {
			return true
		}
	}
	return false
}

// convertDB converts all database dirs in the given homePath from the source underlying type to the target type.
func convertDB(command *cobra.Command, logger tmlog.Logger, homePath, sourceDBType, targetDBType string) error {
	logger.Info("Setting up database migration.", "home", homePath, "source type", sourceDBType, "target type", targetDBType)

	sourceDataDir := filepath.Join(homePath, "data")
	dbDirs, nonDBEntries, err := getDataDirContents(sourceDataDir)
	if err != nil {
		return fmt.Errorf("error reading database at %q: %w", sourceDataDir, err)
	}
	if len(dbDirs) == 0 {
		return fmt.Errorf("no database directories found in %q", sourceDataDir)
	}
	targetDataDir, err := ioutil.TempDir(homePath, "data-dbmigrate-tmp-*")
	if err != nil {
		return fmt.Errorf("error creating temporariy target data directory: %w", err)
	}

	logger.Info(fmt.Sprintf("Converting %d database directories from %q %s to %q %s",
		len(dbDirs), sourceDBType, sourceDataDir, targetDBType, targetDataDir))
	for i, dbDir := range dbDirs {
		err = convertDBDir(logger.With("db dir", fmt.Sprintf("%d/%d: %s", i+1, len(dbDirs), dbDir)),
			sourceDataDir, targetDataDir, dbDir, sourceDBType, targetDBType)
		if err != nil {
			return fmt.Errorf("could not convert %q from %q to %q: %w", dbDir, sourceDBType, targetDBType, err)
		}
	}

	logger.Info(fmt.Sprintf("Copying %d items from %s to %s", len(nonDBEntries), sourceDataDir, targetDataDir))
	for i, entry := range nonDBEntries {
		logger.Info(fmt.Sprintf("%d/%d: Copying %s", i+1, len(nonDBEntries), entry))
		if err = copy.Copy(filepath.Join(sourceDataDir, entry), filepath.Join(targetDataDir, entry)); err != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err)
		}
	}

	backupDataDir := filepath.Join(homePath, fmt.Sprintf("data-%s-%s", time.Now().Format("2006-01-02-15-04"), sourceDBType))
	logger.Info("Moving existing data directory to backup location.", "from", sourceDataDir, "to", backupDataDir)
	if err = os.Rename(sourceDataDir, backupDataDir); err != nil {
		return fmt.Errorf("could not back up existing db: %w", err)
	}

	logger.Info("Moving new data directory into place.", "from", targetDataDir, "to", sourceDataDir)
	if err = os.Rename(targetDataDir, sourceDataDir); err != nil {
		return fmt.Errorf("could not move new db into place: %w", err)
	}

	// TODO: Do superficial check of permissions/ownership and issue warning if different.

	logger.Info("Updating config.", "key", "db_backend", "from", sourceDBType, "to", targetDBType)
	tmConfig, err := config.ExtractTmConfig(command)
	if err != nil {
		return fmt.Errorf("could not extract Tendermint config: %w", err)
	}
	tmConfig.DBBackend = targetDBType
	config.SaveConfigs(command, nil, tmConfig, nil, false)
	logger.Info(fmt.Sprintf("%s Was: %s, Is Now: %s", "db_backend", sourceDBType, targetDBType))

	logger.Info("Done migrating database.")
	return nil
}

// getDataDirContents gets the contents of a directory split into database directories and non-database entries.
// The first return value will contain an entry for each database directory (including if they in sub-directories).
// The second return value will contain all entries (files or directories) under dataDirPath that are not part of a database directory.
// That is, it's all entries that will not be migrated, and should copied.
func getDataDirContents(dataDirPath string) ([]string, []string, error) {
	contents, err := ioutil.ReadDir(dataDirPath)
	if err != nil {
		return nil, nil, err
	}
	dbs := make([]string, 0)
	nonDBs := make([]string, 0)
	for _, entry := range contents {
		if entry.IsDir() {
			if filepath.Ext(entry.Name()) == ".db" {
				dbs = append(dbs, entry.Name())
			} else {
				subDBs, subNonDBs, err := getDataDirContents(filepath.Join(dataDirPath, entry.Name()))
				if err != nil {
					return nil, nil, err
				}
				for _, dbDir := range subDBs {
					dbs = append(dbs, filepath.Join(entry.Name(), dbDir))
				}
				if len(subDBs) > 0 {
					for _, nonDBDir := range subNonDBs {
						nonDBs = append(nonDBs, filepath.Join(entry.Name(), nonDBDir))
					}
				} else {
					nonDBs = append(nonDBs, entry.Name())
				}
			}
		} else {
			nonDBs = append(nonDBs, entry.Name())
		}
		if filepath.Ext(entry.Name()) == ".db" && entry.IsDir() {
			dbs = append(dbs, entry.Name())
		}
	}
	return dbs, nonDBs, nil
}

// convertDBDir converts a single db directory from one underlying type to another.
func convertDBDir(logger tmlog.Logger, sourceDataDir, targetDataDir, dbDir, sourceDBType, targetDBType string) error {
	sourceDir, dbName := splitDBPath(filepath.Join(sourceDataDir, dbDir))
	targetDir, _ := splitDBPath(filepath.Join(targetDataDir, dbDir))
	targetDataDirInfo, err := os.Stat(targetDataDir)
	if err != nil {
		return fmt.Errorf("could not stat target data dir: %w", err)
	}
	targetDataDirMode := targetDataDirInfo.Mode()

	logger.Info("Setting up migration of directory.", "from", sourceDir, "to", targetDir)

	sourceDB, err := tmdb.NewDB(dbName, tmdb.BackendType(sourceDBType), sourceDir)
	if err != nil {
		return fmt.Errorf("could not open %q source db: %w", dbName, err)
	}
	defer sourceDB.Close()

	if targetDir != targetDataDir {
		err = os.MkdirAll(targetDir, targetDataDirMode)
		if err != nil {
			return fmt.Errorf("could not create target sub-directory: %w", err)
		}
	}
	targetDB, err := tmdb.NewDB(dbName, tmdb.BackendType(targetDBType), targetDir)
	if err != nil {
		return fmt.Errorf("could not open %q target db: %w", dbName, err)
	}
	defer targetDB.Close()

	iter, err := sourceDB.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("could not create %q source iterator: %w", dbName, err)
	}
	defer iter.Close()

	batch := targetDB.NewBatch()
	defer batch.Close()

	logger.Info("Starting migration")
	for ; iter.Valid(); iter.Next() {
		if err = batch.Set(iter.Key(), iter.Value()); err != nil {
			return fmt.Errorf("could not set %q key/value: %w", dbName, err)
		}
	}

	logger.Info("Writing batch")
	if err = batch.Write(); err != nil {
		return fmt.Errorf("could not write %q batch: %w", dbName, err)
	}

	logger.Info("Done")
	return nil
}

// splitDBPath breaks down a path to a db directory into the path to the directory containing that and the db name.
// For example: "/foo/bar/baz.db" will return "/foo/bar" and "baz".
func splitDBPath(path string) (string, string) {
	base, name := filepath.Split(path)
	return filepath.Clean(base), strings.TrimSuffix(name, ".db")
}
