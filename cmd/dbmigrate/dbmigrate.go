package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	tmdb "github.com/tendermint/tm-db"
)

func main() {
	pio, dbs, err := getDatabaseFolders()
	if err != nil {
		fmt.Printf("Unable to get  database folders: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PIO Home is: %v\n", pio)

	targetDir := pio + "/data"
	backupDir := pio + "/backup"
	for _, file := range dbs {
		backupAndConvertDatabase(targetDir, backupDir, file)
	}
}

// Backs up database file and converts it into the original location.
func backupAndConvertDatabase(targetDir, backupDir string, file os.FileInfo) error {
	nodbsuffix := file.Name()
	nodbsuffix = strings.Split(nodbsuffix, ".")[0]

	original := targetDir + "/" + file.Name()
	backup := backupDir + "/" + file.Name()
	fmt.Printf("Backing up original database to '%v'\n", backup)
	err := os.Rename(original, backup)
	if err != nil {
		return err
	}

	// Open source database from backup location.
	fmt.Printf("Converting database '%v'...\n", file.Name())
	sdb, err := tmdb.NewGoLevelDB(nodbsuffix, backupDir)
	if err != nil {
		return err
	}
	defer sdb.Close()

	// Open target database in destination location.
	tdb, err := tmdb.NewBadgerDB(nodbsuffix, targetDir)
	if err != nil {
		return err
	}
	defer tdb.Close()

	// Start a batch write operation on target database.
	wb := tdb.NewBatch()
	defer wb.Close()

	// Iterate through all keys in LevelDB and write to BadgerDB.
	iter, err := sdb.Iterator(nil, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		err := wb.Set(key, value)
		if err != nil {
			return err
		}
	}

	err = wb.Write()
	if err != nil {
		return err
	}

	return nil
}

// Get PIO_ROOT and database folders in its data folder.
func getDatabaseFolders() (string, []os.FileInfo, error) {
	pio, found := os.LookupEnv("PIO_HOME")
	if !found {
		pio, _ = os.Getwd()
	}

	files, err := ioutil.ReadDir(pio + "/data")
	if err != nil {
		return pio, nil, err
	}

	dbs := make([]os.FileInfo, 0)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".db" {
			dbs = append(dbs, file)
		}
	}
	return pio, dbs, nil
}
