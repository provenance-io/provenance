package config

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	tmconfig "github.com/tendermint/tendermint/config"
	"os"
	"reflect"

	"github.com/spf13/cobra"
)

func PackConfig(cmd *cobra.Command) error {
	packedFile := GetFullPathToPackedConf(cmd)
	if _, err := os.Stat(packedFile); os.IsNotExist(err) {
		return fmt.Errorf("config is already packed: %s", packedFile)
	}
	configFiles := []string{
		GetFullPathToAppConf(cmd),
		GetFullPathToTmConf(cmd),
		GetFullPathToClientConf(cmd),
	}
	for _, f := range configFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("cannot pack config because a config file is missing: %s", f)
		}
	}
	return fmt.Errorf("not implemented")
}

// GetAppConfigAndMap gets the app/cosmos configuration object and related string->value map.
func GetAppConfigAndMap(cmd *cobra.Command) (*serverconfig.Config, map[string]reflect.Value, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := serverconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := GetFieldValueMap(conf, true)
	return conf, fields, nil
}

// GetTmConfigAndMap gets the tendermint/config configuration object and related string->value map.
func GetTmConfigAndMap(cmd *cobra.Command) (*tmconfig.Config, map[string]reflect.Value, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := tmconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := GetFieldValueMap(conf, true)
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
func removeUndesirableTmConfigEntries(fields map[string]reflect.Value) map[string]reflect.Value {
	delete(fields, "home")
	for k := range fields {
		if (len(k) > 5 && k[len(k)-5:] == ".home") || (len(k) > 9 && k[:9] == "p2p.test_") {
			delete(fields, k)
		}
	}
	return fields
}

// GetClientConfigAndMap gets the client configuration object and related string->value map.
func GetClientConfigAndMap(cmd *cobra.Command) (*ClientConfig, map[string]reflect.Value, error) {
	v := client.GetClientContextFromCmd(cmd).Viper
	conf := DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := GetFieldValueMap(conf, true)
	return conf, fields, nil
}

// GetAllConfigDefaults gets a field map from the defaults of all the configs.
func GetAllConfigDefaults() map[string]reflect.Value {
	defaultMaps := []map[string]reflect.Value{
		GetFieldValueMap(serverconfig.DefaultConfig(), false),
		removeUndesirableTmConfigEntries(GetFieldValueMap(tmconfig.DefaultConfig(), false)),
		GetFieldValueMap(DefaultClientConfig(), false),
	}
	rv := map[string]reflect.Value{}
	for _, m := range defaultMaps {
		for k, v := range m {
			rv[k] = v
		}
	}
	return rv
}


