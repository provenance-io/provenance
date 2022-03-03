package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	copier "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

// Migrator is an object to help guide a migration.
type Migrator struct {
	// HomePath is the path to the home directory (should contain the config and data directories)
	HomePath string
	// SourceDBType is the type of the source (current) DB.
	SourceDBType string
	// TargetDBType is the type of the target (new) DB.
	TargetDBType string
	// SourceDataDir is the path to the source (current) data directory.
	SourceDataDir string
	// TargetDataDir is the path to the target (new) data directory.
	TargetDataDir string
	// BackupDataDir is the path to where the current data directory will be moved to when done.
	// That is, the current data directory will become this directory (as opposed to being placed in it as a sub-directory).
	BackupDataDir string

	// BatchSize is the threshold (in bytes) after which a batch is written and a new batch is created.
	BatchSize uint

	// ToConvert is all of the DB directories to migrate/convert.
	// Each entry is relative to the data directory.
	ToConvert []string
	// ToCopy is all the non-DB files and directories that should be copied from the source to the target.
	// Each entry is relative to the data directory.
	ToCopy []string
}

// SetUpMigrator creates a new Migrator with the given info and initializes it.
func SetUpMigrator(homePath, sourceDBType, targetDBType, backupDir string, batchSizeBytes uint) (*Migrator, error) {
	m := &Migrator{
		HomePath:      homePath,
		SourceDBType:  sourceDBType,
		TargetDBType:  targetDBType,
		SourceDataDir: filepath.Join(homePath, "data"),
		BackupDataDir: backupDir,
		BatchSize:     batchSizeBytes,
	}
	return m, m.Initialize()
}

// Initialize prepares this Migrator by doing the following:
//  1. If BackupDataDir is not set, sets it to the default.
//  2. If ToConvert is empty, Analyzes the SourceDataDir to identify ToConvert and ToCopy. Otherwise, ToConvert and ToCopy are not changed.
//  3. If TargetDataDir is defined, makes sure it exists. Otherwise, sets it and creates it.
func (m *Migrator) Initialize() error {
	var err error
	if len(m.BackupDataDir) == 0 {
		m.BackupDataDir = filepath.Join(m.HomePath, fmt.Sprintf("data-%s-%s", time.Now().Format("2006-01-02-15-04"), m.SourceDBType))
	}
	if len(m.ToConvert) == 0 {
		m.ToConvert, m.ToCopy, err = GetDataDirContents(m.SourceDataDir)
		if err != nil {
			return fmt.Errorf("error reading %q: %w", m.SourceDataDir, err)
		}
		if len(m.ToConvert) == 0 {
			return fmt.Errorf("no database directories found in %q", m.SourceDataDir)
		}
	}
	if len(m.TargetDataDir) == 0 {
		m.TargetDataDir, err = os.MkdirTemp(m.HomePath, "data-dbmigrate-tmp-*")
		if err != nil {
			return fmt.Errorf("error creating temporariy target data directory: %w", err)
		}
	} else {
		//nolint:gosec // This is the correct permissions for this backup directory.
		err = os.MkdirAll(m.TargetDataDir, 0755)
	}
	return nil
}

// Migrate converts all database dirs in the given homePath from the source underlying type to the target type.
// It then copies everything else in the data dir and swaps out the existing data dir for the newly created one.
func (m Migrator) Migrate(logger tmlog.Logger) error {
	logger.Info(fmt.Sprintf("Converting %d database directories from %q %s to %q %s",
		len(m.ToConvert), m.SourceDBType, m.SourceDataDir, m.TargetDBType, m.TargetDataDir))
	for i, dbDir := range m.ToConvert {
		err := m.MigrateDBDir(logger.With("db dir", fmt.Sprintf("%d/%d: %s", i+1, len(m.ToConvert), dbDir)), dbDir)
		if err != nil {
			return fmt.Errorf("could not convert %q from %q to %q: %w", dbDir, m.SourceDBType, m.TargetDBType, err)
		}
	}

	logger.Info(fmt.Sprintf("Copying %d items from %s to %s", len(m.ToCopy), m.SourceDataDir, m.TargetDataDir))
	for i, entry := range m.ToCopy {
		logger.Info(fmt.Sprintf("%d/%d: Copying %s", i+1, len(m.ToCopy), entry))
		if err := copier.Copy(filepath.Join(m.SourceDataDir, entry), filepath.Join(m.TargetDataDir, entry)); err != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err)
		}
	}

	logger.Info("Moving existing data directory to backup location.", "from", m.SourceDataDir, "to", m.BackupDataDir)
	if err := os.Rename(m.SourceDataDir, m.BackupDataDir); err != nil {
		return fmt.Errorf("could not back up existing data directory: %w", err)
	}

	logger.Info("Moving new data directory into place.", "from", m.TargetDataDir, "to", m.SourceDataDir)
	if err := os.Rename(m.TargetDataDir, m.SourceDataDir); err != nil {
		return fmt.Errorf("could not move new data directory into place: %w", err)
	}

	return nil
}

