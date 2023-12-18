package main

import (
	"errors"
	"os"

	"github.com/provenance-io/provenance/cmd/dbmigrate/cmd"
	"github.com/provenance-io/provenance/helpers"
)

func main() {
	rootCmd := cmd.NewDBMigrateCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		var srvErrP *helpers.ExitCodeError
		var srvErr helpers.ExitCodeError
		switch {
		case errors.As(err, &srvErrP):
			os.Exit(int(*srvErrP))
		case errors.As(err, &srvErr):
			os.Exit(int(srvErr))
		default:
			os.Exit(1)
		}
	}
}
