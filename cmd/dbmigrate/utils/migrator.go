package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

// BytesPerMB is the number of bytes in a megabyte.
const BytesPerMB = 1_048_576

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

	// TimeStarted is the time that the migration was started.
	TimeStarted time.Time
	// TimeFinished is the time that the migration was finished.
	TimeFinished time.Time
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
		err = os.MkdirAll(m.TargetDataDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating target data directory: %w", err)
		}
	}
	return nil
}

// Migrate converts all database dirs in the given homePath from the source underlying type to the target type.
// It then copies everything else in the data dir and swaps out the existing data dir for the newly created one.
func (m *Migrator) Migrate(logger tmlog.Logger) (err error) {
	// If this func doesn't complete fully, we want to output a message that the staging directory might still exist (and be quite large).
	// Make a done channel for indicating normal finish and a signal channel for capturing interrupt signals like ctrl+c.
	doneChan := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT)
	// If we're returning an error, add the log message, and then always do a little cleanup.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
		if err != nil {
			logger.Error("The staging directory might still exists due to error.", "dir", m.TargetDataDir)
		}
		close(doneChan)
		signal.Stop(sigChan)
		close(sigChan)
	}()
	// We need to identify the currently running process so that any captured signal can be sent back to it.
	proc, pErr := os.FindProcess(os.Getpid())
	if pErr != nil {
		return fmt.Errorf("could not identify the running process: %w", pErr)
	}
	// Monitor for the signals and handle them appropriately.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("The staging directory might still due to panic.", "dir", m.TargetDataDir)
			}
		}()
		select {
		case s := <-sigChan:
			signal.Stop(sigChan)
			logger.Error("The staging directory might still due to early termination.", "dir", m.TargetDataDir)
			err2 := proc.Signal(s)
			if err2 != nil {
				logger.Error("Error propagating signal.", "error", err2)
			}
			return
		case <-doneChan:
			return
		}
	}()

	// Now we can get started.
	m.TimeStarted = time.Now()
	logger.Info(fmt.Sprintf("Converting %d Individual DBs.", len(m.ToConvert)),
		"source type", m.SourceDBType, "source dir", m.SourceDataDir,
		"target type", m.TargetDBType, "staging dir", m.TargetDataDir,
		"batch size", commaString(m.BatchSize/BytesPerMB),
	)
	counts := map[string]uint{}
	for i, dbDir := range m.ToConvert {
		count, err2 := m.MigrateDBDir(logger.With("db", strings.TrimSuffix(dbDir, ".db"), "progress", fmt.Sprintf("%d/%d", i+1, len(m.ToConvert))), dbDir)
		if err2 != nil {
			return fmt.Errorf("could not convert %q from %q to %q: %w", dbDir, m.SourceDBType, m.TargetDBType, err2)
		}
		counts[dbDir] = count
	}

	logger.Info(fmt.Sprintf("Copying %d items from %s to %s", len(m.ToCopy), m.SourceDataDir, m.TargetDataDir))
	for i, entry := range m.ToCopy {
		logger.Info(fmt.Sprintf("%d/%d: Copying %s", i+1, len(m.ToCopy), entry))
		if err2 := copier.Copy(filepath.Join(m.SourceDataDir, entry), filepath.Join(m.TargetDataDir, entry)); err2 != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err2)
		}
	}

	logger.Info("Moving existing data directory to backup location.", "from", m.SourceDataDir, "to", m.BackupDataDir)
	if err2 := os.Rename(m.SourceDataDir, m.BackupDataDir); err2 != nil {
		return fmt.Errorf("could not back up existing data directory: %w", err2)
	}

	logger.Info("Moving new data directory into place.", "from", m.TargetDataDir, "to", m.SourceDataDir)
	if err2 := os.Rename(m.TargetDataDir, m.SourceDataDir); err2 != nil {
		return fmt.Errorf("could not move new data directory into place: %w", err2)
	}
	m.TimeFinished = time.Now()

	logger.Info(m.MakeSummaryString(counts))
	return nil
}

