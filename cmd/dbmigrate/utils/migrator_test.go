package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tmdb "github.com/tendermint/tm-db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MigratorTestSuite struct {
	suite.Suite
}

func (s MigratorTestSuite) SetupTest() {

}

func TestMigratorTestSuite(t *testing.T) {
	suite.Run(t, new(MigratorTestSuite))
}

func (s MigratorTestSuite) TestInitialize() {
	tdir := s.T().TempDir()
	dbdir := "some.db"
	someFile := "somefile.txt"
	s.Require().NoError(os.MkdirAll(filepath.Join(tdir, "data", dbdir), 0700), "making dbdir")
	s.Require().NoError(os.WriteFile(filepath.Join(tdir, "data", someFile), []byte{}, 0600), "making somefile")

	s.T().Run("ApplyDefaults called before ValidateBasic", func(t *testing.T) {
		m := &Migrator{
			TargetDBType: "", // Will cause error.
			HomePath:     tdir,
		}
		err := m.Initialize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TargetDBType")
		assert.Equal(t, m.SourceDataDir, filepath.Join(tdir, "data"))
	})

	s.T().Run("ReadSourceDataDir not called if ValidateBasic gives error", func(t *testing.T) {
		m := &Migrator{
			TargetDBType: "", // Will cause error.
			HomePath:     tdir,
		}
		err := m.Initialize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TargetDBType", "err")
		assert.Len(t, m.ToConvert, 0, "ToConvert")
		assert.Len(t, m.ToCopy, 0, "ToCopy")
	})

	s.T().Run("ReadSourceDataDir called if valid", func(t *testing.T) {
		m := &Migrator{
			TargetDBType: "goleveldb",
			HomePath:     tdir,
		}
		err := m.Initialize()
		require.NoError(t, err)
		assert.Len(t, m.ToConvert, 1, "ToConvert")
		assert.Contains(t, m.ToConvert, dbdir, "ToConvert")
		assert.Len(t, m.ToCopy, 1, "ToCopy")
		assert.Contains(t, m.ToCopy, someFile, "ToCopy")
	})
}

