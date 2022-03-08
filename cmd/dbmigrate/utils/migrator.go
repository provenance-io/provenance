package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	copier "github.com/otiai10/copy"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

const (
	// BytesPerMB is the number of bytes in a megabyte.
	BytesPerMB = 1_048_576
	// TickerOff is a really long time period that can effectively turn a ticker off.
	// 1,000,000 hours is a little over 114 years (and is also about 1/3 max int64 as nanoseconds).
	TickerOff = 1_000_000 * time.Hour
)

// Migrator is an object to help guide a migration.
// The following MUST be defined: SourceDBType, TargetDBType.
//
// These must also be defined, but can have defaults applied using ApplyDefaults:
// StagingDir, BackupDir, SourceDataDir, StagingDataDir, BackupDataDir, Permissions, StatusPeriod, DirDateFormat.
//
// HomePath is not required, but should be set if you want to use any defaults.
//
// If SourceDataDir is defined, ReadSourceDataDir can be used to populate ToConvert and ToCopy.
//
// See Also: Initialize.
type Migrator struct {
	// HomePath is the path to the home directory (should contain the config and data directories).
	HomePath string
	// StagingDir is the directory that will hold the staging data directory.
	// Default is HomePath
	StagingDir string
	// BackupDir is the directory that will hold the backup data directory.
	// Default is HomePath
	BackupDir string

	// SourceDBType is the type of the source (current) DB.
	SourceDBType string
	// TargetDBType is the type of the target (new) DB.
	TargetDBType string

	// SourceDataDir is the path to the source (current) data directory.
	// Default is { HomePath }/data
	SourceDataDir string
	// StagingDataDir is the path to the staging (new) data directory.
	// Default is { StagingDir }/data-dbmigrate-tmp-{timestamp}-{ TargetDBType }
	StagingDataDir string
	// BackupDataDir is the path to where the current data directory will be moved to when done.
	// Default is { BackupDir }/data-{timestamp}-{ SourceDBType }
	BackupDataDir string

	// BatchSize is the threshold (in bytes) after which a batch is written and a new batch is created.
	// Batch sizes are measured using only key and value lengths (as opposed to disk space).
	// Default is 0 (unlimited)
	BatchSize uint

	// ToConvert is all of the DB directories to migrate/convert.
	// Each entry is relative to the data directory.
	ToConvert []string
	// ToCopy is all the non-DB files and directories that should be copied from the source to the new data directory.
	// Each entry is relative to the data directory.
	ToCopy []string

	// Permissions are the permissions to use on any directories created.
	Permissions os.FileMode

	// StatusPeriod is the max time period between status messages.
	// Must be at least 1 second. Default is 5 seconds.
	StatusPeriod time.Duration
	// DirDateFormat is the format string used in dated directory names.
	// Default is "2006-01-02-15-04-05".
	DirDateFormat string

	// TimeStarted is the time that the migration was started.
	// This is set during the call to Migrate.
	TimeStarted time.Time
	// TimeFinished is the time that the migration was finished.
	// This is set during the call to Migrate.
	TimeFinished time.Time
}

// Initialize prepares this Migrator by doing the following:
//  1. Calls ApplyDefaults()
//  2. Checks ValidateBasic()
//  3. Calls ReadSourceDataDir()
func (m *Migrator) Initialize() error {
	m.ApplyDefaults()
	var err error
	if err = m.ValidateBasic(); err != nil {
		return err
	}
	if err = m.ReadSourceDataDir(); err != nil {
		return err
	}
	return nil
}

