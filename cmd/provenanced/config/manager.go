package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmtconfig "github.com/cometbft/cometbft/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

// DefaultConsensusTimeoutCommit is the default value used for the consensus.timeout_commit config value.
var DefaultConsensusTimeoutCommit = 1500 * time.Millisecond

// PackConfig generates and saves the packed config file then removes the individual config files.
func PackConfig(cmd *cobra.Command) error {
	generateAndWritePackedConfig(cmd, nil, nil, nil, true)
	err := deleteUnpackedConfig(cmd, true)
	return err
}

// UnpackConfig generates the saves the individual config files and removes the packed config file.
func UnpackConfig(cmd *cobra.Command) error {
	appConfig, appConfErr := ExtractAppConfig(cmd)
	if appConfErr != nil {
		return fmt.Errorf("could not get app config values: %w", appConfErr)
	}
	cmtConfig, cmtConfErr := ExtractCmtConfig(cmd)
	if cmtConfErr != nil {
		return fmt.Errorf("could not get cometbft config values: %w", cmtConfErr)
	}
	clientConfig, clientConfErr := ExtractClientConfig(cmd)
	if clientConfErr != nil {
		return fmt.Errorf("could not get client config values: %w", clientConfErr)
	}
	writeUnpackedConfig(cmd, appConfig, cmtConfig, clientConfig, true)
	err := deletePackedConfig(cmd, true)
	return err
}

// IsPacked checks to see if we're using a packed config or not.
// returns true if using the packed config.
// returns false if using the unpacked multiple config files.
func IsPacked(cmd *cobra.Command) bool {
	return FileExists(GetFullPathToPackedConf(cmd))
}

// FileExists returns true if there is no error getting stat info of the file.
// Even returns false if os.IsNotExist(err) is false.
// I.e. if the file exists but isn't readable, this will still return false.
func FileExists(fullFilePath string) bool {
	_, err := os.Stat(fullFilePath)
	return err == nil
}

// ExtractAppConfig creates an app/cosmos config from the command context.
func ExtractAppConfig(cmd *cobra.Command) (*serverconfig.Config, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := DefaultAppConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("error extracting app config: %w", err)
	}
	return conf, nil
}

// ExtractAppConfigAndMap from the command context, creates an app/cosmos config and related string->value map.
func ExtractAppConfigAndMap(cmd *cobra.Command) (*serverconfig.Config, FieldValueMap, error) {
	conf, err := ExtractAppConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	fields := MakeFieldValueMap(conf, true)
	return conf, fields, nil
}

// DefaultAppConfig gets our default app config.
func DefaultAppConfig() *serverconfig.Config {
	rv := serverconfig.DefaultConfig()
	rv.MinGasPrices = pioconfig.GetProvConfig().ProvMinGasPrices
	rv.IAVLDisableFastNode = true
	return rv
}

// ExtractCmtConfig creates a cometbft config from the command context.
func ExtractCmtConfig(cmd *cobra.Command) (*cmtconfig.Config, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := DefaultCmtConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("error extracting cometbft config: %w", err)
	}
	conf.SetRoot(GetHomeDir(cmd))
	// When the RPCServers value is "", it gets read as []string{}.
	// But from a packed config without that entry, it gets read as nil.
	// The default is also nil.
	// So, if there's nothing in it, just set it to nil for consistency.
	if len(conf.StateSync.RPCServers) == 0 {
		conf.StateSync.RPCServers = nil
	}
	return conf, nil
}

// ExtractCmtConfigAndMap from the command context, creates a cometbft config and related string->value map.
func ExtractCmtConfigAndMap(cmd *cobra.Command) (*cmtconfig.Config, FieldValueMap, error) {
	conf, err := ExtractCmtConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	fields := MakeFieldValueMap(conf, true)
	removeUndesirableCmtConfigEntries(fields)
	return conf, fields, nil
}

func DefaultCmtConfig() *cmtconfig.Config {
	rv := cmtconfig.DefaultConfig()
	rv.Consensus.TimeoutCommit = DefaultConsensusTimeoutCommit
	rv.TxIndex.Indexer = "null"
	return rv
}

