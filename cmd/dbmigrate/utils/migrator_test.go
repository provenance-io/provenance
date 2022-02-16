package utils

import (
	"os"
	"path/filepath"
	"testing"

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

// TODO: SetUpMigrator tests
// TODO: Migrate tests
// TODO: MigrateDBDir tests
// TODO: UpdateConfig tests

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