// ApplyDefaults fills in the defaults that it can, for values that aren't set yet.
func (m *Migrator) ApplyDefaults() {
	if len(m.StagingDir) == 0 && len(m.HomePath) > 0 {
		m.StagingDir = m.HomePath
	}
	if len(m.BackupDir) == 0 && len(m.HomePath) > 0 {
		m.BackupDir = m.HomePath
	}
	if len(m.SourceDataDir) == 0 && len(m.HomePath) > 0 {
		m.SourceDataDir = filepath.Join(m.HomePath, "data")
	}
	if len(m.DirDateFormat) == 0 {
		m.DirDateFormat = "2006-01-02-15-04-05"
	}
	if len(m.StagingDataDir) == 0 && len(m.StagingDir) > 0 {
		m.StagingDataDir = filepath.Join(m.StagingDir, fmt.Sprintf("data-dbmigrate-tmp-%s-%s", time.Now().Format(m.DirDateFormat), m.TargetDBType))
	}
	if len(m.BackupDataDir) == 0 && len(m.BackupDir) > 0 {
		m.BackupDataDir = filepath.Join(m.BackupDir, fmt.Sprintf("data-%s-%s", time.Now().Format(m.DirDateFormat), m.SourceDBType))
	}
	// If we can't source the data directory, we probably can't read it and an error will be returned from something else.
	// For simplicity, we're not really going to care about that error right here, though.
	if m.Permissions == 0 && len(m.SourceDataDir) > 0 {
		sourceDirInfo, err := os.Stat(m.SourceDataDir)
		if err == nil {
			m.Permissions = sourceDirInfo.Mode()
		}
	}
	if m.Permissions == 0 {
		m.Permissions = 0700
	}
	if m.StatusPeriod == 0 {
		m.StatusPeriod = 5 * time.Second
	}
}

// ValidateBasic makes sure that everything is set in this Migrator.
func (m Migrator) ValidateBasic() error {
	if len(m.StagingDir) == 0 {
		return errors.New("no StagingDir defined")
	}
	if len(m.BackupDir) == 0 {
		return errors.New("no BackupDir defined")
	}
	if len(m.SourceDBType) == 0 {
		return errors.New("no SourceDBType defined")
	}
	if len(m.TargetDBType) == 0 {
		return errors.New("no TargetDBType defined")
	}
	if len(m.SourceDataDir) == 0 {
		return errors.New("no SourceDataDir defined")
	}
	if len(m.StagingDataDir) == 0 {
		return errors.New("no StagingDataDir defined")
	}
	if len(m.BackupDataDir) == 0 {
		return errors.New("no BackupDataDir defined")
	}
	if m.Permissions == 0 {
		return errors.New("no Permissions defined")
	}
	if m.StatusPeriod < time.Second {
		return fmt.Errorf("status period %s cannot be less than 1s", m.StatusPeriod)
	}
	if len(m.DirDateFormat) == 0 {
		return errors.New("no DirDateFormat defined")
	}
	return nil
}

// ReadSourceDataDir gets the contents of the SourceDataDir and identifies ToConvert and ToCopy.
// Anything in those two fields prior to calling this, will be overwritten.
//
// Does nothing if SourceDataDir is not set.
func (m *Migrator) ReadSourceDataDir() error {
	if len(m.SourceDataDir) > 0 {
		var err error
		m.ToConvert, m.ToCopy, err = GetDataDirContents(m.SourceDataDir)
		if err != nil {
			return fmt.Errorf("error reading %q: %w", m.SourceDataDir, err)
		}
		if len(m.ToConvert) == 0 {
			return fmt.Errorf("could not identify any db directories in %s", m.SourceDataDir)
		}
	}
	return nil
}

