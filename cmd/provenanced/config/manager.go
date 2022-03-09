package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmconfig "github.com/tendermint/tendermint/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

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
		return fmt.Errorf("could not get app config values: %v", appConfErr)
	}
	tmConfig, tmConfErr := ExtractTmConfig(cmd)
	if tmConfErr != nil {
		return fmt.Errorf("could not get tendermint config values: %v", tmConfErr)
	}
	clientConfig, clientConfErr := ExtractClientConfig(cmd)
	if clientConfErr != nil {
		return fmt.Errorf("could not get client config values: %v", clientConfErr)
	}
	writeUnpackedConfig(cmd, appConfig, tmConfig, clientConfig, true)
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
	conf := serverconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("error extracting app config: %v", err)
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

// ExtractTmConfig creates a tendermint config from the command context.
func ExtractTmConfig(cmd *cobra.Command) (*tmconfig.Config, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := tmconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("error extracting tendermint config: %v", err)
	}
	conf.SetRoot(GetHomeDir(cmd))
	return conf, nil
}

// ExtractTmConfigAndMap from the command context, creates a tendermint config and related string->value map.
func ExtractTmConfigAndMap(cmd *cobra.Command) (*tmconfig.Config, FieldValueMap, error) {
	conf, err := ExtractTmConfig(cmd)
	if err != nil {
		return nil, nil, err
	}
	fields := MakeFieldValueMap(conf, true)
	removeUndesirableTmConfigEntries(fields)
	return conf, fields, nil
}