// removeUndesirableCmtConfigEntries deletes some keys from the provided fields map that we don't want included.
// The provided map is altered during this call. It is also returned from this func.
// There are several fields in the cometbft config struct that don't correspond to entries in the config files.
// None of the "home" keys have entries in the config files:
// "home", "consensus.home", "mempool.home", "p2p.home", "rpc.home"
// There are several "p2p.test_" fields that should be ignored too.
// "p2p.test_dial_fail", "p2p.test_fuzz",
// "p2p.test_fuzz_config.*" ("maxdelay", "mode", "probdropconn", "probdroprw", "probsleep")
// This info is accurate in Cosmos SDK 0.43 (on 2021-08-16).
func removeUndesirableCmtConfigEntries(fields FieldValueMap) FieldValueMap {
	delete(fields, "home")
	for k := range fields {
		if strings.HasSuffix(k, ".home") || strings.HasPrefix(k, "p2p.test_") {
			delete(fields, k)
		}
	}
	return fields
}

// ExtractClientConfig creates a client config from the command context.
func ExtractClientConfig(cmd *cobra.Command) (*ClientConfig, error) {
	v := client.GetClientContextFromCmd(cmd).Viper
	conf := DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("error extracting client config: %w", err)
	}
	return conf, nil
}

// ExtractClientConfigAndMap from the command context, creates a client config and related string->value map.
func ExtractClientConfigAndMap(cmd *cobra.Command) (*ClientConfig, FieldValueMap, error) {
	conf, err := ExtractClientConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	fields := MakeFieldValueMap(conf, true)
	return conf, fields, nil
}

// GetAllConfigDefaults gets a field map from the defaults of all the configs.
func GetAllConfigDefaults() FieldValueMap {
	rv := FieldValueMap{}
	rv.AddEntriesFrom(
		MakeFieldValueMap(DefaultAppConfig(), false),
		removeUndesirableCmtConfigEntries(MakeFieldValueMap(DefaultCmtConfig(), false)),
		MakeFieldValueMap(DefaultClientConfig(), false),
	)
	return rv
}

// SaveConfigs saves the configs to files.
// If the config is packed, any nil configs provided will extracted from the cmd.
// If the config is unpacked, only the configs provided will be written.
// Any errors encountered will result in a panic.
func SaveConfigs(
	cmd *cobra.Command,
	appConfig *serverconfig.Config,
	cmtConfig *cmtconfig.Config,
	clientConfig *ClientConfig,
	verbose bool,
) {
	if IsPacked(cmd) {
		generateAndWritePackedConfig(cmd, appConfig, cmtConfig, clientConfig, verbose)
	} else {
		writeUnpackedConfig(cmd, appConfig, cmtConfig, clientConfig, verbose)
	}
}

// writeUnpackedConfig writes the provided configs to their files.
// Any config parameter provided as nil will be skipped.
// Any errors encountered will result in a panic or exit.
func writeUnpackedConfig(
	cmd *cobra.Command,
	appConfig *serverconfig.Config,
	cmtConfig *cmtconfig.Config,
	clientConfig *ClientConfig,
	verbose bool,
) {
	mustEnsureConfigDir(cmd)
	if appConfig != nil {
		confFile := GetFullPathToAppConf(cmd)
		if verbose {
			cmd.Printf("Writing app config to: %s ... ", confFile)
		}
		serverconfig.WriteConfigFile(confFile, appConfig)
		if verbose {
			cmd.Printf("Done.\n")
		}
	}
	if cmtConfig != nil {
		confFile := GetFullPathToCmtConf(cmd)
		if verbose {
			cmd.Printf("Writing cometbft config to: %s ... ", confFile)
		}
		cmtconfig.WriteConfigFile(confFile, cmtConfig)
		if verbose {
			cmd.Printf("Done.\n")
		}
	}
	if clientConfig != nil {
		confFile := GetFullPathToClientConf(cmd)
		if verbose {
			cmd.Printf("Writing client config to: %s ... ", confFile)
		}
		WriteConfigToFile(confFile, clientConfig)
		if verbose {
			cmd.Printf("Done.\n")
		}
	}
}

// deleteUnpackedConfig deletes all the unpacked config files.
// An attempt will be made to remove each file before returning.
// The error returned might reflect multiple failures to delete.
// Any files that don't exist, are ignored.
func deleteUnpackedConfig(cmd *cobra.Command, verbose bool) error {
	configFiles := []string{
		GetFullPathToAppConf(cmd),
		GetFullPathToCmtConf(cmd),
		GetFullPathToClientConf(cmd),
	}
	var rvErr error
	for _, f := range configFiles {
		err := deleteConfigFile(cmd, f, verbose)
		if err != nil {
			rvErr = appendError(rvErr, err)
		}
	}
	if rvErr != nil {
		rvErr = fmt.Errorf("one or more unpacked config files could not be removed\n%w", rvErr)
	}
	return rvErr
}