// Migrate converts all database dirs in ToConvert from the source underlying type in the SourceDataDir
// to the target type in the StagingDataDir.
// It then copies everything in ToCopy from the SourceDataDir to the StagingDataDir.
// It then moves the SourceDataDir to BackupDataDir and moves StagingDataDir into place where SourceDataDir was.
func (m *Migrator) Migrate(logger tmlog.Logger) (errRv error) {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	// If this func doesn't complete fully, we want to output a message that the staging directory might still exist (and be quite large).
	// Make a done channel for indicating normal finish and a signal channel for capturing interrupt signals like ctrl+c.
	stagingDirExists := false
	doneChan := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT)
	// If we're returning an error, add the log message, and then always do a little cleanup.
	defer func() {
		if r := recover(); r != nil {
			errRv = fmt.Errorf("recovered from panic: %v", r)
		}
		if stagingDirExists {
			logger.Error("The staging directory still exists due to error.", "dir", m.StagingDataDir)
		}
		close(doneChan)
		signal.Stop(sigChan)
		close(sigChan)
	}()
	// We need to identify the currently running process so that any captured signal can be sent back to it.
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return fmt.Errorf("could not identify the running process: %w", err)
	}
	// Monitor for the signals and handle them appropriately.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("The signal watcher subprocess encountered a panic.", "panic", fmt.Sprintf("%v", r))
			}
		}()
		select {
		case s := <-sigChan:
			signal.Stop(sigChan)
			logger.Error("The staging directory might still due to early termination.", "dir", m.StagingDataDir)
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
	logger.Info(m.MakeSummaryString(nil))
	err = os.MkdirAll(m.StagingDataDir, m.Permissions)
	if err != nil {
		return fmt.Errorf("could not create staging data directory: %w", err)
	}
	stagingDirExists = true
	m.TimeStarted = time.Now()
	logger.Info(fmt.Sprintf("Converting %d Individual DBs.", len(m.ToConvert)),
		"source type", m.SourceDBType, "source", m.SourceDataDir,
		"target type", m.TargetDBType, "staging", m.StagingDataDir,
		"batch size", commaString(m.BatchSize/BytesPerMB),
	)
	counts := map[string]uint{}
	for i, dbDir := range m.ToConvert {
		counts[dbDir], err = m.MigrateDBDir(logger.With("db", strings.TrimSuffix(dbDir, ".db"), "progress", fmt.Sprintf("%d/%d", i+1, len(m.ToConvert))), dbDir)
		if err != nil {
			return fmt.Errorf("could not convert %q from %q to %q: %w", dbDir, m.SourceDBType, m.TargetDBType, err)
		}
	}

	logger.Info(fmt.Sprintf("Copying %d items from %s to %s", len(m.ToCopy), m.SourceDataDir, m.StagingDataDir))
	for i, entry := range m.ToCopy {
		logger.Info(fmt.Sprintf("%d/%d: Copying %s", i+1, len(m.ToCopy), entry))
		if err = copier.Copy(filepath.Join(m.SourceDataDir, entry), filepath.Join(m.StagingDataDir, entry)); err != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err)
		}
	}

	logger.Info("Moving existing data directory to backup location.", "from", m.SourceDataDir, "to", m.BackupDataDir)
	if err = m.MoveWithStatusUpdates(logger, m.SourceDataDir, m.BackupDataDir); err != nil {
		return fmt.Errorf("could not back up existing data directory: %w", err)
	}

	logger.Info("Moving new data directory into place.", "from", m.StagingDataDir, "to", m.SourceDataDir)
	if err = m.MoveWithStatusUpdates(logger, m.StagingDataDir, m.SourceDataDir); err != nil {
		return fmt.Errorf("could not move new data directory into place: %w", err)
	}
	stagingDirExists = false
	m.TimeFinished = time.Now()

	logger.Info(m.MakeSummaryString(counts))
	return nil
}