// MigrateDBDir creates a copy of the given db directory, converting it from one underlying type to another.
func (m Migrator) MigrateDBDir(logger tmlog.Logger, dbDir string) error {
	sourceDir, dbName := splitDBPath(m.SourceDataDir, dbDir)
	targetDir, _ := splitDBPath(m.TargetDataDir, dbDir)
	targetDataDirInfo, err := os.Stat(m.TargetDataDir)
	if err != nil {
		return fmt.Errorf("could not stat target data dir: %w", err)
	}
	targetDataDirMode := targetDataDirInfo.Mode()

	logger.Info("Setting up migration of directory.", "from", sourceDir, "to", targetDir)

	sourceDB, err := tmdb.NewDB(dbName, tmdb.BackendType(m.SourceDBType), sourceDir)
	if err != nil {
		return fmt.Errorf("could not open %q source db: %w", dbName, err)
	}
	defer sourceDB.Close()

	if targetDir != m.TargetDataDir {
		err = os.MkdirAll(targetDir, targetDataDirMode)
		if err != nil {
			return fmt.Errorf("could not create target sub-directory: %w", err)
		}
	}
	targetDB, err := tmdb.NewDB(dbName, tmdb.BackendType(m.TargetDBType), targetDir)
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
	defer func() {
		if batch != nil {
			batch.Close()
		}
	}()

	totalEntries := uint(0)
	batchEntries := uint(0)
	batchBytes := uint(0)
	logger.Info("Starting migration")
	for ; iter.Valid(); iter.Next() {
		totalEntries++
		batchEntries++
		v := iter.Value()
		if v == nil {
			v = []byte{}
		}
		k := iter.Key()
		if err = batch.Set(k, v); err != nil {
			return fmt.Errorf("could not set %q key/value: %w", dbName, err)
		}
		batchBytes += uint(len(v) + len(k))
		if batchBytes >= m.BatchSize {
			logger.Info("Writing batch and creating a new one.",
				"batch size (bytes)", commaString(batchBytes), "batch entries", commaString(batchEntries), "total entries", commaString(totalEntries))
			if err = batch.Write(); err != nil {
				return fmt.Errorf("could not write %q batch: %w", dbName, err)
			}
			if err = batch.Close(); err != nil {
				return fmt.Errorf("could not close %q batch: %w", dbName, err)
			}
			batch = targetDB.NewBatch()
			batchBytes = 0
			batchEntries = 0
		}
	}

	if batchBytes > 0 {
		logger.Info("Writing batch.",
			"batch size (bytes)", commaString(batchBytes), "batch entries", commaString(batchEntries), "total entries", commaString(totalEntries))
		if err = batch.Write(); err != nil {
			return fmt.Errorf("could not write %q batch: %w", dbName, err)
		}
		if err = batch.Close(); err != nil {
			return fmt.Errorf("could not close %q batch: %w", dbName, err)
		}
	}

	logger.Info("Done", "total entries", totalEntries)
	return nil
}