// generateAndWritePackedConfig generates the contents of the packed config file and saves it.
// Any config parameter provided as nil will be retrieved from the cmd.
// Any errors encountered will result in a panic.
func generateAndWritePackedConfig(
	cmd *cobra.Command,
	appConfig *serverconfig.Config,
	cmtConfig *cmtconfig.Config,
	clientConfig *ClientConfig,
	verbose bool,
) {
	mustEnsureConfigDir(cmd)
	var appConfMap, cmtConfMap, clientConfMap FieldValueMap
	if appConfig == nil {
		var err error
		_, appConfMap, err = ExtractAppConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract app config values: %w", err))
		}
	} else {
		appConfMap = MakeFieldValueMap(appConfig, false)
	}
	if cmtConfig == nil {
		var err error
		_, cmtConfMap, err = ExtractCmtConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract cometbft config values: %w", err))
		}
	} else {
		cmtConfMap = MakeFieldValueMap(cmtConfig, false)
	}
	if clientConfig == nil {
		var err error
		_, clientConfMap, err = ExtractClientConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract client config values: %w", err))
		}
	} else {
		clientConfMap = MakeFieldValueMap(clientConfig, false)
	}
	allConf := FieldValueMap{}
	allConf.AddEntriesFrom(appConfMap, cmtConfMap, clientConfMap)
	defaultConf := GetAllConfigDefaults()
	packed := map[string]string{}
	for key, info := range MakeUpdatedFieldMap(defaultConf, allConf, true) {
		packed[key] = unquote(info.IsNow)
	}
	packedJSON, err := json.MarshalIndent(packed, "", "  ")
	if err != nil {
		panic(err)
	}
	if verbose {
		cmd.Printf("Packed config:\n%s\n", packedJSON)
	}
	packedFile := GetFullPathToPackedConf(cmd)

	//nolint:gosec // These are the correct permissions
	err = os.WriteFile(packedFile, packedJSON, 0644)
	if err != nil {
		panic(err)
	}
	if verbose {
		cmd.Printf("Packed config file saved: %s\n", packedFile)
	}
}

// deletePackedConfig deletes the packed config file.
func deletePackedConfig(cmd *cobra.Command, verbose bool) error {
	return deleteConfigFile(cmd, GetFullPathToPackedConf(cmd), verbose)
}

// deleteConfigFile deletes a file assuming it's a config file.
func deleteConfigFile(cmd *cobra.Command, filePath string, verbose bool) error {
	if verbose {
		cmd.Printf("Deleting config file: %s ... ", filePath)
	}
	err := os.Remove(filePath)
	switch {
	case err == nil:
		if verbose {
			cmd.Printf("Done.\n")
		}
	case os.IsNotExist(err):
		if verbose {
			cmd.Print("Does not exist.\n")
		}
	default:
		if verbose {
			cmd.Printf("Error.\n")
		}
		cmd.PrintErrf("Error removing file: %v\n", err)
		return err
	}
	return nil
}

// EnsureConfigDir ensures the given directory exists, creating it if necessary.
// Errors if the config directory doesn't exist and cannot be created (e.g. because of permissions).
// Errors if the path already exists as a non-directory.
func EnsureConfigDir(cmd *cobra.Command) error {
	dir := GetFullPathToConfigDir(cmd)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}
	return nil
}

// mustEnsureConfigDir is the same as EnsureConfigDir except panics on error.
func mustEnsureConfigDir(cmd *cobra.Command) {
	if err := EnsureConfigDir(cmd); err != nil {
		panic(err)
	}
}

// unquote removes a leading and trailing double quote if they're both there.
func unquote(str string) string {
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}

// LoadConfigFromFiles loads configurations appropriately.
func LoadConfigFromFiles(cmd *cobra.Command) (err error) {
	if IsPacked(cmd) {
		err = loadPackedConfig(cmd)
	} else {
		err = loadUnpackedConfig(cmd)
	}
	if err != nil {
		return err
	}
	return loadUnmanagedConfig(cmd)
}

