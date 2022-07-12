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
	// unknownDBBackend is mostly a tmdb.BackendType used in output as a string.
	// It indicates that the backend is unknown.
	unknownDBBackend = tmdb.BackendType("UNKNOWN")
)

// Note: The PossibleDBTypes variable is a map instead of a slice because trying to append to it was causing one type to
//       stomp out the append from another type (concurrency issue?).

// PossibleDBTypes is a map of strings to BackendTypes representing the Backend types that can be used by this utility.
var PossibleDBTypes = map[string]tmdb.BackendType{}

func init() {
	PossibleDBTypes["goleveldb"] = tmdb.GoLevelDBBackend
}

// AddPossibleDBType adds a possible db backend type.
func AddPossibleDBType(dbType tmdb.BackendType) {
	PossibleDBTypes[string(dbType)] = dbType
}

// GetPossibleDBTypes gets a slice of strings listing all db types that this can use.
func GetPossibleDBTypes() []string {
	rv := make([]string, len(PossibleDBTypes))
	i := 0
	for k := range PossibleDBTypes {
		rv[i] = k
		i++
	}
	sort.Strings(rv)
	return rv
}

// IsPossibleDBType checks if the given dbType string is one that this migrator can handle.
func IsPossibleDBType(dbType string) bool {
	_, ok := PossibleDBTypes[dbType]
	return ok
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
	// Default is { BackupDir }/data-dbmigrate-backup-{timestamp}-{dbtypes}
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
	// Default is to match the source directory, or else 0700.
	Permissions os.FileMode

	// StatusPeriod is the max time period between status messages.
	// Must be at least 1 second. Default is 5 seconds.
	StatusPeriod time.Duration
	// DirDateFormat is the format string used in dated directory names.
	// Default is "2006-01-02-15-04-05".
	DirDateFormat string
}

// migrationManager is a struct with information about the status of a migrator.
type migrationManager struct {
	Migrator

	// Status is a short message about what's currently going on.
	Status string
	// TimeStarted is the time that the migration was started.
	// This is set during the call to Migrate.
	TimeStarted time.Time
	// TimeFinished is the time that the migration was finished.
	// This is set during the call to Migrate.
	TimeFinished time.Time
	// Summaries is a map of ToConvert entries, to a short summary string about the migration of that entry.
	// Entries are set during the call to Migrate and are the return values of each MigrateDBDir call.
	Summaries map[string]string
	// SourceTypes is a map of ToConvert entries, to their backend type.
	SourceTypes map[string]tmdb.BackendType

	// Logger is the Logger to use for logging log messages.
	Logger tmlog.Logger

	// StatusTicker is the ticker used to issue regular status log messages.
	StatusTicker *time.Ticker
	// StopTickerChan is a channel used to stop the regular status log messages.
	StopTickerChan chan bool
	// StatusKeyvals is a function that returns keyvals used in status log messages.
	StatusKeyvals func() []interface{}

	// LogStagingDirError indicates whether or not to log an error about the staging dir existing (for abnormal termination).
	LogStagingDirError bool
	// SigChan is a channel used to Notify certain os signals.
	SigChan chan os.Signal
	// StopSigChan is a channel used to stop the special signal handling.
	StopSigChan chan bool
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
		m.DirDateFormat = "2006-01-02-15-04"
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
			// Mask the Mode to get just the permission bits.
			m.Permissions = sourceDirInfo.Mode() & 0777
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
		return fmt.Errorf("invalid TargetDBType: %q - must be one of: %s", m.TargetDBType, strings.Join(GetPossibleDBTypes(), ", "))
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
		return fmt.Errorf("the StatusPeriod %s cannot be less than 1s", m.StatusPeriod)
	}
	if len(m.DirDateFormat) == 0 {
		return errors.New("no DirDateFormat defined")
	}
	return nil
}