// MigrateDBDir creates a copy of the given db directory, converting it from one underlying type to another.
func (m Migrator) MigrateDBDir(logger tmlog.Logger, dbDir string) (uint, error) {
	sourceDir, dbName := splitDBPath(m.SourceDataDir, dbDir)
	targetDir, _ := splitDBPath(m.StagingDataDir, dbDir)
	logger.Info("Individual DB Migration: Setting up.", "from", sourceDir, "to", targetDir)

	// Define some counters used in log messages, and a function to make it easy to add them all to log messages.
	var err error
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
			"run time", time.Since(m.TimeStarted).String(),
		}
	}

	// There's several things that need closing and sometimes calling close can cause a segmentation fault that,
	// for some reason, doesn't reach the sigChan notification set up in the Migrate method.
	// So just to have clearer control over closing order, they're all defined at once and closed in a single defer function.
	var sourceDB, targetDB tmdb.DB
	var iter tmdb.Iterator
	var batch tmdb.Batch
	var setupTicker, statusTicker, writeTicker *time.Ticker
	stopTickers := make(chan bool, 1)
	defer func() {
		// closing the stopTickers chan will trigger the status logging subprocess to finish up.
		close(stopTickers)
		// Then we can stop the tickers (just to be safe).
		if setupTicker != nil {
			setupTicker.Stop()
		}
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
	setupTicker = time.NewTicker(m.StatusPeriod)
	writeTicker = time.NewTicker(TickerOff)
	statusTicker = time.NewTicker(TickerOff)
	go func() {
		for {
			select {
			case <-setupTicker.C:
				logger.Info("Still setting up...", commonKeyVals()...)
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
	if targetDir != m.StagingDataDir {
		err = os.MkdirAll(targetDir, m.Permissions)
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
		// Using WriteSync here instead of Write because sometimes the Close was causing a segfault, and maybe this helps?
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

	setupTicker.Reset(TickerOff)
	logger.Info("Individual DB Migration: Starting.")
	batch = targetDB.NewBatch()
	statusTicker.Reset(m.StatusPeriod)
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
			statusTicker.Reset(TickerOff)

			writeTicker.Reset(m.StatusPeriod)
			logger.Info("Writing intermediate batch.", commonKeyVals()...)
			if err = writeAndCloseBatch(); err != nil {
				return writtenEntries, err
			}
			writeTicker.Reset(TickerOff)

			logger.Info("Starting new batch.", commonKeyVals())
			batch = targetDB.NewBatch()
			batchIndex++
			batchBytes = 0
			batchEntries = 0
			statusTicker.Reset(m.StatusPeriod)
		}
	}
	statusTicker.Reset(TickerOff)

	if err = iter.Error(); err != nil {
		return writtenEntries, fmt.Errorf("iterator error: %w", err)
	}

	writeTicker.Reset(m.StatusPeriod)
	logger.Info("Writing final batch.", commonKeyVals()...)
	if err = writeAndCloseBatch(); err != nil {
		return writtenEntries, err
	}
	writeTicker.Reset(TickerOff)

	logger.Info("Individual DB Migration: Done.", "total entries", commaString(writtenEntries))
	return writtenEntries, nil
}

// MoveWithStatusUpdates calls os.Rename but also outputs a log message every StatusPeriod.
func (m Migrator) MoveWithStatusUpdates(logger tmlog.Logger, from, to string) error {
	var moveTicker *time.Ticker
	stopTicker := make(chan bool, 1)
	defer func() {
		// closing the stopTicker chan will trigger the status logging subprocess to finish up.
		close(stopTicker)
		if moveTicker != nil {
			moveTicker.Stop()
		}
	}()
	commonKeyVals := func() []interface{} {
		return []interface{}{
			"from", from,
			"to", to,
			"run time", time.Since(m.TimeStarted).String(),
		}
	}
	moveTicker = time.NewTicker(m.StatusPeriod)
	go func() {
		for {
			select {
			case <-moveTicker.C:
				logger.Info("Still moving...", commonKeyVals()...)
			case <-stopTicker:
				return
			}
		}
	}()
	return os.Rename(from, to)
}

// MakeSummaryString creates a multi-line string with a summary of a migration.
func (m Migrator) MakeSummaryString(counts map[string]uint) string {
	var sb strings.Builder
	addLine := func(format string, a ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, a...) + "\n")
	}
	addLine("Summary:")
	status := "Not Started"
	runTime := time.Duration(0)
	copyHead := " To Copy"
	migrateHead := "  To Migrate"
	switch {
	case !m.TimeFinished.IsZero() && !m.TimeStarted.IsZero():
		status = "Finished"
		runTime = m.TimeFinished.Sub(m.TimeStarted)
		copyHead = "Copied"
		migrateHead = " Migrated"
	case !m.TimeStarted.IsZero():
		status = "Running"
		runTime = time.Since(m.TimeStarted)
		copyHead = "Copying"
		migrateHead = " Migrating"
	}
	addLine("%16s: %s", "Status", status)
	addLine("%16s: %s", "Run Time", runTime)
	addLine("%16s: %s", "Data Dir", m.SourceDataDir)
	addLine("%16s: %s", "Staging Dir", m.StagingDir)
	addLine("%16s: %s", "Backup Dir", m.BackupDataDir)
	addLine("%16s: %s", "Source DB Type", m.SourceDBType)
	addLine("%16s: %s", "New DB Type", m.TargetDBType)
	addLine("%16s: %s", fmt.Sprintf("%s (%d)", copyHead, len(m.ToCopy)), strings.Join(m.ToCopy, "  "))
	if counts == nil {
		addLine("%16s: %s", fmt.Sprintf("%s (%d)", migrateHead, len(m.ToConvert)), strings.Join(m.ToConvert, "  "))
	} else {
		addLine("%16s:", fmt.Sprintf("%s (%d)", migrateHead, len(m.ToConvert)))
		for _, dbDir := range m.ToConvert {
			addLine("%22s: %11s entries", strings.TrimSuffix(dbDir, ".db"), commaString(counts[dbDir]))
		}
	}
	return sb.String()
}

// splitDBPath combine the provided path elements into a full path to a db directory, then
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