// loadUnmanagedConfig reads the unmanaged config file into viper. It does not apply anything to any contexts though.
func loadUnmanagedConfig(cmd *cobra.Command) error {
	unmanagedConfFile := GetFullPathToUnmanagedConf(cmd)
	// Both the server context and client context should be using the same Viper, so this is good for both.
	vpr := server.GetServerContextFromCmd(cmd).Viper
	// Load the unmanaged config if it exists, or else do nothing.
	switch _, err := os.Stat(unmanagedConfFile); {
	case os.IsNotExist(err):
		// It doesn't exist. Nothing to do.
		return nil
	case err != nil:
		return fmt.Errorf("unmanaged config file stat error: %w", err)
	default:
		vpr.SetConfigFile(unmanagedConfFile)
		rerr := vpr.MergeInConfig()
		if rerr != nil {
			return fmt.Errorf("unmanaged config file read error: %w", rerr)
		}
	}
	return nil
}

// loadUnpackedConfig attempts to read the unpacked config files and apply them to the appropriate contexts.
func loadUnpackedConfig(cmd *cobra.Command) error {
	appConfFile := GetFullPathToAppConf(cmd)
	cmtConfFile := GetFullPathToCmtConf(cmd)
	clientConfFile := GetFullPathToClientConf(cmd)

	// Both the server context and client context should be using the same Viper, so this is good for both.
	vpr := server.GetServerContextFromCmd(cmd).Viper

	// Load the cometbft config defaults, then file if it exists.
	tdErr := addFieldMapToViper(vpr, MakeFieldValueMap(DefaultCmtConfig(), false))
	if tdErr != nil {
		return fmt.Errorf("cometbft config defaults load error: %w", tdErr)
	}
	switch _, err := os.Stat(cmtConfFile); {
	case os.IsNotExist(err):
		// Do nothing.
	case err != nil:
		return fmt.Errorf("cometbft config file stat error: %w", err)
	default:
		vpr.SetConfigFile(cmtConfFile)
		rerr := vpr.MergeInConfig()
		if rerr != nil {
			return fmt.Errorf("cometbft config file read error: %w", rerr)
		}
	}

	// Load the app/cosmos config defaults, then file if it exists.
	adErr := addFieldMapToViper(vpr, MakeFieldValueMap(DefaultAppConfig(), false))
	if adErr != nil {
		return fmt.Errorf("app config defaults load error: %w", adErr)
	}
	switch _, err := os.Stat(appConfFile); {
	case os.IsNotExist(err):
		// Do nothing.
	case err != nil:
		return fmt.Errorf("app config file stat error: %w", err)
	default:
		vpr.SetConfigFile(appConfFile)
		merr := vpr.MergeInConfig()
		if merr != nil {
			return fmt.Errorf("app config file merge error: %w", merr)
		}
	}

	// Load the client config defaults, then file if it exists.
	cdErr := addFieldMapToViper(vpr, MakeFieldValueMap(DefaultClientConfig(), false))
	if cdErr != nil {
		return fmt.Errorf("client config defaults load error: %w", cdErr)
	}
	switch _, err := os.Stat(clientConfFile); {
	case os.IsNotExist(err):
		// Do nothing.
	case err != nil:
		return fmt.Errorf("client config file stat error: %w", err)
	default:
		vpr.SetConfigFile(clientConfFile)
		rerr := vpr.MergeInConfig()
		if rerr != nil {
			return fmt.Errorf("client config file read error: %w", rerr)
		}
	}

	return applyConfigsToContexts(cmd)
}