// UpdateConfig updates the config file to reflect the new database type.
func (m Migrator) UpdateConfig(logger tmlog.Logger, command *cobra.Command) error {
	// Warning: This wipes out all the viper setup stuff up to this point.
	// It needs to be done so that just the file values or defaults are loaded
	// without considering environment variables.
	// This is needed, at least, so that the log_level and log_format entries aren't changed.
	clientCtx := client.GetClientContextFromCmd(command)
	clientCtx.Viper = viper.New()
	server.GetServerContextFromCmd(command).Viper = clientCtx.Viper
	if err := client.SetCmdClientContext(command, clientCtx); err != nil {
		return err
	}

	// Now that we have a clean viper, load the config from files again.
	if err := config.LoadConfigFromFiles(command); err != nil {
		return err
	}

	logger.Info("Updating config.", "key", "db_backend", "from", m.SourceDBType, "to", m.TargetDBType)
	tmConfig, err := config.ExtractTmConfig(command)
	if err != nil {
		return fmt.Errorf("could not extract Tendermint config: %w", err)
	}
	tmConfig.DBBackend = m.TargetDBType
	config.SaveConfigs(command, nil, tmConfig, nil, false)
	logger.Info(fmt.Sprintf("%s Was: %s, Is Now: %s", "db_backend", m.SourceDBType, m.TargetDBType))
	return nil
}

// splitDBPath combine the provided path elements into a full path to a db dirctory, then
// breaks it down two parts:
// 1) A path to the directory to hold the db directory,
// 2) The name of the db.
// For example: "/foo", "bar/baz.db" will return "/foo/bar" and "baz".
func splitDBPath(elem ...string) (string, string) {
	base, name := filepath.Split(filepath.Join(elem...))
	return filepath.Clean(base), strings.TrimSuffix(name, ".db")
}

// GetDataDirContents gets the contents of a directory separated into database directories and non-database entries.
// The first return value will contain an entry for each database directory (including if they are in sub-directories).
// The second return value will contain all entries (files or directories) under dataDirPath that are not part of a database directory.
// Returned strings are relative to dataDirPath.
//
// Example return values:
//   return param 1: []string{"application.db", "blockstore.db", "evidence.db", "snapshots/metadata.db", "state.db", "tx_index.db"}
//   return param 2: []string{"cs.wal", "priv_validator_state.json", "wasm"}
func GetDataDirContents(dataDirPath string) ([]string, []string, error) {
	contents, err := ioutil.ReadDir(dataDirPath)
	if err != nil {
		return nil, nil, err
	}
	dbs := make([]string, 0)
	nonDBs := make([]string, 0)
	for _, entry := range contents {
		switch {
		case entry.IsDir():
			// goleveldb, cleveldb, and rocksdb name their db directories with a .db suffix.
			if filepath.Ext(entry.Name()) == ".db" {
				dbs = append(dbs, entry.Name())
			} else {
				subDBs, subNonDBs, err := GetDataDirContents(filepath.Join(dataDirPath, entry.Name()))
				if err != nil {
					return nil, nil, err
				}
				if len(subDBs) == 1 && subDBs[0] == "." {
					dbs = append(dbs, entry.Name())
				} else {
					for _, dbDir := range subDBs {
						dbs = append(dbs, filepath.Join(entry.Name(), dbDir))
					}
				}
				if len(subDBs) > 0 {
					for _, nonDBDir := range subNonDBs {
						nonDBs = append(nonDBs, filepath.Join(entry.Name(), nonDBDir))
					}
				} else {
					nonDBs = append(nonDBs, entry.Name())
				}
			}
		case strings.HasPrefix(entry.Name(), "MANIFEST"):
			// badger db does not use the .db suffix on their database directories.
			// So to identify them, we have to look for the MANIFEST files.
			// HasPrefix is used here instead of == because the other DB types have files that start with MANIFEST-
			// and so hopefully this will catch other db types that dont use the .db suffix on their directories.
			// The .db test is still also used to save some recursive calls and extra processing.
			return []string{"."}, nil, nil
		default:
			nonDBs = append(nonDBs, entry.Name())
		}
	}
	return dbs, nonDBs, nil
}

// commaString converts a positive integer to a string and adds commas.
func commaString(v uint) string {
	str := fmt.Sprintf("%d", v)
	if len(str) <= 3 {
		return str
	}
	rv := make([]rune, len(str)+(len(str)-1)/3)
	added := 0
	for i, c := range str {
		if i != 0 && (len(str)-i)%3 == 0 {
			rv[i+added] = ','
			added++
		}
		rv[i+added] = c
	}
	return string(rv)
}
