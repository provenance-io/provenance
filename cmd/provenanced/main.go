package main

import (
	"errors"
	"os"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(true)
	if err := cmd.Execute(rootCmd); err != nil {
		var srvErrP *server.ErrorCode
		var srvErr server.ErrorCode
		switch {
		case errors.As(err, &srvErrP):
			os.Exit(srvErrP.Code)
		case errors.As(err, &srvErr):
			os.Exit(srvErr.Code)
		default:
			os.Exit(1)
		}
	}
}
