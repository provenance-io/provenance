package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/provenance-io/provenance/cmd/dbmigrate/cmd"
)

func main() {
	rootCmd := cmd.NewDBMigrateCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
