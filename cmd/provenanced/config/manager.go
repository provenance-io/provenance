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

func PackConfig(cmd *cobra.Command) error {
	packedFile := GetFullPathToPackedConf(cmd)
	configFiles := []string{
		GetFullPathToAppConf(cmd),
		GetFullPathToTmConf(cmd),
		GetFullPathToClientConf(cmd),
	}
	if _, err := os.Stat(packedFile); err == nil || !os.IsNotExist(err) {
		return fmt.Errorf("config is already packed: %s", packedFile)
	}
	allCurrent := FieldValueMap{}
	if _, confMap, err := GetAppConfigAndMap(cmd); err != nil {
		return err
	} else {
		allCurrent.AddEntriesFrom(confMap)
	}
	if _, confMap, err := GetTmConfigAndMap(cmd); err != nil {
		return err
	} else {
		allCurrent.AddEntriesFrom(confMap)
	}
	if _, confMap, err := GetClientConfigAndMap(cmd); err != nil {
		return err
	} else {
		allCurrent.AddEntriesFrom(confMap)
	}
	allDefaults := GetAllConfigDefaults()
	packed := map[string]string{}
	for key, info := range MakeUpdatedFieldMap(allDefaults, allCurrent, true) {
		packed[key] = unquote(info.IsNow)
	}
	prettyJson, err := json.MarshalIndent(packed, "", "  ")
	if err != nil {
		return err
	}
	cmd.Printf("Packed config:\n%s\n", prettyJson)
	err = os.WriteFile(packedFile, prettyJson, 0644)
	if err != nil {
		return err
	}
	cmd.Printf("Saved to: %s\n", packedFile)
	hadRmErr := false
	for _, f := range configFiles {
		if rmErr := os.Remove(f); rmErr != nil && !os.IsNotExist(rmErr) {
			hadRmErr = true
			cmd.Printf("Error removing file: %v\n", rmErr)
		}
	}
	if hadRmErr {
		return fmt.Errorf("One or more config files could not be removed.")
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
	defaultMaps := []FieldValueMap{
		MakeFieldValueMap(serverconfig.DefaultConfig(), false),
		removeUndesirableTmConfigEntries(MakeFieldValueMap(tmconfig.DefaultConfig(), false)),
		MakeFieldValueMap(DefaultClientConfig(), false),
	}
	rv := FieldValueMap{}
	for _, m := range defaultMaps {
		rv.AddEntriesFrom(m)
	}
	return rv
}