// ReadSourceDataDir gets the contents of the SourceDataDir and populates ToConvert and ToCopy.
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
	defer func() {
		if r := recover(); r != nil {
			errRv = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	if err := m.ValidateBasic(); err != nil {
		return err
	}

	manager, err := m.startMigratorManager(logger)
	if err != nil {
		return err
	}
	defer func() {
		manager.Logger = logger
		manager.Close()
	}()

	// Now we can get started.
	logger.Info(manager.MakeSummaryString())
	manager.Status = "making staging dir"
	err = os.MkdirAll(m.StagingDataDir, m.Permissions)
	if err != nil {
		return fmt.Errorf("could not create staging data directory: %w", err)
	}
	manager.LogStagingDirError = true
	manager.LogWithRunTime(fmt.Sprintf("Converting %d Individual DBs.", len(m.ToConvert)))
	for i, dbDir := range m.ToConvert {
		manager.Logger = logger.With(
			"db", strings.TrimSuffix(dbDir, ".db"),
			"progress", fmt.Sprintf("%d/%d", i+1, len(m.ToConvert)),
		)
		manager.Summaries[dbDir], err = manager.MigrateDBDir(dbDir)
		if err != nil {
			return err
		}
	}
	manager.Logger = logger

	if len(manager.SourceTypes) != 0 {
		m.BackupDataDir = m.BackupDataDir + "-" + strings.Join(manager.GetSourceDBTypes(), "-")
	}

	manager.LogWithRunTime(fmt.Sprintf("Copying %d items.", len(m.ToCopy)))
	for i, entry := range m.ToCopy {
		manager.LogWithRunTime(fmt.Sprintf("%d/%d: Copying %s", i+1, len(m.ToCopy), entry))
		if err = copier.Copy(filepath.Join(m.SourceDataDir, entry), filepath.Join(m.StagingDataDir, entry)); err != nil {
			return fmt.Errorf("could not copy %s: %w", entry, err)
		}
	}

	if m.StageOnly {
		manager.LogWithRunTime("Stage Only flag provided.", "dir", m.StagingDir)
	} else {
		manager.Status = "moving old data dir"
		manager.StatusKeyvals = func() []interface{} {
			return []interface{}{
				"from", m.SourceDataDir,
				"to", m.BackupDataDir,
			}
		}
		manager.LogWithRunTime("Moving existing data directory to backup location.", manager.StatusKeyvals()...)
		if err = os.Rename(m.SourceDataDir, m.BackupDataDir); err != nil {
			return fmt.Errorf("could not back up existing data directory: %w", err)
		}

		manager.Status = "moving new data dir"
		manager.StatusKeyvals = func() []interface{} {
			return []interface{}{
				"from", m.StagingDataDir,
				"to", m.SourceDataDir,
			}
		}
		manager.LogWithRunTime("Moving new data directory into place.", manager.StatusKeyvals()...)
		if err = os.Rename(m.StagingDataDir, m.SourceDataDir); err != nil {
			return fmt.Errorf("could not move new data directory into place: %w", err)
		}
		manager.StatusKeyvals = noKeyvals
	}
	manager.LogStagingDirError = false
	manager.Finish()

	logger.Info(manager.MakeSummaryString())
	return nil
}

// startMigratorManager creates a migrationManager and initializes it.
// It must later be closed.
func (m Migrator) startMigratorManager(logger tmlog.Logger) (*migrationManager, error) {
	rv := &migrationManager{
		Migrator:       m,
		Status:         "starting",
		TimeStarted:    time.Now(),
		Summaries:      map[string]string{},
		SourceTypes:    map[string]tmdb.BackendType{},
		Logger:         logger,
		StatusTicker:   time.NewTicker(m.StatusPeriod),
		StatusKeyvals:  noKeyvals,
		StopTickerChan: make(chan bool, 1),
		SigChan:        make(chan os.Signal, 1),
		StopSigChan:    make(chan bool, 1),
	}
	// Monitor for the signals and handle them appropriately.
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return nil, fmt.Errorf("could not identify the running process: %w", err)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				rv.LogErrorWithRunTime("The signal watcher subprocess encountered a panic.", "panic", fmt.Sprintf("%v", r))
			}
		}()
		select {
		case s := <-rv.SigChan:
			signal.Stop(rv.SigChan)
			if rv.LogStagingDirError {
				rv.LogErrorWithRunTime("The staging directory still exists due to early termination.", "dir", m.StagingDataDir)
			}
			err2 := proc.Signal(s)
			if err2 != nil {
				rv.LogErrorWithRunTime("Error propagating signal.", "error", err2)
			}
			return
		case <-rv.StopSigChan:
			return
		}
	}()
	signal.Notify(rv.SigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT)
	// Fire up another sub-process for outputting a status message every now and then.
	rv.StatusTicker = time.NewTicker(m.StatusPeriod)
	go func() {
		for {
			select {
			case <-rv.StatusTicker.C:
				rv.LogWithRunTime(fmt.Sprintf("Status: %s", rv.Status), rv.StatusKeyvals()...)
			case <-rv.StopTickerChan:
				return
			}
		}
	}()
	return rv, nil
}