// removeUndesirableTmConfigEntries deletes some keys from the provided fields map that we don't want included.
// The provided map is altered during this call. It is also returned from this func.
// There are several fields in the tendermint config struct that don't correspond to entries in the config files.
// None of the "home" keys have entries in the config files:
// "home", "consensus.home", "mempool.home", "p2p.home", "rpc.home"
// There are several "p2p.test_" fields that should be ignored too.
// "p2p.test_dial_fail", "p2p.test_fuzz",
// "p2p.test_fuzz_config.*" ("maxdelay", "mode", "probdropconn", "probdroprw", "probsleep")
// This info is accurate in Cosmos SDK 0.43 (on 2021-08-16).
func removeUndesirableTmConfigEntries(fields FieldValueMap) FieldValueMap {
	delete(fields, "home")
	for k := range fields {
		if (len(k) > 5 && k[len(k)-5:] == ".home") || (len(k) > 9 && k[:9] == "p2p.test_") {
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
		return nil, fmt.Errorf("error extracting client config: %v", err)
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
		MakeFieldValueMap(serverconfig.DefaultConfig(), false),
		removeUndesirableTmConfigEntries(MakeFieldValueMap(tmconfig.DefaultConfig(), false)),
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
	tmConfig *tmconfig.Config,
	clientConfig *ClientConfig,
	verbose bool,
) {
	if IsPacked(cmd) {
		generateAndWritePackedConfig(cmd, appConfig, tmConfig, clientConfig, verbose)
	} else {
		writeUnpackedConfig(cmd, appConfig, tmConfig, clientConfig, verbose)
	}
}

// writeUnpackedConfig writes the provided configs to their files.
// Any config parameter provided as nil will be skipped.
// Any errors encountered will result in a panic or exit.
func writeUnpackedConfig(
	cmd *cobra.Command,
	appConfig *serverconfig.Config,
	tmConfig *tmconfig.Config,
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
		appConfigIndexEventsWorkAround(confFile, appConfig)
		if verbose {
			cmd.Printf("Done.\n")
		}
	}
	if tmConfig != nil {
		confFile := GetFullPathToTmConf(cmd)
		if verbose {
			cmd.Printf("Writing tendermint config to: %s ... ", confFile)
		}
		tmconfig.WriteConfigFile(confFile, tmConfig)
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

// appConfigIndexEventsWorkAround fixes the incorrect marshaling of the index-events config field into the toml file.
// Issue: https://github.com/cosmos/cosmos-sdk/issues/10016
// Once that issue is fixed, this can be removed.
func appConfigIndexEventsWorkAround(configFilePath string, config *serverconfig.Config) {
	// If there aren't any index events, it's marshaled just fine.
	if len(config.IndexEvents) == 0 {
		return
	}
	// Manually read in the file, change the index-events line to the correct format and write it again.
	indexEventsBz, merr := json.Marshal(config.IndexEvents)
	if merr != nil {
		panic(fmt.Errorf("marshaling index events to json: %v", merr))
	}

	bz, rerr := os.ReadFile(configFilePath)
	if rerr != nil {
		panic(fmt.Errorf("reading app config file: %v", rerr))
	}
	fixedFileBz := []byte{}
	for _, line := range strings.Split(string(bz), "\n") {
		if len(line) >= 12 && line[:12] == "index-events" {
			fixedFileBz = append(fixedFileBz, []byte("index-events = ")...)
			fixedFileBz = append(fixedFileBz, indexEventsBz...)
		} else {
			fixedFileBz = append(fixedFileBz, []byte(line)...)
		}
		fixedFileBz = append(fixedFileBz, '\n')
	}

	//nolint:gosec // These are the correct permissions
	werr := os.WriteFile(configFilePath, fixedFileBz, 0644)
	if werr != nil {
		panic(fmt.Errorf("writing fixec app config: %v", werr))
	}
}

// deleteUnpackedConfig deletes all the unpacked config files.
// An attempt will be made to remove each file before returning.
// The error returned might reflect multiple failures to delete.
// Any files that don't exist, are ignored.
func deleteUnpackedConfig(cmd *cobra.Command, verbose bool) error {
	configFiles := []string{
		GetFullPathToAppConf(cmd),
		GetFullPathToTmConf(cmd),
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
		rvErr = fmt.Errorf("one or more unpacked config files could not be removed\n%v", rvErr)
	}
	return rvErr
}

// generateAndWritePackedConfig generates the contents of the packed config file and saves it.
// Any config parameter provided as nil will be retrieved from the cmd.
// Any errors encountered will result in a panic.
func generateAndWritePackedConfig(
	cmd *cobra.Command,
	appConfig *serverconfig.Config,
	tmConfig *tmconfig.Config,
	clientConfig *ClientConfig,
	verbose bool,
) {
	mustEnsureConfigDir(cmd)
	var appConfMap, tmConfMap, clientConfMap FieldValueMap
	if appConfig == nil {
		var err error
		_, appConfMap, err = ExtractAppConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract app config values: %v", err))
		}
	} else {
		appConfMap = MakeFieldValueMap(appConfig, false)
	}
	if tmConfig == nil {
		var err error
		_, tmConfMap, err = ExtractTmConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract tm config values: %v", err))
		}
	} else {
		tmConfMap = MakeFieldValueMap(tmConfig, false)
	}
	if clientConfig == nil {
		var err error
		_, clientConfMap, err = ExtractClientConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract client config values: %v", err))
		}
	} else {
		clientConfMap = MakeFieldValueMap(clientConfig, false)
	}
	allConf := FieldValueMap{}
	allConf.AddEntriesFrom(appConfMap, tmConfMap, clientConfMap)
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
func LoadConfigFromFiles(cmd *cobra.Command) error {
	if IsPacked(cmd) {
		return loadPackedConfig(cmd)
	}
	return loadUnpackedConfig(cmd)
}

// loadUnpackedConfig attempts to read the unpacked config files and apply them to the appropriate contexts.
func loadUnpackedConfig(cmd *cobra.Command) error {
	appConfFile := GetFullPathToAppConf(cmd)
	tmConfFile := GetFullPathToTmConf(cmd)
	clientConfFile := GetFullPathToClientConf(cmd)

	// Both the server context and client context should be using the same Viper, so this is good for both.
	vpr := server.GetServerContextFromCmd(cmd).Viper

	// Load the tendermint config if it exists, or else defaults.
	switch _, err := os.Stat(tmConfFile); {
	case os.IsNotExist(err):
		lerr := addFieldMapToViper(vpr, MakeFieldValueMap(tmconfig.DefaultConfig(), false))
		if lerr != nil {
			return fmt.Errorf("tendermint config file load error: %v", lerr)
		}
	case err != nil:
		return fmt.Errorf("tendermint config file stat error: %v", err)
	default:
		vpr.SetConfigFile(tmConfFile)
		rerr := vpr.MergeInConfig()
		if rerr != nil {
			return fmt.Errorf("tendermint config file read error: %v", rerr)
		}
	}

	// Load the app/cosmos config if it exists, or else defaults.
	switch _, err := os.Stat(appConfFile); {
	case os.IsNotExist(err):
		lerr := addFieldMapToViper(vpr, MakeFieldValueMap(serverconfig.DefaultConfig(), false))
		if lerr != nil {
			return fmt.Errorf("app config load error: %v", lerr)
		}
	case err != nil:
		return fmt.Errorf("app config file stat error: %v", err)
	default:
		vpr.SetConfigFile(appConfFile)
		merr := vpr.MergeInConfig()
		if merr != nil {
			return fmt.Errorf("app config file merge error: %v", merr)
		}
	}

	// Load the client config if it exists, or else defaults.
	switch _, err := os.Stat(clientConfFile); {
	case os.IsNotExist(err):
		lerr := addFieldMapToViper(vpr, MakeFieldValueMap(DefaultClientConfig(), false))
		if lerr != nil {
			return fmt.Errorf("client config file load error: %v", lerr)
		}
	case err != nil:
		return fmt.Errorf("client config file stat error: %v", err)
	default:
		vpr.SetConfigFile(clientConfFile)
		rerr := vpr.MergeInConfig()
		if rerr != nil {
			return fmt.Errorf("client config file read error: %v", rerr)
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
		return fmt.Errorf("packed config file read error: %v", rerr)
	default:
		jerr := json.Unmarshal(packedJSON, &packedConf)
		if jerr != nil {
			return fmt.Errorf("packed config file parse error: %v", jerr)
		}
	}

	// Start with the defaults
	appConfigMap := MakeFieldValueMap(serverconfig.DefaultConfig(), false)
	tmConfigMap := MakeFieldValueMap(tmconfig.DefaultConfig(), false)
	clientConfigMap := MakeFieldValueMap(DefaultClientConfig(), false)

	// Apply the packed config entries to the defaults.
	var rvErr error
	for k, v := range packedConf {
		found := false
		if appConfigMap.Has(k) {
			found = true
			err := appConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("app config key: %s, value: %s, err: %v", k, v, err))
			}
		}
		if tmConfigMap.Has(k) {
			found = true
			err := tmConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("tendermint config key: %s, value: %s, err: %v", k, v, err))
			}
		}
		if clientConfigMap.Has(k) {
			found = true
			err := clientConfigMap.SetFromString(k, v)
			if err != nil {
				rvErr = appendError(rvErr, fmt.Errorf("client config key: %s, value: %s, err: %v", k, v, err))
			}
		}
		if !found {
			cmd.PrintErrf("unknown packed config key: %s", k)
		}
	}
	if rvErr != nil {
		return fmt.Errorf("one or more fields in the packed config could not be set\n%s", rvErr)
	}

	// Set the config values as defaults in viper.
	// Viper doesn't really have a way to directly set a config value,
	// and a set value takes precedence over flags. So I guess defaults are what we go with.
	// The server and client should both have the same viper, so we only need the one.
	vpr := server.GetServerContextFromCmd(cmd).Viper
	if lerr := addFieldMapToViper(vpr, tmConfigMap); lerr != nil {
		return fmt.Errorf("tendermint packed config load error: %v", lerr)
	}
	if lerr := addFieldMapToViper(vpr, appConfigMap); lerr != nil {
		return fmt.Errorf("app packed config load error: %v", lerr)
	}
	if lerr := addFieldMapToViper(vpr, clientConfigMap); lerr != nil {
		return fmt.Errorf("client packed config load error: %v", lerr)
	}

	return applyConfigsToContexts(cmd)
}

