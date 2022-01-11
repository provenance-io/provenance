package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/pkg/profile"
	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
)

func main() {

	//defer profile.Start(profile.ProfilePath(".")).Stop()
	defer profile.Start(profile.MemProfileHeap,profile.ProfilePath(".")).Stop()
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
