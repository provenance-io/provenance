package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	copier "github.com/otiai10/copy"
	"github.com/spf13/cobra"

	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

// Migrator is an object to help guide a migration.
type Migrator struct {
	HomePath      string
	SourceDBType  string
	TargetDBType  string
	SourceDataDir string
	TargetDataDir string
	BackupDataDir string

	ToConvert []string
	ToCopy    []string
}

// SetUpMigrator creates a new Migrator with the given info and sets it up so that it's ready for a migration.
func SetUpMigrator(homePath, sourceDBType, targetDBType, backupDir string) (*Migrator, error) {
	var err error
	m := Migrator{
		HomePath:      homePath,
		SourceDBType:  sourceDBType,
		TargetDBType:  targetDBType,
		SourceDataDir: filepath.Join(homePath, "data"),
		BackupDataDir: backupDir,
	}
	if len(m.BackupDataDir) == 0 {
		m.BackupDataDir = filepath.Join(homePath, fmt.Sprintf("data-%s-%s", time.Now().Format("2006-01-02-15-04"), sourceDBType))
	}
	m.ToConvert, m.ToCopy, err = GetDataDirContents(m.SourceDataDir)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", m.SourceDataDir, err)
	}
	if len(m.ToConvert) == 0 {
		return nil, fmt.Errorf("no database directories found in %q", m.SourceDataDir)
	}
	m.TargetDataDir, err = os.MkdirTemp(homePath, "data-dbmigrate-tmp-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temporariy target data directory: %w", err)
	}
	return &m, nil
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
	defer batch.Close()

	logger.Info("Starting migration")
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		if v == nil {
			v = []byte{}
		}
		if err = batch.Set(iter.Key(), v); err != nil {
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

// UpdateConfig updates the config file to reflect the new database type.
func (m Migrator) UpdateConfig(logger tmlog.Logger, command *cobra.Command) error {
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
