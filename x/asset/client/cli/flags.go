package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	FlagURI         = "uri"
	FlagURIHash     = "uri-hash"
	FlagSymbol      = "symbol"
	FlagDescription = "description"
	FlagOwner       = "owner"
)

// AddFlagsURI adds the --uri and --uri-hash flags to the provided command.
func AddFlagsURI(cmd *cobra.Command, typeName string) {
	cmd.Flags().String(FlagURI, "", "URI of the "+typeName)
	cmd.Flags().String(FlagURIHash, "", "hash of the content of the "+typeName+" URI")
}

// ReadFlagURI returns the value provided with the --uri flag.
func ReadFlagURI(flagSet *pflag.FlagSet) string {
	// GetString only returns an error if the flag wasn't set up on the cmd, but we don't care here.
	rv, _ := flagSet.GetString(FlagURI)
	return rv
}

// ReadFlagURIHash returns the value provided with the --uri-hash flag.
func ReadFlagURIHash(flagSet *pflag.FlagSet) string {
	// GetString only returns an error if the flag wasn't set up on the cmd, but we don't care here.
	rv, _ := flagSet.GetString(FlagURIHash)
	return rv
}

// AddFlagSymbol adds the --symbol flag to the provided command.
func AddFlagSymbol(cmd *cobra.Command) {
	cmd.Flags().String(FlagSymbol, "", "symbol of the asset class")
}

// ReadFlagSymbol returns the value provided with the --symbol flag.
func ReadFlagSymbol(flagSet *pflag.FlagSet) string {
	// GetString only returns an error if the flag wasn't set up on the cmd, but we don't care here.
	rv, _ := flagSet.GetString(FlagSymbol)
	return rv
}

// AddFlagDescription adds the --description flag to the provided command.
func AddFlagDescription(cmd *cobra.Command) {
	cmd.Flags().String(FlagDescription, "", "description of the asset class")
}

// ReadFlagDescription returns the value provided with the --description flag.
func ReadFlagDescription(flagSet *pflag.FlagSet) string {
	// GetString only returns an error if the flag wasn't set up on the cmd, but we don't care here.
	rv, _ := flagSet.GetString(FlagDescription)
	return rv
}

// AddFlagOwner adds the --owner flag to the provided command.
func AddFlagOwner(cmd *cobra.Command) {
	cmd.Flags().String(FlagOwner, "", "owner address")
}

// ReadFlagOwner returns the value provided with the --owner flag.
func ReadFlagOwner(flagSet *pflag.FlagSet) string {
	rv, _ := flagSet.GetString(FlagOwner)
	return rv
}