func addFieldMapToViper(vpr *viper.Viper, fvmap FieldValueMap) error {
	configMap := make(map[string]interface{})
	for k, v := range fvmap {
		configMap[k] = v.Interface()
	}
	// The telemetry.global-labels field in the app config struct is a `[][]string`.
	// But in serverconfig.GetConfig, it expects viper to return it as a `[]interface{}`.
	// Then each element of that is expected to also be a `[]interface{}`.
	// So we need to convert that field before adding it to viper.
	if gli, hasGL := configMap["telemetry.global-labels"]; hasGL {
		newv := make([]interface{}, 0)
		if gli != nil {
			if gl, ok := gli.([][]string); ok {
				for _, p := range gl {
					var newp []interface{}
					for _, k := range p {
						newp = append(newp, k)
					}
					newv = append(newv, newp)
				}
			}
		}
		configMap["telemetry.global-labels"] = newv
	}
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
		return fmt.Errorf("could not apply client config %s to client context - it may need to be updated manually: %v", f, err)
	}
	if err = client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
		return fmt.Errorf("could not update client context on command: %v", err)
	}

	// Set the server context's config to what Viper has now.
	serverCtx := server.GetServerContextFromCmd(cmd)
	tmConfig, err := ExtractTmConfig(cmd)
	if err != nil {
		return err
	}
	serverCtx.Config = tmConfig
	serverCtx.Config.SetRoot(clientCtx.HomeDir)

	// Set the server context's logger using what Viper has now.
	var logWriter io.Writer
	switch logFmt := serverCtx.Viper.GetString(flags.FlagLogFormat); strings.ToLower(logFmt) {
	case tmconfig.LogFormatPlain:
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	case tmconfig.LogFormatJSON:
		logWriter = os.Stderr
	default:
		return fmt.Errorf("unknown log format: %s", logFmt)
	}
	logLvlStr := serverCtx.Viper.GetString(flags.FlagLogLevel)
	logLvl, perr := zerolog.ParseLevel(strings.ToLower(logLvlStr))
	if perr != nil {
		return fmt.Errorf("failed to parse log level (%s): %w", logLvlStr, perr)
	}
	logger := zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()
	serverCtx.Logger = server.ZeroLogWrapper{Logger: logger}

	return nil
}

// appendError appends one error onto another.
// If the origErr is nil, the newErr is returned.
// Otherwise, a new error is created with a combined error message.
func appendError(origErr, newErr error) error {
	if origErr == nil {
		return newErr
	}
	return fmt.Errorf("%v\n%v", origErr, newErr)
}
