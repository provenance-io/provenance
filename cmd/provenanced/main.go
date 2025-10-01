package main

import (
	"errors"
	"os"

	cmderrors "github.com/provenance-io/provenance/cmd/errors"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(true)
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
