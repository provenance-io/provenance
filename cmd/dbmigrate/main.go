package main

import (
	"errors"
	"os"

	"github.com/provenance-io/provenance/cmd/dbmigrate/cmd"
	cmderrors "github.com/provenance-io/provenance/cmd/errors"
)

func main() {
	rootCmd := cmd.NewDBMigrateCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		var srvErrP *cmderrors.ExitCodeError
		var srvErr cmderrors.ExitCodeError
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
