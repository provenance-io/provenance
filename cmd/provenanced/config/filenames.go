package config

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

const (
	// These are not configuration values.
	// They are strings that are also defined deep inside some other stuff that is out of our control.
	// Do not change them.

	// ConfigSubDir is the subdirectory of HOME that contains the configuration files.
	ConfigSubDir = "config"
	// AppConfFilename is the filename of the app/cosmos configuration file.
	AppConfFilename = "app.toml"
	// TmConfFilename is the filename of the tendermint configuration file.
	TmConfFilename = "config.toml"
	// ClientConfFilename is the filename of the client configuration file.
	ClientConfFilename = "client.toml"

	// This one is in our full control but it's still probably not a good idea to change it.

	// PackedConfFilename is the filename of the packed (non-defaults) file.
	PackedConfFilename = "packed-conf.json"
)

// GetHomeDir gets the home directory from the provided cobra command.
func GetHomeDir(cmd *cobra.Command) string {
	return client.GetClientContextFromCmd(cmd).HomeDir
}

// GetFullPathToConfigDir gets the full path to the config directory.
func GetFullPathToConfigDir(cmd *cobra.Command) string {
	return filepath.Join(GetHomeDir(cmd), ConfigSubDir)
}

// GetFullPathToAppConf gets the full path to the app/cosmos config file.
func GetFullPathToAppConf(cmd *cobra.Command) string {
	return filepath.Join(GetHomeDir(cmd), ConfigSubDir, AppConfFilename)
}

// GetFullPathToTmConf gets the full path to the tendermint config file.
func GetFullPathToTmConf(cmd *cobra.Command) string {
	return filepath.Join(GetHomeDir(cmd), ConfigSubDir, TmConfFilename)
}

// GetFullPathToClientConf gets the full path to the client config file.
func GetFullPathToClientConf(cmd *cobra.Command) string {
	return filepath.Join(GetHomeDir(cmd), ConfigSubDir, ClientConfFilename)
}

// GetFullPathToPackedConf gets the full path to the packed/non-defaults config file.
func GetFullPathToPackedConf(cmd *cobra.Command) string {
	return filepath.Join(GetHomeDir(cmd), ConfigSubDir, PackedConfFilename)
}
