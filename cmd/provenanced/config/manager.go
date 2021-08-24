package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/spf13/cobra"

	tmconfig "github.com/tendermint/tendermint/config"
)

// PackConfig generates and saves the packed config file then removes the individual config files.
func PackConfig(cmd *cobra.Command) error {
	generateAndWritePackedConfig(cmd, nil, nil, nil, true)
	err := deleteUnpackedConfig(cmd, true)
	return err
}

// UnpackConfig generates the saves the individual config files and removes the packed config file.
func UnpackConfig(cmd *cobra.Command) error {
	appConfig, _, appConfErr := GetAppConfigAndMap(cmd)
	if appConfErr != nil {
		return fmt.Errorf("could not get app config values: %v", appConfErr)
	}
	tmConfig, _, tmConfErr := GetTmConfigAndMap(cmd)
	if tmConfErr != nil {
		return fmt.Errorf("could not get tendermint config values: %v", tmConfErr)
	}
	clientConfig, _, clientConfErr := GetClientConfigAndMap(cmd)
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
	_, err := os.Stat(GetFullPathToPackedConf(cmd))
	return err == nil
}

// GetAppConfigAndMap gets the app/cosmos configuration object and related string->value map.
func GetAppConfigAndMap(cmd *cobra.Command) (*serverconfig.Config, FieldValueMap, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := serverconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := MakeFieldValueMap(conf, true)
	return conf, fields, nil
}

// GetTmConfigAndMap gets the tendermint/config configuration object and related string->value map.
func GetTmConfigAndMap(cmd *cobra.Command) (*tmconfig.Config, FieldValueMap, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := tmconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
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

// GetClientConfigAndMap gets the client configuration object and related string->value map.
func GetClientConfigAndMap(cmd *cobra.Command) (*ClientConfig, FieldValueMap, error) {
	v := client.GetClientContextFromCmd(cmd).Viper
	conf := DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
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

// SaveConfig saves the configs to files.
// If the config is packed, any nil configs provided will extracted from the cmd.
// If the config is unpacked, only the configs provided will be written.
// Any errors encountered will result in a panic.
func SaveConfig(cmd *cobra.Command, appConfig *serverconfig.Config, tmConfig *tmconfig.Config, clientConfig *ClientConfig, verbose bool) {
	if IsPacked(cmd) {
		generateAndWritePackedConfig(cmd, appConfig, tmConfig, clientConfig, verbose)
	} else {
		writeUnpackedConfig(cmd, appConfig, tmConfig, clientConfig, verbose)
	}
}

// writeUnpackedConfig writes the provided configs to their files.
// Any config parameter provided as nil will be skipped.
// Any errors encountered will result in a panic.
func writeUnpackedConfig(cmd *cobra.Command, appConfig *serverconfig.Config, tmConfig *tmconfig.Config, clientConfig *ClientConfig, verbose bool) {
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
func generateAndWritePackedConfig(cmd *cobra.Command, appConfig *serverconfig.Config, tmConfig *tmconfig.Config, clientConfig *ClientConfig, verbose bool) {
	var appConfMap, tmConfMap, clientConfMap FieldValueMap
	if appConfig == nil {
		var err error
		_, appConfMap, err = GetAppConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract app config values: %v", err))
		}
	} else {
		appConfMap = MakeFieldValueMap(appConfig, false)
	}
	if tmConfig == nil {
		var err error
		_, tmConfMap, err = GetTmConfigAndMap(cmd)
		if err != nil {
			panic(fmt.Errorf("could not extract tm config values: %v", err))
		}
	} else {
		tmConfMap = MakeFieldValueMap(tmConfig, false)
	}
	if clientConfig == nil {
		var err error
		_, clientConfMap, err = GetClientConfigAndMap(cmd)
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
	serverCtx := server.GetServerContextFromCmd(cmd)
	serverViper := serverCtx.Viper

	// Load the tendermint defaults and config if it exists.
	for k, v := range MakeFieldValueMap(tmconfig.DefaultConfig(), false) {
		serverViper.SetDefault(k, v.Interface())
	}
	switch _, err := os.Stat(tmConfFile); {
	case os.IsNotExist(err):
		// App/cosmos config file doesn't exist. Do nothing. Just let it use the defaults.
	case err != nil:
		return fmt.Errorf("tendermint config file stat error: %v", err)
	default:
		serverViper.SetConfigFile(tmConfFile)
		rerr := serverViper.ReadInConfig()
		if rerr != nil {
			return fmt.Errorf("tendermint config file read error: %v", rerr)
		}
	}

	// Load the app/cosmos defaults and config if it exists.
	for k, v := range MakeFieldValueMap(serverconfig.DefaultConfig(), false) {
		serverViper.SetDefault(k, v.Interface())
	}
	switch _, err := os.Stat(appConfFile); {
	case os.IsNotExist(err):
		// App/cosmos config file doesn't exist. Do nothing. Just let it use the defaults.
	case err != nil:
		return fmt.Errorf("cosmos config file stat error: %v", err)
	default:
		serverViper.SetConfigFile(appConfFile)
		merr := serverViper.MergeInConfig()
		if merr != nil {
			return fmt.Errorf("cosmos config file merge error: %v", merr)
		}
	}

	clientConfFile := GetFullPathToClientConf(cmd)
	clientCtx := client.GetClientContextFromCmd(cmd)
	clientViper := clientCtx.Viper

	// Load the client defaults and config if it exists.
	for k, v := range MakeFieldValueMap(DefaultClientConfig(), false) {
		clientViper.SetDefault(k, v.Interface())
	}
	switch _, err := os.Stat(clientConfFile); {
	case os.IsNotExist(err):
		// Client config file doesn't exist. Do nothing. Just let it use the defaults.
	case err != nil:
		return fmt.Errorf("client config file stat error: %v", err)
	default:
		clientViper.SetConfigFile(clientConfFile)
		rerr := clientViper.ReadInConfig()
		if rerr != nil {
			return fmt.Errorf("client config file read error: %v", rerr)
		}
	}

	return nil
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

	// Apply the packed config differences to the defaults.
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

	// Set the tendermint and cosmos config defaults.
	serverCtx := server.GetServerContextFromCmd(cmd)
	serverViper := serverCtx.Viper
	for k, v := range tmConfigMap {
		serverViper.SetDefault(k, v.Interface())
	}
	for k, v := range appConfigMap {
		serverViper.SetDefault(k, v.Interface())
	}

	// Set the client config defaults.
	clientCtx := client.GetClientContextFromCmd(cmd)
	clientViper := clientCtx.Viper
	for k, v := range clientConfigMap {
		clientViper.SetDefault(k, v.Interface())
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
	return fmt.Errorf("%v\n%v", origErr, newErr)
}