func (s MigratorTestSuite) TestApplyDefaults() {
	defaultDateFormat := "2006-01-02-15-04"
	tdir := s.T().TempDir()
	dirForPermTest := filepath.Join(tdir, "permissions-test")
	permForPermTest := os.FileMode(0750)
	os.MkdirAll(dirForPermTest, permForPermTest)
	var tests = []struct {
		name     string
		migrator *Migrator
		getter   func(m *Migrator) interface{}
		expected interface{}
	}{
		{
			name: "staging dir empty home path empty",
			migrator: &Migrator{
				HomePath:   "",
				StagingDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDir },
			expected: "",
		},
		{
			name: "staging dir empty home path not empty",
			migrator: &Migrator{
				HomePath:   "homepath",
				StagingDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDir },
			expected: "homepath",
		},
		{
			name: "staging dir not empty home path not",
			migrator: &Migrator{
				HomePath:   "",
				StagingDir: "stagingdir",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDir },
			expected: "stagingdir",
		},
		{
			name: "staging dir not empty home path not empty",
			migrator: &Migrator{
				HomePath:   "homepath",
				StagingDir: "stagingdir",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDir },
			expected: "stagingdir",
		},

		{
			name: "backup dir empty home path empty",
			migrator: &Migrator{
				HomePath:  "",
				BackupDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDir },
			expected: "",
		},
		{
			name: "backup dir empty home path not empty",
			migrator: &Migrator{
				HomePath:  "homepath",
				BackupDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDir },
			expected: "homepath",
		},
		{
			name: "backup dir not empty home path not",
			migrator: &Migrator{
				HomePath:  "",
				BackupDir: "backupdir",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDir },
			expected: "backupdir",
		},
		{
			name: "backup dir not empty home path not empty",
			migrator: &Migrator{
				HomePath:  "homepath",
				BackupDir: "backupdir",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDir },
			expected: "backupdir",
		},

		{
			name: "source data dir empty home path empty",
			migrator: &Migrator{
				HomePath:      "",
				SourceDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.SourceDataDir },
			expected: "",
		},
		{
			name: "source data dir empty home path not empty",
			migrator: &Migrator{
				HomePath:      "homepath",
				SourceDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.SourceDataDir },
			expected: filepath.Join("homepath", "data"),
		},
		{
			name: "source data dir not empty home path not",
			migrator: &Migrator{
				HomePath:      "",
				SourceDataDir: "sourcedatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.SourceDataDir },
			expected: "sourcedatadir",
		},
		{
			name: "source data dir not empty home path not empty",
			migrator: &Migrator{
				HomePath:      "homepath",
				SourceDataDir: "sourcedatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.SourceDataDir },
			expected: "sourcedatadir",
		},

		{
			name: "dir date format empty",
			migrator: &Migrator{
				DirDateFormat: "",
			},
			getter:   func(m *Migrator) interface{} { return m.DirDateFormat },
			expected: defaultDateFormat,
		},
		{
			name: "dir date format not empty",
			migrator: &Migrator{
				DirDateFormat: "04-15-02-01-2006",
			},
			getter:   func(m *Migrator) interface{} { return m.DirDateFormat },
			expected: "04-15-02-01-2006",
		},

		{
			name: "staging data dir empty staging dir empty",
			migrator: &Migrator{
				StagingDir:     "",
				StagingDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDataDir },
			expected: "",
		},
		{
			name: "staging data dir empty staging dir not empty",
			migrator: &Migrator{
				TargetDBType:   "targetdb",
				StagingDir:     "stagingdir",
				StagingDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDataDir },
			expected: filepath.Join("stagingdir", fmt.Sprintf("data-dbmigrate-tmp-%s-%s", time.Now().Format(defaultDateFormat), "targetdb")),
		},
		{
			name: "staging data dir not empty staging dir empty",
			migrator: &Migrator{
				StagingDir:     "",
				StagingDataDir: "stagingdatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDataDir },
			expected: "stagingdatadir",
		},
		{
			name: "staging data dir not empty staging dir empty not",
			migrator: &Migrator{
				StagingDir:     "homepath",
				StagingDataDir: "stagingdatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.StagingDataDir },
			expected: "stagingdatadir",
		},

		{
			name: "backup data dir empty staging dir empty",
			migrator: &Migrator{
				BackupDir:     "",
				BackupDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDataDir },
			expected: "",
		},
		{
			name: "backup data dir empty staging dir not empty",
			migrator: &Migrator{
				BackupDir:     "backupdir",
				BackupDataDir: "",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDataDir },
			expected: filepath.Join("backupdir", "data-dbmigrate-backup-"+time.Now().Format(defaultDateFormat)),
		},
		{
			name: "backup data dir not empty staging dir empty",
			migrator: &Migrator{
				BackupDir:     "",
				BackupDataDir: "backupdatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDataDir },
			expected: "backupdatadir",
		},
		{
			name: "backup data dir not empty staging dir empty not",
			migrator: &Migrator{
				BackupDir:     "homepath",
				BackupDataDir: "backupdatadir",
			},
			getter:   func(m *Migrator) interface{} { return m.BackupDataDir },
			expected: "backupdatadir",
		},

		{
			name: "permissions not set source data dir does not exist",
			migrator: &Migrator{
				Permissions:   0,
				SourceDataDir: "this-definitely-does-not-exist",
			},
			getter:   func(m *Migrator) interface{} { return m.Permissions },
			expected: os.FileMode(0700),
		},
		{
			name: "permissions not set source data dir exists",
			migrator: &Migrator{
				Permissions:   0,
				SourceDataDir: dirForPermTest,
			},
			getter:   func(m *Migrator) interface{} { return m.Permissions },
			expected: permForPermTest,
		},
		{
			name: "permissions set source data dir does not exist",
			migrator: &Migrator{
				Permissions:   0777,
				SourceDataDir: "this-definitely-does-not-exist",
			},
			getter:   func(m *Migrator) interface{} { return m.Permissions },
			expected: os.FileMode(0777),
		},
		{
			name: "permissions set source data dir exists",
			migrator: &Migrator{
				Permissions:   0775,
				SourceDataDir: dirForPermTest,
			},
			getter:   func(m *Migrator) interface{} { return m.Permissions },
			expected: os.FileMode(0775),
		},

		{
			name: "status period not set",
			migrator: &Migrator{
				StatusPeriod: 0,
			},
			getter:   func(m *Migrator) interface{} { return m.StatusPeriod },
			expected: 5 * time.Second,
		},
		{
			name: "status period set",
			migrator: &Migrator{
				StatusPeriod: 10 * time.Second,
			},
			getter:   func(m *Migrator) interface{} { return m.StatusPeriod },
			expected: 10 * time.Second,
		},

		{
			name: "target db type not set unchanged",
			migrator: &Migrator{
				TargetDBType: "",
			},
			getter:   func(m *Migrator) interface{} { return m.TargetDBType },
			expected: "",
		},
		{
			name: "target db type set unchanged",
			migrator: &Migrator{
				TargetDBType: "target type",
			},
			getter:   func(m *Migrator) interface{} { return m.TargetDBType },
			expected: "target type",
		},

		{
			name: "batch size not set unchanged",
			migrator: &Migrator{
				BatchSize: 0,
			},
			getter:   func(m *Migrator) interface{} { return m.BatchSize },
			expected: uint(0),
		},
		{
			name: "batch size set unchanged",
			migrator: &Migrator{
				BatchSize: 1234,
			},
			getter:   func(m *Migrator) interface{} { return m.BatchSize },
			expected: uint(1234),
		},

		{
			name: "to convert not set unchanged",
			migrator: &Migrator{
				ToConvert: nil,
			},
			getter:   func(m *Migrator) interface{} { return m.ToConvert },
			expected: []string(nil),
		},
		{
			name: "to convert set unchanged",
			migrator: &Migrator{
				ToConvert: []string{"foo"},
			},
			getter:   func(m *Migrator) interface{} { return m.ToConvert },
			expected: []string{"foo"},
		},

		{
			name: "to copy not set unchanged",
			migrator: &Migrator{
				ToCopy: nil,
			},
			getter:   func(m *Migrator) interface{} { return m.ToCopy },
			expected: []string(nil),
		},
		{
			name: "to copy set unchanged",
			migrator: &Migrator{
				ToCopy: []string{"bar"},
			},
			getter:   func(m *Migrator) interface{} { return m.ToCopy },
			expected: []string{"bar"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.migrator.ApplyDefaults()
			actual := tc.getter(tc.migrator)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func (s MigratorTestSuite) TestValidateBasic() {
	makeValidMigrator := func() *Migrator {
		rv := &Migrator{
			HomePath:     "testing",
			TargetDBType: "goleveldb",
		}
		rv.ApplyDefaults()
		return rv
	}
	tests := []struct {
		name       string
		modifier   func(m *Migrator)
		expInError []string
	}{
		{
			name:       "all valid",
			modifier:   func(m *Migrator) {},
			expInError: nil,
		},
		{
			name:       "StagingDir empty",
			modifier:   func(m *Migrator) { m.StagingDir = "" },
			expInError: []string{"StagingDir"},
		},
		{
			name:       "BackupDir empty",
			modifier:   func(m *Migrator) { m.BackupDir = "" },
			expInError: []string{"BackupDir"},
		},
		{
			name:       "TargetDBType empty",
			modifier:   func(m *Migrator) { m.TargetDBType = "" },
			expInError: []string{"TargetDBType"},
		},
		{
			name:       "TargetDBType not possible",
			modifier:   func(m *Migrator) { m.TargetDBType = "not-possible" },
			expInError: []string{"TargetDBType", "goleveldb", "\"not-possible\""},
		},
		{
			name:       "SourceDataDir empty",
			modifier:   func(m *Migrator) { m.SourceDataDir = "" },
			expInError: []string{"SourceDataDir"},
		},
		{
			name:       "StagingDataDir empty",
			modifier:   func(m *Migrator) { m.StagingDataDir = "" },
			expInError: []string{"StagingDataDir"},
		},
		{
			name:       "BackupDataDir empty",
			modifier:   func(m *Migrator) { m.BackupDataDir = "" },
			expInError: []string{"BackupDataDir"},
		},
		{
			name:       "Permissions empty",
			modifier:   func(m *Migrator) { m.Permissions = 0 },
			expInError: []string{"Permissions"},
		},
		{
			name:       "StatusPeriod empty",
			modifier:   func(m *Migrator) { m.StatusPeriod = 0 },
			expInError: []string{"StatusPeriod"},
		},
		{
			name:       "StatusPeriod just under 1 second",
			modifier:   func(m *Migrator) { m.StatusPeriod = time.Second - time.Nanosecond },
			expInError: []string{"StatusPeriod", "999.999999ms", "1s"},
		},
		{
			name:       "DirDateFormat empty",
			modifier:   func(m *Migrator) { m.DirDateFormat = "" },
			expInError: []string{"DirDateFormat"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			m := makeValidMigrator()
			tc.modifier(m)
			actual := m.ValidateBasic()
			if len(tc.expInError) > 0 {
				require.Error(t, actual)
				for _, exp := range tc.expInError {
					assert.Contains(t, actual.Error(), exp)
				}
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

func (s MigratorTestSuite) TestReadSourceDataDir() {

	s.T().Run("no source data dir", func(t *testing.T) {
		m := &Migrator{
			SourceDataDir: "",
			ToConvert:     []string{"something"},
			ToCopy:        []string{"anotherthing"},
		}
		err := m.ReadSourceDataDir()
		// It shouldn't give an error.
		require.NoError(t, err)
		// And the ToConvert and ToCopy slices shouldn't have changed.
		assert.Len(t, m.ToConvert, 1, "ToConvert")
		assert.Contains(t, m.ToConvert, "something", "ToConvert")
		assert.Len(t, m.ToCopy, 1, "ToCopy")
		assert.Contains(t, m.ToCopy, "anotherthing", "ToCopy")
	})

	s.T().Run("source data dir does not exist", func(t *testing.T) {
		m := &Migrator{
			SourceDataDir: "not-gonna-find-me",
			ToConvert:     []string{"something"},
			ToCopy:        []string{"anotherthing"},
		}
		err := m.ReadSourceDataDir()
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading \"not-gonna-find-me\":", "err")
		// And the ToConvert and ToCopy slices should be gone.
		assert.Len(t, m.ToConvert, 0, "ToConvert")
		assert.Len(t, m.ToCopy, 0, "ToCopy")
	})

	s.T().Run("source data dir has a file but no db", func(t *testing.T) {
		tdir := t.TempDir()
		someFile := "somefile.txt"
		dataDir := filepath.Join(tdir, "data")
		require.NoError(t, os.MkdirAll(dataDir, 0700), "making dbdir")
		require.NoError(t, os.WriteFile(filepath.Join(dataDir, someFile), []byte{}, 0600), "making somefile")
		m := &Migrator{
			SourceDataDir: dataDir,
			ToConvert:     []string{"something"},
			ToCopy:        []string{"anotherthing"},
		}
		err := m.ReadSourceDataDir()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not identify any db directories in")
		assert.Contains(t, err.Error(), dataDir)
		// And the ToConvert and ToCopy slices should be changed.
		assert.Len(t, m.ToConvert, 0, "ToConvert")
		assert.Len(t, m.ToCopy, 1, "ToCopy")
		assert.Contains(t, m.ToCopy, someFile, "ToCopy")
	})

	s.T().Run("source data dir has a db", func(t *testing.T) {
		tdir := s.T().TempDir()
		dbdir := "some.db"
		someFile := "somefile.txt"
		dataDir := filepath.Join(tdir, "data")
		require.NoError(t, os.MkdirAll(filepath.Join(dataDir, dbdir), 0700), "making dbdir")
		require.NoError(t, os.WriteFile(filepath.Join(dataDir, someFile), []byte{}, 0600), "making somefile")
		m := &Migrator{
			SourceDataDir: dataDir,
			ToConvert:     []string{"something"},
			ToCopy:        []string{"anotherthing"},
		}
		err := m.ReadSourceDataDir()
		require.NoError(t, err)
		// And the ToConvert and ToCopy slices should be changed.
		assert.Len(t, m.ToConvert, 1, "ToConvert")
		assert.Contains(t, m.ToConvert, dbdir, "ToConvert")
		assert.Len(t, m.ToCopy, 1, "ToCopy")
		assert.Contains(t, m.ToCopy, someFile, "ToCopy")
	})
}

// TODO: Migrate tests
// TODO: migrationManager tests

func (s MigratorTestSuite) TestNoKeyvals() {
	f := noKeyvals()
	s.Require().NotNil(f)
	s.Assert().Len(f, 0)
}

func (s MigratorTestSuite) TestSplitDBPath() {
	tests := []struct {
		name   string
		elem   []string
		dbPath string
		dbName string
	}{
		{
			name:   "absolute path and simple db name",
			elem:   []string{"/foo/bar", "baz.db"},
			dbPath: "/foo/bar",
			dbName: "baz",
		},
		{
			name:   "absolute path and simple db name no suffix",
			elem:   []string{"/foo/bar", "baz"},
			dbPath: "/foo/bar",
			dbName: "baz",
		},
		{
			name:   "absolute path and simple db name weird suffix",
			elem:   []string{"/foo/bar", "baz.db2"},
			dbPath: "/foo/bar",
			dbName: "baz.db2",
		},
		{
			name:   "absolute path and db in sub dir",
			elem:   []string{"/foo", "bar/baz.db"},
			dbPath: "/foo/bar",
			dbName: "baz",
		},
		{
			name:   "absolute path and db in sub dir no suffix",
			elem:   []string{"/foo", "bar/baz"},
			dbPath: "/foo/bar",
			dbName: "baz",
		},
		{
			name:   "absolute path and db in sub dir weird suffix",
			elem:   []string{"/foo", "bar/baz.db2"},
			dbPath: "/foo/bar",
			dbName: "baz.db2",
		},
		{
			name:   "relative path and simple db name",
			elem:   []string{"foo/bar", "baz.db"},
			dbPath: "foo/bar",
			dbName: "baz",
		},
		{
			name:   "relative path and simple db name no suffix",
			elem:   []string{"foo/bar", "baz"},
			dbPath: "foo/bar",
			dbName: "baz",
		},
		{
			name:   "relative path and simple db name weird suffix",
			elem:   []string{"foo/bar", "baz.db2"},
			dbPath: "foo/bar",
			dbName: "baz.db2",
		},
		{
			name:   "relative path and db in sub dir",
			elem:   []string{"foo", "bar/baz.db"},
			dbPath: "foo/bar",
			dbName: "baz",
		},
		{
			name:   "relative path and db in sub dir no suffix",
			elem:   []string{"foo", "bar/baz"},
			dbPath: "foo/bar",
			dbName: "baz",
		},
		{
			name:   "relative path and db in sub dir weird suffix",
			elem:   []string{"foo", "bar/baz.db2"},
			dbPath: "foo/bar",
			dbName: "baz.db2",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			dbPath, dbName := splitDBPath(tc.elem...)
			assert.Equal(t, tc.dbPath, dbPath, "dbPath")
			assert.Equal(t, tc.dbName, dbName, "dbName")
		})
	}
}

func (s MigratorTestSuite) TestGetDataDirContents() {
	// Setup a temp directory with the following:
	// 1) A directory named dbdir1.db with nothing in it
	// 2) A directory named dbdir2 with files named MANIFEST, other1.txt, other2.log
	// 3) A directory named subdir1 with:
	//       a) a directory named dbdir3.db with nothing in it
	//       b) a directory named dbdir4 with files named MANIFEST, other3.txt, other4.log
	//       c) A file named not-a-db-1.txt
	// 4) A directory named subdir2 with files: other5.txt, other6.log
	// 5) a file named not-a-db-2.txt
	// 6) A directory named subdir3 with:
	//       a) a directory named subsubdir1 with a file named other7.txt
	//       b) a directory named subsubdir2 with a file named other8.txt

	tDir := s.T().TempDir()

	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "dbdir1.db"), 0700), "making dbdir1.db")

	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "dbdir2"), 0700), "making dbdir2")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "dbdir2", "MANIFEST"), []byte{}, 0700), "making dbdir2/MANIFEST")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "dbdir2", "other1.txt"), []byte{}, 0700), "making dbdir2/other1.txt")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "dbdir2", "other2.log"), []byte{}, 0700), "making dbdir2/other2.log")

	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "subdir1", "dbdir3.db"), 0700), "making subdir1/dbdir3.db")
	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "subdir1", "dbdir4"), 0700), "making subdir1/dbdir4")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir1", "dbdir4", "MANIFEST"), []byte{}, 0700), "making subdir1/dbdir4/MANIFEST")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir1", "dbdir4", "other3.txt"), []byte{}, 0700), "making subdir1/dbdir4/other3.txt")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir1", "dbdir4", "other4.log"), []byte{}, 0700), "making subdir1/dbdir4/other4.log")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir1", "not-a-db-1.txt"), []byte{}, 0700), "making subdir1/not-a-db-1.txt")

	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "subdir2"), 0700), "making subdir2")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir2", "other5.txt"), []byte{}, 0700), "making subdir2/other5.txt")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir2", "other6.log"), []byte{}, 0700), "making subdir2/other6.log")

	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "not-a-db-2.txt"), []byte{}, 0700), "making not-a-db-2.txt")

	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "subdir3", "subsubdir1"), 0700), "making subsubdir1")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir3", "subsubdir1", "other7.txt"), []byte{}, 0700), "making subdir2/other5.txt")
	s.Require().NoError(os.MkdirAll(filepath.Join(tDir, "subdir3", "subsubdir2"), 0700), "making subsubdir2")
	s.Require().NoError(os.WriteFile(filepath.Join(tDir, "subdir3", "subsubdir2", "other8.txt"), []byte{}, 0700), "making subdir2/other5.txt")

	s.T().Run("standard use case", func(t *testing.T) {
		expectedDbs := []string{"dbdir1.db", "dbdir2", "subdir1/dbdir3.db", "subdir1/dbdir4"}
		expectedNonDBs := []string{"subdir1/not-a-db-1.txt", "subdir2", "not-a-db-2.txt", "subdir3"}

		dbs, nonDBs, err := GetDataDirContents(tDir)

		require.NoError(t, err, "calling GetDataDirContents")

		assert.Len(t, dbs, len(expectedDbs), "dbs")
		for _, eDB := range expectedDbs {
			assert.Contains(t, dbs, eDB, "dbs")
		}

		assert.Len(t, nonDBs, len(expectedNonDBs), "nonDBs")
		for _, eNonDB := range expectedNonDBs {
			assert.Contains(t, nonDBs, eNonDB, "nonDBs")
		}
	})

	s.T().Run("directory does not exist", func(t *testing.T) {
		_, _, err := GetDataDirContents(tDir + "-nope-not-gonna-exist")
		require.Error(t, err, "GetDataDirContents on directory that doesn't exist.")
		assert.Contains(t, err.Error(), "no such file or directory", "err")
	})
}