// MigrateDBDir creates a copy of the given db directory, converting it from one underlying type to another.
func (m *migrationManager) MigrateDBDir(dbDir string) (summary string, err error) {
	m.LogWithRunTime("Individual DB Migration: Setting up.")
	m.Status = "setting up"

	summaryError := "error"
	sourceDir, dbName := splitDBPath(m.SourceDataDir, dbDir)
	targetDir, _ := splitDBPath(m.StagingDataDir, dbDir)

	// Define some counters used in log messages, and a function to make it easy to add them all to log messages.
	writtenEntries := uint(0)
	batchEntries := uint(0)
	batchBytes := uint(0)
	batchIndex := uint(1)
	m.StatusKeyvals = func() []interface{} {
		return []interface{}{
			"batch index", commaString(batchIndex),
			"batch size (megabytes)", commaString(batchBytes / BytesPerMB),
			"batch entries", commaString(batchEntries),
			"db total entries", commaString(writtenEntries + batchEntries),
		}
	}
	logWithStats := func(msg string, keyvals ...interface{}) {
		m.LogWithRunTime(msg, append(m.StatusKeyvals(), keyvals...)...)
	}

	// There's several things that need closing and sometimes calling close can cause a segmentation fault that,
	// for some reason, doesn't reach the SigChan notification set up in the Migrate method.
	// So just to have clearer control over closing order, they're all defined at once and closed in a single defer function.
	var sourceDB, targetDB tmdb.DB
	var iter tmdb.Iterator
	var batch tmdb.Batch
	sourceDBType := unknownDBBackend
	defer func() {
		m.StatusKeyvals = noKeyvals
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

	m.Status = "detecting db type"
	var ok bool
	sourceDBType, ok = DetectDBType(dbName, sourceDir)
	if !ok {
		return summaryError, fmt.Errorf("could not determine db type: %s", filepath.Join(m.SourceDataDir, dbDir))
	}
	if !IsPossibleDBType(string(sourceDBType)) {
		return summaryError, fmt.Errorf("cannot read source db of type %q", sourceDBType)
	}
	m.SourceTypes[dbDir] = sourceDBType

	// In at least one case (the snapshots/metadata db), there's a sub-directory that needs to be created in order to
	// safely open a new database in it.
	if targetDir != m.StagingDataDir {
		m.Status = "making sub-dir"
		err = os.MkdirAll(targetDir, m.Permissions)
		if err != nil {
			return summaryError, fmt.Errorf("could not create target sub-directory: %w", err)
		}
	}

	// If they're both the same type, just copy it and be done.
	targetDBBackendType := tmdb.BackendType(m.TargetDBType)
	if sourceDBType == targetDBBackendType {
		m.Status = "copying db"
		from := filepath.Join(m.SourceDataDir, dbDir)
		to := filepath.Join(m.StagingDataDir, dbDir)
		m.StatusKeyvals = func() []interface{} {
			return []interface{}{
				"from", from,
				"to", to,
			}
		}
		m.LogWithRunTime("Source and Target DB Types are the same. Copying instead of migrating.", "db type", m.TargetDBType)
		if err = copier.Copy(from, to); err != nil {
			return summaryError, fmt.Errorf("could not copy db: %w", err)
		}
		m.Status = "done"
		m.LogWithRunTime("Individual DB Migration: Done.")
		return "Copied", nil
	}

	m.Status = "opening source db"
	sourceDB, err = tmdb.NewDB(dbName, sourceDBType, sourceDir)
	if err != nil {
		return summaryError, fmt.Errorf("could not open %q source db: %w", dbName, err)
	}

	m.Status = "opening target db"
	targetDB, err = tmdb.NewDB(dbName, targetDBBackendType, targetDir)
	if err != nil {
		return summaryError, fmt.Errorf("could not open %q target db: %w", dbName, err)
	}

	m.Status = "making iterator"
	iter, err = sourceDB.Iterator(nil, nil)
	if err != nil {
		return summaryError, fmt.Errorf("could not create %q source iterator: %w", dbName, err)
	}

	// There's a couple places in here where we need to write and close the batch. But the safety stuff (on errors)
	// is needed in both places, so it's pulled out into this anonymous function.
	writeAndCloseBatch := func() error {
		m.Status = "writing batch"
		// Using WriteSync here instead of Write because sometimes the Close was causing a segfault, and maybe this helps?
		if err = batch.WriteSync(); err != nil {
			// If the write fails, closing the db can sometimes cause a segmentation fault.
			targetDB = nil
			return fmt.Errorf("could not write %q batch: %w", dbName, err)
		}
		writtenEntries += batchEntries
		m.Status = "closing batch"
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
		// 13 characters accounts right-justifies the numbers as long as they're under 10 billion (9,999,999,999).
		// Testnet's application db (as of writing this) has just over 1 billion entries.
		return fmt.Sprintf("Migrated %13s entries from %s to %s.", commaString(writtenEntries), sourceDBType, m.TargetDBType)
	}

	m.LogWithRunTime("Individual DB Migration: Starting.", "source db type", sourceDBType)
	batch = targetDB.NewBatch()
	m.Status = "starting iteration"
	for ; iter.Valid(); iter.Next() {
		m.Status = "getting entry key"
		k := iter.Key()
		m.Status = "getting entry value"
		v := iter.Value()
		if v == nil {
			v = []byte{}
		}
		m.Status = "adding entry to batch"
		if err = batch.Set(k, v); err != nil {
			return summaryWrittenEntries(), fmt.Errorf("could not set %q key/value: %w", dbName, err)
		}
		m.Status = "counting"
		batchEntries++
		batchBytes += uint(len(v) + len(k))
		if m.BatchSize > 0 && batchBytes >= m.BatchSize {
			logWithStats("Writing intermediate batch.")
			if err = writeAndCloseBatch(); err != nil {
				return summaryWrittenEntries(), err
			}

			m.Status = "batch reset"
			batchIndex++
			batchBytes = 0
			batchEntries = 0
			logWithStats("Starting new batch.")
			batch = targetDB.NewBatch()
		}
		m.Status = "getting next entry"
	}

	m.Status = "done iterating"
	if err = iter.Error(); err != nil {
		return summaryWrittenEntries(), fmt.Errorf("iterator error: %w", err)
	}

	logWithStats("Writing final batch.")
	if err = writeAndCloseBatch(); err != nil {
		return summaryWrittenEntries(), err
	}

	m.Status = "done"
	m.LogWithRunTime("Individual DB Migration: Done.", "total entries", commaString(writtenEntries))
	return summaryWrittenEntries(), nil
}

func (m *migrationManager) Finish() {
	m.StopSigChan <- true
	m.StopTickerChan <- true
	m.StatusTicker.Stop()
	m.TimeFinished = time.Now()
}

// Close closes up shop on a migrationManager.
func (m *migrationManager) Close() {
	m.StatusKeyvals = noKeyvals
	close(m.StopSigChan)
	signal.Stop(m.SigChan)
	close(m.SigChan)
	if m.LogStagingDirError {
		m.LogErrorWithRunTime("The staging directory still exists.", "dir", m.StagingDataDir)
	}
}

// MakeSummaryString creates a multi-line string with a summary of a migration.
func (m migrationManager) MakeSummaryString() string {
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
	if m.StageOnly {
		addLine("%16s: %s", "Staging Only", "true")
	} else {
		addLine("%16s: %s", "Backup Dir", m.BackupDataDir)
	}
	addLine("%16s: %s megabytes", "Batch Size", commaString(m.BatchSize/BytesPerMB))
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

// GetRunTime gets a string of the run time of this manager.
// Output is a time.Duration string, e.g. "25m36.910647946s"
func (m migrationManager) GetRunTime() string {
	if m.TimeStarted.IsZero() {
		return "0.000000000s"
	}
	if m.TimeFinished.IsZero() {
		return time.Since(m.TimeStarted).String()
	}
	return m.TimeFinished.Sub(m.TimeStarted).String()
}

// LogWithRunTime is a wrapper on Logger.Info that always includes the run time.
func (m migrationManager) LogWithRunTime(msg string, keyvals ...interface{}) {
	m.Logger.Info(msg, append(keyvals, "run time", m.GetRunTime())...)
}

// LogErrorWithRunTime is a wrapper on Logger.Error that always includes the run time.
func (m migrationManager) LogErrorWithRunTime(msg string, keyvals ...interface{}) {
	m.Logger.Error(msg, append(keyvals, "run time", m.GetRunTime())...)
}

func (m migrationManager) GetSourceDBTypes() []string {
	rv := []string{}
	for _, dbType := range m.SourceTypes {
		found := false
		for _, v := range rv {
			if v == string(dbType) {
				found = true
				break
			}
		}
		if !found {
			rv = append(rv, string(dbType))
		}
	}
	sort.Strings(rv)
	return rv
}

// noKeyvals returns an empty slice. It's handy for setting migrationManager.StatusKeyvals
func noKeyvals() []interface{} {
	return []interface{}{}
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
		case filepath.Ext(entry.Name()) == ".db":
			// boltdb has executable files with a .db extension.
			info, err := entry.Info()
			if err != nil {
				return nil, nil, err
			}
			// Check if the file has at least one executable bit set.
			if info.Mode()&0111 != 1 {
				dbs = append(dbs, entry.Name())
			} else {
				nonDBs = append(nonDBs, entry.Name())
			}
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
	// * Has the following files: CURRENT, IDENTITY, LOG, MANIFEST-{6 digits} (multiple), OPTIONS-{6 digits} (multiple)
	// * Might also have files: LOCK, LOG.old, LOG.old.{16 digits} (multiple)
	// leveldb:
	// * In a directory named "dir/name.db".
	// * There are numbered files with the extension ".log".
	// * There are numbered files with the extension ".ldb" (might possibly be missing if the db is empty).
	// * Has the following files: CURRENT, LOG, MANIFEST-{6 digits} (multiple)
	// * Might also have files: LOCK, LOG.old
	// boltdb:
	// * Is an executable file named "dir/name.db"

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

	// Now lets check for boltdb. It's a file instead of directory with the same name used by rocksdb and leveldb.
	dbDir = filepath.Join(dir, name+".db")
	if fileExists(dbDir) {
		return tmdb.BoltDBBackend, true
	}

	// The other two (rocksdb and leveldb) should be in directories named "dir/name.db".
	// and should have files CURRENT and LOG
	if !dirExists(dbDir) || !fileExists(filepath.Join(dbDir, "CURRENT")) || !fileExists(filepath.Join(dbDir, "LOG")) {
		return unknownDBBackend, false
	}

	// Okay, assuming it's either a rocksdb or leveldb directory now.

	// The only statically named file difference between rocksdb and leveldb is IDENTITY with rocksdb.
	if fileExists(filepath.Join(dbDir, "IDENTITY")) {
		return tmdb.RocksDBBackend, true
	}

	// At this point, we assume it's either cleveldb or goleveldb.
	// Unfortunately, they both use the same files, but possibly with different formats.
	// Sometimes you can treat a goleveldb as cleveldb and vice versa, but sometimes you can't.
	// The only way I can think of to differentiate them here is to just try to open them.
	// I didn't test like this with the other types because the tmdb.NewDB function will create
	// a db if it doesn't exist which can cause weird behavior if trying with the wrong db type.
	// Goleveldb and cleveldb are close enough, though that it won't cause problems.
	canOpenDB := func(backend tmdb.BackendType) (rv bool) {
		defer func() {
			if r := recover(); r != nil {
				rv = false
			}
		}()
		db, err := tmdb.NewDB(name, backend, dir)
		if err != nil {
			return false
		}
		iter, err := db.Iterator(nil, nil)
		if err != nil {
			return false
		}
		// Check up to the first 10 entries. 10 randomly picked to not be t0o big, but bigger than 0 or 1.
		i := 0
		for ; iter.Valid(); iter.Next() {
			_ = iter.Key()
			_ = iter.Value()
			i++
			if i >= 10 {
				break
			}
		}
		if iter.Error() != nil {
			return false
		}
		if iter.Close() != nil {
			return false
		}
		if db.Close() != nil {
			return false
		}
		return true
	}

	if canOpenDB(tmdb.GoLevelDBBackend) {
		return tmdb.GoLevelDBBackend, true
	}
	if IsPossibleDBType(string(tmdb.CLevelDBBackend)) && canOpenDB(tmdb.CLevelDBBackend) {
		return tmdb.CLevelDBBackend, true
	}

	return unknownDBBackend, false
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
