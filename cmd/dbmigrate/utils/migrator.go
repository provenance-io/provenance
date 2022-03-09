package utils

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
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
	// unknownDBBackend is mostly a tmdb.BackendType used in output as a string.
	// It indicates that the backend is unknown.
	unknownDBBackend = tmdb.BackendType("UNKNOWN")
)

var PossibleDBTypes = []string{
	string(tmdb.RocksDBBackend), string(tmdb.BadgerDBBackend),
	string(tmdb.GoLevelDBBackend), string(tmdb.CLevelDBBackend),
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

// Migrator is an object to help guide a migration.
// TargetDBType must be defined. All others can be left unset, or can use defaults from ApplyDefaults().
// If using defaults for any directories, you probably need to set HomePath too though.
type Migrator struct {
	// HomePath is the path to the home directory (should contain the config and data directories).
	HomePath string
	// StagingDir is the directory that will hold the staging data directory.
	// Default is HomePath
	StagingDir string
	// BackupDir is the directory that will hold the backup data directory.
	// Default is HomePath
	BackupDir string

	// TargetDBType is the type of the target (new) DB.
	TargetDBType string

	// SourceDataDir is the path to the source (current) data directory.
	// Default is { HomePath }/data
	SourceDataDir string
	// StagingDataDir is the path to the staging (new) data directory.
	// Default is { StagingDir }/data-dbmigrate-tmp-{timestamp}-{ TargetDBType }
	StagingDataDir string
	// BackupDataDir is the path to where the current data directory will be moved when done.
	// Default is { BackupDir }/data-dbmigrate-backup-{timestamp}
	BackupDataDir string

	// BatchSize is the threshold (in bytes) after which a batch is written and a new batch is created.
	// Batch sizes are measured using only key and value lengths (as opposed to disk space).
	// Default is 0 (unlimited)
	BatchSize uint

	// StageOnly indicates that only the data migration and data copying should happen.
	// If true, the migrator should stop after finishing the staging data directory.
	// That is, it won't move the data dir to the backup location, move the staging directory into place, or update the config.
	StageOnly bool

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
	// Summaries is a map of ToConvert entries, to a short summary string about the migration of that entry.
	// Entries are set during the call to Migrate and are the return values of each MigrateDBDir call.
	Summaries map[string]string
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
		m.BackupDataDir = filepath.Join(m.BackupDir, "data-dbmigrate-backup-"+time.Now().Format(m.DirDateFormat))
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
	if len(m.TargetDBType) == 0 {
		return errors.New("no TargetDBType defined")
	}
	if !IsPossibleDBType(m.TargetDBType) {
		return fmt.Errorf("invalid target type: %q - must be one of: %s", m.TargetDBType, strings.Join(PossibleDBTypes, ", "))
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
	logStagingDirError := false
	doneChan := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT)
	// If we're returning an error, add the log message, and then always do a little cleanup.
	defer func() {
		if r := recover(); r != nil {
			errRv = fmt.Errorf("recovered from panic: %v", r)
		}
		if logStagingDirError {
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
			if logStagingDirError {
				logger.Error("The staging directory still exists due to early termination.", "dir", m.StagingDataDir)
			}
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
	logger.Info(m.MakeSummaryString())
	err = os.MkdirAll(m.StagingDataDir, m.Permissions)
	if err != nil {
		return fmt.Errorf("could not create staging data directory: %w", err)
	}
	logStagingDirError = true
	m.TimeStarted = time.Now()
	logger.Info(fmt.Sprintf("Converting %d Individual DBs.", len(m.ToConvert)),
		"source", m.SourceDataDir, "target type", m.TargetDBType, "staging", m.StagingDataDir,
		"batch size", commaString(m.BatchSize/BytesPerMB),
	)
	m.Summaries = map[string]string{}
	for i, dbDir := range m.ToConvert {
		m.Summaries[dbDir], err = m.MigrateDBDir(logger.With("db", strings.TrimSuffix(dbDir, ".db"), "progress", fmt.Sprintf("%d/%d", i+1, len(m.ToConvert))), dbDir)
		if err != nil {
			return err
		}
	}

	logger.Info(fmt.Sprintf("Copying %d items from %s to %s", len(m.ToCopy), m.SourceDataDir, m.StagingDataDir))
	for i, entry := range m.ToCopy {
		logger.Info(fmt.Sprintf("%d/%d: Copying %s", i+1, len(m.ToCopy), entry), "run time", m.GetRunTime())
		if err = copier.Copy(filepath.Join(m.SourceDataDir, entry), filepath.Join(m.StagingDataDir, entry)); err != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err)
		}
	}

	if m.StageOnly {
		logger.Info("Stage Only flag provided.", "dir", m.StagingDir)
	} else {
		logger.Info("Moving existing data directory to backup location.", "from", m.SourceDataDir, "to", m.BackupDataDir)
		if err = m.MoveWithStatusUpdates(logger, m.SourceDataDir, m.BackupDataDir); err != nil {
			return fmt.Errorf("could not back up existing data directory: %w", err)
		}

		logger.Info("Moving new data directory into place.", "from", m.StagingDataDir, "to", m.SourceDataDir)
		if err = m.MoveWithStatusUpdates(logger, m.StagingDataDir, m.SourceDataDir); err != nil {
			return fmt.Errorf("could not move new data directory into place: %w", err)
		}
	}
	logStagingDirError = false
	m.TimeFinished = time.Now()

	logger.Info(m.MakeSummaryString())
	return nil
}

// MigrateDBDir creates a copy of the given db directory, converting it from one underlying type to another.
func (m Migrator) MigrateDBDir(logger tmlog.Logger, dbDir string) (summary string, err error) {
	summaryError := "error"
	sourceDir, dbName := splitDBPath(m.SourceDataDir, dbDir)
	targetDir, _ := splitDBPath(m.StagingDataDir, dbDir)
	logger.Info("Individual DB Migration: Setting up.", "from", sourceDir, "to", targetDir, "run time", m.GetRunTime().String())

	// Define some counters used in log messages, and a function to make it easy to add them all to log messages.
	writtenEntries := uint(0)
	batchEntries := uint(0)
	batchBytes := uint(0)
	batchIndex := uint(1)
	commonKeyVals := func() []interface{} {
		return []interface{}{
			"batch index", commaString(batchIndex),
			"batch size (megabytes)", commaString(batchBytes / BytesPerMB),
			"batch entries", commaString(batchEntries),
			"db total entries", commaString(writtenEntries + batchEntries),
			"run time", m.GetRunTime().String(),
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
	sourceDBType := unknownDBBackend
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
		// always wrap any error with some extra context.
		if err != nil {
			err = fmt.Errorf("could not convert %q from %q to %q: %w", dbDir, sourceDBType, m.TargetDBType, err)
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

	var ok bool
	sourceDBType, ok = DetectDBType(dbName, sourceDir)
	if !ok {
		return summaryError, fmt.Errorf("could not determine db type: %s", filepath.Join(m.SourceDataDir, dbDir))
	}

	// In at least one case (the snapshots/metadata db), there's a sub-directory that needs to be created in order to
	// safely open a new database in it.
	if targetDir != m.StagingDataDir {
		err = os.MkdirAll(targetDir, m.Permissions)
		if err != nil {
			return summaryError, fmt.Errorf("could not create target sub-directory: %w", err)
		}
	}

	// If they're both the same type, just copy it.
	targetDBBackendType := getBestType(tmdb.BackendType(m.TargetDBType))
	if sourceDBType == targetDBBackendType {
		logger.Info("Source and Target DB Types are the same. Copying instead of migrating.", "db type", m.TargetDBType, "run time", m.GetRunTime().String())
		if err = m.CopyWithStatusUpdates(logger, filepath.Join(m.SourceDataDir, dbDir), filepath.Join(m.StagingDataDir, dbDir)); err != nil {
			return summaryError, fmt.Errorf("could not copy db: %w", err)
		}
		logger.Info("Individual DB Migration: Done.", "run time", m.GetRunTime().String())
		return "Copied", nil
	}

	sourceDB, err = tmdb.NewDB(dbName, sourceDBType, sourceDir)
	if err != nil {
		return summaryError, fmt.Errorf("could not open %q source db: %w", dbName, err)
	}

	targetDB, err = tmdb.NewDB(dbName, targetDBBackendType, targetDir)
	if err != nil {
		return summaryError, fmt.Errorf("could not open %q target db: %w", dbName, err)
	}

	iter, err = sourceDB.Iterator(nil, nil)
	if err != nil {
		return summaryError, fmt.Errorf("could not create %q source iterator: %w", dbName, err)
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
	summaryWrittenEntries := func() string {
		return fmt.Sprintf("Migrated %11s entries", commaString(writtenEntries))
	}

	setupTicker.Reset(TickerOff)
	logger.Info("Individual DB Migration: Starting.", "source db type", sourceDBType, "run time", m.GetRunTime().String())
	batch = targetDB.NewBatch()
	statusTicker.Reset(m.StatusPeriod)
	for ; iter.Valid(); iter.Next() {
		v := iter.Value()
		if v == nil {
			v = []byte{}
		}
		k := iter.Key()
		if err = batch.Set(k, v); err != nil {
			return summaryWrittenEntries(), fmt.Errorf("could not set %q key/value: %w", dbName, err)
		}
		batchEntries++
		batchBytes += uint(len(v) + len(k))
		if m.BatchSize > 0 && batchBytes >= m.BatchSize {
			statusTicker.Reset(TickerOff)

			writeTicker.Reset(m.StatusPeriod)
			logger.Info("Writing intermediate batch.", commonKeyVals()...)
			if err = writeAndCloseBatch(); err != nil {
				return summaryWrittenEntries(), err
			}
			writeTicker.Reset(TickerOff)

			batchIndex++
			batchBytes = 0
			batchEntries = 0
			logger.Info("Starting new batch.", commonKeyVals()...)
			batch = targetDB.NewBatch()
			statusTicker.Reset(m.StatusPeriod)
		}
	}
	statusTicker.Reset(TickerOff)

	if err = iter.Error(); err != nil {
		return summaryWrittenEntries(), fmt.Errorf("iterator error: %w", err)
	}

	writeTicker.Reset(m.StatusPeriod)
	logger.Info("Writing final batch.", commonKeyVals()...)
	if err = writeAndCloseBatch(); err != nil {
		return summaryWrittenEntries(), err
	}
	writeTicker.Reset(TickerOff)

	logger.Info("Individual DB Migration: Done.", "total entries", commaString(writtenEntries), "run time", m.GetRunTime().String())
	return summaryWrittenEntries(), nil
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
			"run time", m.GetRunTime().String(),
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

func (m Migrator) CopyWithStatusUpdates(logger tmlog.Logger, from, to string) error {
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
			"run time", m.GetRunTime().String(),
		}
	}
	moveTicker = time.NewTicker(m.StatusPeriod)
	go func() {
		for {
			select {
			case <-moveTicker.C:
				logger.Info("Still copying...", commonKeyVals()...)
			case <-stopTicker:
				return
			}
		}
	}()
	return copier.Copy(from, to)
}

// MakeSummaryString creates a multi-line string with a summary of a migration.
func (m Migrator) MakeSummaryString() string {
	var sb strings.Builder
	addLine := func(format string, a ...interface{}) {
		sb.WriteString(fmt.Sprintf(format, a...) + "\n")
	}
	addLine("Summary:")
	status := "Not Started"
	copyHead := " To Copy"
	migrateHead := "  To Migrate"
	switch {
	case !m.TimeFinished.IsZero() && !m.TimeStarted.IsZero():
		status = "Finished"
		copyHead = "Copied"
		migrateHead = " Migrated"
	case !m.TimeStarted.IsZero():
		status = "Running"
		copyHead = "Copying"
		migrateHead = " Migrating"
	}
	addLine("%16s: %s", "Status", status)
	addLine("%16s: %s", "Run Time", m.GetRunTime())
	addLine("%16s: %s", "Data Dir", m.SourceDataDir)
	addLine("%16s: %s", "Staging Dir", m.StagingDir)
	addLine("%16s: %s", "Backup Dir", m.BackupDataDir)
	addLine("%16s: %s", "New DB Type", m.TargetDBType)
	addLine("%16s: %s", fmt.Sprintf("%s (%d)", copyHead, len(m.ToCopy)), strings.Join(m.ToCopy, "  "))
	if len(m.Summaries) == 0 {
		addLine("%16s: %s", fmt.Sprintf("%s (%d)", migrateHead, len(m.ToConvert)), strings.Join(m.ToConvert, "  "))
	} else {
		addLine("%16s:", fmt.Sprintf("%s (%d)", migrateHead, len(m.ToConvert)))
		for _, dbDir := range m.ToConvert {
			s, k := m.Summaries[dbDir]
			if !k {
				s = "UNKNOWN"
			}
			addLine("%22s: %s", strings.TrimSuffix(dbDir, ".db"), s)
		}
	}
	return sb.String()
}

func (m Migrator) GetRunTime() time.Duration {
	if m.TimeStarted.IsZero() {
		return 0
	}
	if m.TimeFinished.IsZero() {
		return time.Since(m.TimeStarted)
	}
	return m.TimeFinished.Sub(m.TimeStarted)
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
	contents, err := os.ReadDir(dataDirPath)
	if err != nil {
		return nil, nil, err
	}
	dbs := make([]string, 0)
	nonDBs := make([]string, 0)
	// The db dirs can have a TON of files (10k+). Most of them are just numbers with an extension.
	// This loop short-circuits when it finds a file that starts with "MANIFEST", which is significantly
	// more likely to be closer to the back than the front. So to save lots of iterations, the contents is looped through backwards.
	for i := len(contents) - 1; i >= 0; i-- {
		entry := contents[i]
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
	sort.Strings(dbs)
	sort.Strings(nonDBs)
	return dbs, nonDBs, nil
}

// DetectDBType attempts to identify the type database in the given dir with the given name.
// The name and dir would be the same things that would be provided to tmdb.NewDB.
//
// The return bool indicates whether or not the DB type was identified.
//
// The only types this detects are LevelDB, RocksDB, and BadgerDB.
// If the DB is another type, the behavior of this is unknown.
// There's a chance this will return false, but there's also a chance it is falsely identified as something else.
func DetectDBType(name, dir string) (tmdb.BackendType, bool) {
	// Here are the key differences used to differentiate the DB types.
	// badgerdb:
	// * In a directory named "dir/name".
	// * There are numbered files with the extension ".vlog".
	// * There are numbered files with the extension ".sst" (might possibly be missing if the db is empty).
	// * Has the following files: KEYREGISTRY, MANIFEST
	// * Might also have files: LOCK
	// rocksdb:
	// * In a directory named "dir/name.db".
	// * There are numbered files with the extension ".log".
	// * There are numbered files with the extension ".sst" (might possibly be missing if the db is empty).
	// * Has the following files: CURRENT, IDENTITY, LOG, MANIFEST-{6 digis} (multiple), OPTIONS-{6 digits} (multiple)
	// * Might also have files: LOCK, LOG.old, LOG.old.{16 digits} (multiple)
	// leveldb:
	// * In a directory named "dir/name.db".
	// * There are numbered files with the extension ".log".
	// * There are numbered files with the extension ".ldb" (might possibly be missing if the db is empty).
	// * Has the following files: CURRENT, LOG, MANIFEST-{6 digis} (multiple)
	// * Might also have files: LOCK, LOG.old

	// Note: I'm not sure of an easy way to look for files that start or end with certain strings (e.g files ending in ".sst").
	// The only way I know of is to get the entire dir contents and loop through the entries.
	// However, specially for large DBs, that can be well over 10k files.
	// If the list is sorted (e.g. from os.ReadDir), the ".sst" or ".ldb" files would be one of the first few.
	// And stuff like MANIFEST-{numbers} and OPTIONS-{numbers} would be one of the last few.
	// But just getting that list is a whole lot of work that should be avoided if possible.
	// Additionally, this is only being written with badgerdb, rocksdb, and leveldb in mind.
	//
	// So in here, rather than being more certain and checking for those types of files, we'll skip those checks and
	// put up with possible false positives. Hopefully a false positive would error out at some later point (open or reading).

	// Let's first check for badgerdb since it's the easiest.
	dbDir := filepath.Join(dir, name)
	if dirExists(dbDir) {
		// Since that's a pretty standard dir name, do an easy check for a couple files that should be there.
		if !fileExists(filepath.Join(dbDir, "KEYREGISTRY")) || !fileExists(filepath.Join(dbDir, "MANIFEST")) {
			return unknownDBBackend, false
		}
		// Could also check for a numbered files with the extension ".vlog", but that's expensive.
		// And for the types involved in here, what's been done should be enough to hopefully prevent false positives.
		return tmdb.BadgerDBBackend, true
	}

	// The other two (rocksdb and leveldb) should be in directories named "dir/name.db".
	// and should have files CURRENT and LOG
	dbDir = filepath.Join(dir, name+".db")
	if !dirExists(dbDir) || !fileExists(filepath.Join(dbDir, "CURRENT")) || !fileExists(filepath.Join(dbDir, "LOG")) {
		return unknownDBBackend, false
	}

	// Okay, assuming it's either a rocksdb or leveldb directory now.

	// The only statically named file difference between rocksdb and leveldb is IDENTITY with rocksdb.
	if fileExists(filepath.Join(dbDir, "IDENTITY")) {
		return tmdb.RocksDBBackend, true
	}

	// There are no statically named files that are used by leveldb but not rocksdb. So at this point, just assume it's leveldb.
	// GolevelDB and CLevelDB are both a LevelDB as far as the files go. And CLevelDB is much faster, so use that.
	return tmdb.CLevelDBBackend, true
}

// getBestType returns CLevelDB when provided GoLevelDB. Anything is is returned as is.
// This is because the performance of GoLevelDB is so much worse than CLevelDB.
// We can say that we're using GoLevelDB, but use CLevelDB for all the interactions.
func getBestType(dbType tmdb.BackendType) tmdb.BackendType {
	if dbType == tmdb.GoLevelDBBackend {
		return tmdb.CLevelDBBackend
	}
	return dbType
}

// dirExists returns true if the path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// fileExists returns true if the path exists and is a file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
