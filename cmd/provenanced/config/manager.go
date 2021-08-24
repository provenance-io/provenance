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
			if rvErr == nil {
				rvErr = err
			} else {
				rvErr = fmt.Errorf("%v\n%v", rvErr, err)
			}
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