// MigrateDBDir creates a copy of the given db directory, converting it from one underlying type to another.
func (m Migrator) MigrateDBDir(logger tmlog.Logger, dbDir string) (uint, error) {
	sourceDir, dbName := splitDBPath(m.SourceDataDir, dbDir)
	targetDir, _ := splitDBPath(m.TargetDataDir, dbDir)
	logger.Info("Individual DB Migration: Setting up.", "from", sourceDir, "to", targetDir)

	// Define some counters used in log messages, and a function to make it easy to add them all to log messages.
	var err error
	tickerOn := 5 * time.Second
	tickerOff := 1_000_000 * time.Hour // = a little more than 114 years.
	writtenEntries := uint(0)
	batchEntries := uint(0)
	batchBytes := uint(0)
	batchIndex := uint(1)
	commonKeyVals := func() []interface{} {
		return []interface{}{
			"batch index", commaString(batchIndex),
			"batch size (megabytes)", commaString(batchBytes / BytesPerMB),
			"batch entries", commaString(batchEntries),
			"total entries", commaString(writtenEntries + batchEntries),
			"run time", fmt.Sprintf("%s", time.Since(m.TimeStarted)),
		}
	}

	// There's several things that need closing and sometimes calling close can cause a segmentation fault that,
	// for some reason, doesn't reach the sigChan notification set up in the Migrate method.
	// So just to have clearer control over closing order, they're all defined at once and closed in a single defer function.
	var sourceDB, targetDB tmdb.DB
	var iter tmdb.Iterator
	var batch tmdb.Batch
	var statusTicker, writeTicker *time.Ticker
	stopTickers := make(chan bool, 1)
	defer func() {
		// closing the stopTickers chan will trigger the status logging subprocess to finish up.
		close(stopTickers)
		// Then we can stop the tickers (just to be safe).
		if statusTicker != nil {
			statusTicker.Stop()
		}
		if writeTicker != nil {
			writeTicker.Stop()
		}
		// iter before sourceDB because closing the sourceDB might remove things needed for the iterator to close.
		if iter != nil {
			iter.Close()
		}
		if sourceDB != nil {
			sourceDB.Close()
		}
		// batch before targetDB because closing the targetDB might remove things needed for the batch to close.
		if batch != nil {
			batch.Close()
		}
		if targetDB != nil {
			targetDB.Close()
		}
	}()

	// Set up a couple different tickers for outputting different status messages at different times.
	writeTicker = time.NewTicker(tickerOff)
	statusTicker = time.NewTicker(tickerOff)
	go func() {
		for {
			select {
			case <-statusTicker.C:
				logger.Info("Status", commonKeyVals()...)
			case <-writeTicker.C:
				logger.Info("Still writing...", commonKeyVals()...)
			case <-stopTickers:
				return
			}
		}
	}()

	// In at least one case (the snapshots/metadata db), there's a sub-directory that needs to be created in order to
	// safely open a new database in it.
	if targetDir != m.TargetDataDir {
		var targetDataDirInfo os.FileInfo
		targetDataDirInfo, err = os.Stat(m.TargetDataDir)
		if err != nil {
			return 0, fmt.Errorf("could not stat target data dir: %w", err)
		}
		err = os.MkdirAll(targetDir, targetDataDirInfo.Mode())
		if err != nil {
			return 0, fmt.Errorf("could not create target sub-directory: %w", err)
		}
	}

	sourceDB, err = tmdb.NewDB(dbName, tmdb.BackendType(m.SourceDBType), sourceDir)
	if err != nil {
		return 0, fmt.Errorf("could not open %q source db: %w", dbName, err)
	}

	targetDB, err = tmdb.NewDB(dbName, tmdb.BackendType(m.TargetDBType), targetDir)
	if err != nil {
		return 0, fmt.Errorf("could not open %q target db: %w", dbName, err)
	}

	iter, err = sourceDB.Iterator(nil, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create %q source iterator: %w", dbName, err)
	}

	// There's a couple places in here where we need to write and close the batch. But the safety stuff (on errors)
	// is needed in both places, so it's pulled out into this anonymous function.
	writeAndCloseBatch := func() error {
		// Using WriteSync here instead of Write because sometimes the Close was causing a segfault, and mabye this helps?
		if err = batch.WriteSync(); err != nil {
			// If the write fails, closing the db can sometimes cause a segmentation fault.
			targetDB = nil
			return fmt.Errorf("could not write %q batch: %w", dbName, err)
		}
		writtenEntries += batchEntries
		err = batch.Close()
		if err != nil {
			// If closing the batch fails, closing the db can sometimes cause a segmentation fault.
			targetDB = nil
			// Similarly, calling close a second time can segfault.
			batch = nil
			return fmt.Errorf("could not close %q batch: %w", dbName, err)
		}
		return nil
	}

	logger.Info("Individual DB Migration: Starting.")
	batch = targetDB.NewBatch()
	statusTicker.Reset(tickerOn)
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		if v == nil {
			v = []byte{}
		}
		k := iter.Key()
		if err = batch.Set(k, v); err != nil {
			return writtenEntries, fmt.Errorf("could not set %q key/value: %w", dbName, err)
		}
		batchEntries++
		batchBytes += uint(len(v) + len(k))
		if m.BatchSize > 0 && batchBytes >= m.BatchSize {
			statusTicker.Reset(tickerOff)

			writeTicker.Reset(tickerOn)
			logger.Info("Writing intermediate batch.", commonKeyVals()...)
			if err = writeAndCloseBatch(); err != nil {
				return writtenEntries, err
			}
			writeTicker.Reset(tickerOff)

			logger.Info("Starting new batch.", commonKeyVals())
			batch = targetDB.NewBatch()
			batchIndex++
			batchBytes = 0
			batchEntries = 0
			statusTicker.Reset(tickerOn)
		}
	}
	statusTicker.Reset(tickerOff)

	if err = iter.Error(); err != nil {
		return writtenEntries, fmt.Errorf("iterator error: %w", err)
	}

	writeTicker.Reset(tickerOn)
	logger.Info("Writing final batch.", commonKeyVals()...)
	if err = writeAndCloseBatch(); err != nil {
		return writtenEntries, err
	}
	writeTicker.Reset(tickerOff)

	logger.Info("Individual DB Migration: Done.", "total entries", commaString(writtenEntries))
	return writtenEntries, nil
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

// MakeSummaryString creates a multi-line string with a summary of a migration.
func (m Migrator) MakeSummaryString(counts map[string]uint) string {
	var sb strings.Builder
	addLine := func(format string, a ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, a...) + "\n")
	}
	addLine("Summary:")
	if !m.TimeFinished.IsZero() && !m.TimeStarted.IsZero() {
		addLine("     Duration: %s", m.TimeFinished.Sub(m.TimeStarted))
	} else {
		addLine("     Duration: unknown")
	}
	addLine("         Home: %s", m.HomePath)
	addLine("         Data: %s", m.SourceDataDir)
	addLine("  Data Backup: %s", m.BackupDataDir)
	addLine("  New DB Type: %s", m.TargetDBType)
	addLine("  Copied (%d): %s", len(m.ToCopy), strings.Join(m.ToCopy, "  "))
	addLine("Migrated (%d):", len(m.ToConvert))
	for _, dbDir := range m.ToConvert {
		addLine("%22s: %11s entries", strings.TrimSuffix(dbDir, ".db"), commaString(counts[dbDir]))
	}
	return sb.String()
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