func (s MigratorTestSuite) TestDetectDBType() {
	tDir := s.T().TempDir()

	s.T().Run("badger", func(t *testing.T) {
		expected := tmdb.BadgerDBBackend
		name := "badger1"
		dataDir := filepath.Join(tDir, "badger")
		dbDir := filepath.Join(dataDir, name)
		require.NoError(t, os.MkdirAll(dbDir, 0700), "making dbDir")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "KEYREGISTRY"), []byte{}, 0600), "making KEYREGISTRY")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "MANIFEST"), []byte{}, 0600), "making KEYREGISTRY")
		actual, ok := DetectDBType(name, dataDir)
		assert.True(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	s.T().Run("rocks", func(t *testing.T) {
		expected := tmdb.RocksDBBackend
		name := "rocks2"
		dataDir := filepath.Join(tDir, "rocks")
		dbDir := filepath.Join(dataDir, name+".db")
		require.NoError(t, os.MkdirAll(dbDir, 0700), "making dbDir")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "CURRENT"), []byte{}, 0600), "making CURRENT")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "LOG"), []byte{}, 0600), "making LOG")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "IDENTITY"), []byte{}, 0600), "making IDENTITY")
		actual, ok := DetectDBType(name, dataDir)
		assert.True(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	// To run this test, you'll need to provide the tag 'cleveldb' to the test command.
	// Both make test and the github action should have that tag, but you might need
	// to tell your IDE about it in order to use it to run this test.
	if IsPossibleDBType("cleveldb") {
		s.T().Run("clevel", func(t *testing.T) {
			// As far as I can tell, you can always open a cleveldb using goleveldb, but not vice versa.
			// Since DetectDBType checks for goleveldb first, it should return as goleveldb in this test.
			expected := tmdb.GoLevelDBBackend
			name := "clevel3"
			dataDir := filepath.Join(tDir, "clevel")
			require.NoError(t, os.MkdirAll(dataDir, 0700), "making data dir")
			// The reason the other db types aren't done this way (creating the db with NewDB) is that
			// I didn't want to cause confusion with regard to build tags and external library dependencies.
			db, err := tmdb.NewDB(name, tmdb.CLevelDBBackend, dataDir)
			require.NoError(t, err, "NewDB")
			for i := 0; i < 15; i++ {
				assert.NoError(t, db.Set([]byte(fmt.Sprintf("%s-key-%d", name, i)), []byte(fmt.Sprintf("%s-value-%d", name, i))), "setting key/value %d", i)
			}
			require.NoError(t, db.Close(), "closing db")
			actual, ok := DetectDBType(name, dataDir)
			assert.True(t, ok, "DetectDBType bool")
			assert.Equal(t, expected, actual, "DetectDBType BackendType")
		})
	}

	s.T().Run("golevel", func(t *testing.T) {
		expected := tmdb.GoLevelDBBackend
		name := "golevel8"
		dataDir := filepath.Join(tDir, "golevel")
		require.NoError(t, os.MkdirAll(dataDir, 0700), "making data dir")
		// The reason the other db types aren't done this way (creating the db with NewDB) is that
		// I didn't want to cause confusion with regard to build tags and external library dependencies.
		db, err := tmdb.NewDB(name, expected, dataDir)
		require.NoError(t, err, "NewDB")
		for i := 0; i < 15; i++ {
			assert.NoError(t, db.Set([]byte(fmt.Sprintf("%s-key-%d", name, i)), []byte(fmt.Sprintf("%s-value-%d", name, i))), "setting key/value %d", i)
		}
		require.NoError(t, db.Close(), "closing db")
		actual, ok := DetectDBType(name, dataDir)
		assert.True(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	s.T().Run("boltdb", func(t *testing.T) {
		expected := tmdb.BoltDBBackend
		name := "bolt7"
		dataDir := filepath.Join(tDir, "bolt")
		dbFile := filepath.Join(dataDir, name+".db")
		require.NoError(t, os.MkdirAll(dataDir, 0700), "making dataDir")
		require.NoError(t, os.WriteFile(dbFile, []byte{}, 0700), "making dbFile")
		actual, ok := DetectDBType(name, dataDir)
		assert.True(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	s.T().Run("empty", func(t *testing.T) {
		expected := unknownDBBackend
		name := "empty4"
		dataDir := filepath.Join(tDir, "empty")
		dbDir := filepath.Join(dataDir, name)
		require.NoError(t, os.MkdirAll(dbDir, 0700), "making dbDir")
		actual, ok := DetectDBType(name, dataDir)
		assert.False(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	s.T().Run("only current", func(t *testing.T) {
		expected := unknownDBBackend
		name := "only-current5"
		dataDir := filepath.Join(tDir, "only-current")
		dbDir := filepath.Join(dataDir, name+".db")
		require.NoError(t, os.MkdirAll(dbDir, 0700), "making dbDir")
		require.NoError(t, os.WriteFile(filepath.Join(dbDir, "CURRENT"), []byte{}, 0600), "making CURRENT")
		actual, ok := DetectDBType(name, dataDir)
		assert.False(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})

	s.T().Run("does not exist", func(t *testing.T) {
		expected := unknownDBBackend
		name := "does-not-exist6"
		dataDir := filepath.Join(tDir, "only-current")
		actual, ok := DetectDBType(name, dataDir)
		assert.False(t, ok, "DetectDBType bool")
		assert.Equal(t, expected, actual, "DetectDBType BackendType")
	})
}

func (s MigratorTestSuite) TestDirExists() {
	s.T().Run("does not exist", func(t *testing.T) {
		assert.False(t, dirExists("does not exist"))
	})

	s.T().Run("containing dir exists", func(t *testing.T) {
		tdir := t.TempDir()
		dir := filepath.Join(tdir, "nope")
		assert.False(t, dirExists(dir))
	})

	s.T().Run("is file", func(t *testing.T) {
		tdir := t.TempDir()
		file := filepath.Join(tdir, "filiename.txt")
		require.NoError(t, os.WriteFile(file, []byte{}, 0600), "making file")
		assert.False(t, dirExists(file))
	})

	s.T().Run("is dir", func(t *testing.T) {
		tdir := t.TempDir()
		dir := filepath.Join(tdir, "immadir")
		require.NoError(t, os.MkdirAll(dir, 0700), "making dir")
		assert.True(t, dirExists(dir))
	})
}

func (s MigratorTestSuite) TestFileExists() {
	s.T().Run("does not exist", func(t *testing.T) {
		assert.False(t, fileExists("does not exist"))
	})

	s.T().Run("containing dir exists", func(t *testing.T) {
		tdir := t.TempDir()
		file := filepath.Join(tdir, "nope.tar")
		assert.False(t, fileExists(file))
	})

	s.T().Run("is file", func(t *testing.T) {
		tdir := t.TempDir()
		file := filepath.Join(tdir, "filiename.txt")
		require.NoError(t, os.WriteFile(file, []byte{}, 0600), "making file")
		assert.True(t, fileExists(file))
	})

	s.T().Run("is dir", func(t *testing.T) {
		tdir := t.TempDir()
		dir := filepath.Join(tdir, "immadir")
		require.NoError(t, os.MkdirAll(dir, 0700), "making dir")
		assert.False(t, fileExists(dir))
	})
}

func (s MigratorTestSuite) TestCommaString() {
	tests := []struct {
		v   uint
		exp string
	}{
		{v: 0, exp: "0"},
		{v: 1, exp: "1"},
		{v: 22, exp: "22"},
		{v: 333, exp: "333"},
		{v: 999, exp: "999"},
		{v: 1_000, exp: "1,000"},
		{v: 4_444, exp: "4,444"},
		{v: 55_555, exp: "55,555"},
		{v: 666_666, exp: "666,666"},
		{v: 999_999, exp: "999,999"},
		{v: 1_000_000, exp: "1,000,000"},
		{v: 7_777_777, exp: "7,777,777"},
		{v: 88_888_888, exp: "88,888,888"},
		{v: 999_999_999, exp: "999,999,999"},
		{v: 1_000_000_000, exp: "1,000,000,000"},
		{v: 1_010_101_010, exp: "1,010,101,010"},
		{v: 11_011_011_011, exp: "11,011,011,011"},
		{v: 120_120_120_120, exp: "120,120,120,120"},
		{v: 999_999_999_999, exp: "999,999,999,999"},
		{v: 1_000_000_000_000, exp: "1,000,000,000,000"},
		{v: 1_301_301_301_301, exp: "1,301,301,301,301"},
		{v: 14_814_714_614_514, exp: "14,814,714,614,514"},
		{v: 150_151_152_153_154, exp: "150,151,152,153,154"},
		{v: 999_999_999_999_999, exp: "999,999,999,999,999"},
		{v: 1_000_000_000_000_000, exp: "1,000,000,000,000,000"},
		{v: 1_651_651_651_651_651, exp: "1,651,651,651,651,651"},
		{v: 17_017_017_017_017_017, exp: "17,017,017,017,017,017"},
		{v: 189_189_189_189_189_189, exp: "189,189,189,189,189,189"},
		{v: 999_999_999_999_999_999, exp: "999,999,999,999,999,999"},
		{v: 1_000_000_000_000_000_000, exp: "1,000,000,000,000,000,000"},
		{v: 1_981_981_981_981_981_981, exp: "1,981,981,981,981,981,981"},
		{v: 18_446_744_073_709_551_615, exp: "18,446,744,073,709,551,615"},
	}

	for _, tc := range tests {
		s.T().Run(tc.exp, func(t *testing.T) {
			act := commaString(tc.v)
			assert.Equal(t, tc.exp, act)
		})
	}
}
