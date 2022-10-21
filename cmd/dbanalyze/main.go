package main

import (
	"errors"
	"os"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/provenance-io/provenance/cmd/dbanalyze/cmd"
)

func main() {
	rootCmd := cmd.NewDBAnalyzeCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		var srvErr *server.ErrorCode
		switch {
		case errors.As(err, &srvErr):
			os.Exit(srvErr.Code)
		default:
			os.Exit(1)
		}
	}
}