// loadPackedConfig attempts to read the packed config and applies it to the appropriate contexts.
func loadPackedConfig(cmd *cobra.Command) error {
	packedConfFile := GetFullPathToPackedConf(cmd)

	// Read in the packed config if it exists.
	packedConf := map[string]string{}

	switch packedJSON, rerr := os.ReadFile(packedConfFile); {
	case os.IsNotExist(rerr):
		// Packed config file doesn't exist. Do nothing. Just let it use the defaults.
	case rerr != nil:
		return fmt.Errorf("packed config file read error: %w", rerr)
	default:
		jerr := json.Unmarshal(packedJSON, &packedConf)
		if jerr != nil {
			return fmt.Errorf("packed config file parse error: %w", jerr)
		}
	}

	// Start with the defaults
	appConfigMap := MakeFieldValueMap(DefaultAppConfig(), false)
	cmtConfigMap := MakeFieldValueMap(DefaultCmtConfig(), false)
	clientConfigMap := MakeFieldValueMap(DefaultClientConfig(), false)

	// Apply the packed config entries to the defaults.
	var rvErr error
	for k, v := range packedConf {
		found := false
		if appConfigMap.Has(k) {
			found = true
			err := appConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("app config key: %s, value: %s, err: %w", k, v, err))
			}
		}
		if cmtConfigMap.Has(k) {
			found = true
			err := cmtConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("cometbft config key: %s, value: %s, err: %w", k, v, err))
			}
		}
		if clientConfigMap.Has(k) {
			found = true
			err := clientConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("client config key: %s, value: %s, err: %w", k, v, err))
			}
		}
		if !found {
			cmd.PrintErrf("unknown packed config key: %s", k)
		}
	}
	if rvErr != nil {
		return fmt.Errorf("one or more fields in the packed config could not be set\n%w", rvErr)
	}

	// Set the config values as defaults in viper.
	// Viper doesn't really have a way to directly set a config value,
	// and a set value takes precedence over flags. So I guess defaults are what we go with.
	// The server and client should both have the same viper, so we only need the one.
	vpr := server.GetServerContextFromCmd(cmd).Viper
	if lerr := addFieldMapToViper(vpr, cmtConfigMap); lerr != nil {
		return fmt.Errorf("cometbft packed config load error: %w", lerr)
	}
	if lerr := addFieldMapToViper(vpr, appConfigMap); lerr != nil {
		return fmt.Errorf("app packed config load error: %w", lerr)
	}
	if lerr := addFieldMapToViper(vpr, clientConfigMap); lerr != nil {
		return fmt.Errorf("client packed config load error: %w", lerr)
	}

	return applyConfigsToContexts(cmd)
}

func addFieldMapToViper(vpr *viper.Viper, fvmap FieldValueMap) error {
	configMap, err := fvmap.AsConfigMap()
	if err != nil {
		return err
	}
	// The CometBFT BaseConfig struct has a RootDir field with the mapstruct name "home".
	// So the fvmap created from a CometBFT config struct will have a "home" entry with a value of "".
	// If we were to then include that when calling MergeConfigMap, it would tell viper that the "home" value is "".
	// This prevents the default value (defined with the --home flag) from being used.
	// Since the "home" value defines where the config files are, it's safe to assume none of the config
	// files will contain a "home" field that we need to pay attention to.
	// So for this, remove the "home" field from the configMap before calling MergeConfigMap.
	delete(configMap, flags.FlagHome)
	return vpr.MergeConfigMap(configMap)
}

func applyConfigsToContexts(cmd *cobra.Command) error {
	var err error
	// Apply what viper now has to the client context.
	clientCtx := client.GetClientContextFromCmd(cmd)
	clientConfig, err := ExtractClientConfig(cmd)
	if err != nil {
		return err
	}
	if clientCtx, err = ApplyClientConfigToContext(clientCtx, clientConfig); err != nil {
		f := GetFullPathToClientConf(cmd)
		if IsPacked(cmd) {
			f = GetFullPathToPackedConf(cmd)
		}
		return fmt.Errorf("could not apply client config %s to client context - it may need to be updated manually: %w", f, err)
	}
	if err = client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
		return fmt.Errorf("could not update client context on command: %w", err)
	}

	// Set the server context's config to what Viper has now.
	serverCtx := server.GetServerContextFromCmd(cmd)
	cmtConfig, err := ExtractCmtConfig(cmd)
	if err != nil {
		return err
	}
	serverCtx.Config = cmtConfig
	serverCtx.Config.SetRoot(clientCtx.HomeDir)

	// Set the server context's logger using what Viper has now.
	serverCtx.Logger, err = server.CreateSDKLogger(serverCtx, cmd.ErrOrStderr())
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}

	// Copy all settings from our viper to the global viper since some stuff uses the global one.
	if err = viper.MergeConfigMap(serverCtx.Viper.AllSettings()); err != nil {
		return fmt.Errorf("error copying all settings to global viper: %w", err)
	}

	return nil
}

// appendError appends one error onto another.
// If the origErr is nil, the newErr is returned.
// Otherwise, a new error is created with a combined error message.
func appendError(origErr, newErr error) error {
	if origErr == nil {
		return newErr
	}
	return fmt.Errorf("%v\n%v", origErr, newErr) //nolint:errorlint // Can't wrap two errors at once.
}
