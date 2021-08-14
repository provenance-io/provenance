package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	tmconfig "github.com/tendermint/tendermint/config"
)

// Cmd returns a CLI command to interactively create an application CLI
// config file.
func ClientConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [<key> [value]] | get {client|app|config|all|<key1> [<key2> ...]} | set <key1> <value1> [<key2> <value2> ...]",
		Short: "Get or Set configuration values",
		RunE:  runConfigCmd,
	}
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "get":
			return runConfigGetCmd(cmd, args[1:])
		case "set":
			return runConfigSetCmd(cmd, args[1:])
		}
	}
	switch len(args) {
	case 0, 1:
		return runConfigGetCmd(cmd, args)
	case 2:
		return runConfigSetCmd(cmd, args)
	}
	return errors.New("when more than two arguments are provided, the first must either be \"get\" or \"set\"")
}

func runConfigGetCmd(cmd *cobra.Command, args []string) error {
	_, appFields, acerr := getAppConfigAndMap(cmd)
	if acerr != nil {
		return fmt.Errorf("couldn't get app config: %v", acerr)
	}
	_, tmFields, tmcerr := getTmConfigAndMap(cmd)
	if tmcerr != nil {
		return fmt.Errorf("couldn't get tendermint config: %v", tmcerr)
	}
	_, clientFields, ccerr := getClientConfigAndMap(cmd)
	if ccerr != nil {
		return fmt.Errorf("couldn't get client config: %v", ccerr)
	}

	if len(args) == 0 {
		args = append(args, "all")
	}
	for _, key := range args {
		switch key {
		case "all":
			cmd.Println("App Config")
			cmd.Println("----------")
			cmd.Println(getFieldMapString(appFields))
			cmd.Println("Tendermint Config")
			cmd.Println("-----------------")
			cmd.Println(getFieldMapString(tmFields))
			cmd.Println("Client Config")
			cmd.Println("-------------")
			cmd.Println(getFieldMapString(clientFields))
		case "app":
			cmd.Println("App Config")
			cmd.Println("----------")
			cmd.Println(getFieldMapString(appFields))
		case "config", "tendermint", "tm":
			cmd.Println("Tendermint Config")
			cmd.Println("-----------------")
			cmd.Println(getFieldMapString(tmFields))
		case "client":
			cmd.Println("Client Config")
			cmd.Println("-------------")
			cmd.Println(getFieldMapString(clientFields))
		default:
			v, ok := appFields[key]
			if !ok {
				v, ok = tmFields[key]
			}
			if !ok {
				v, ok = clientFields[key]
			}
			if ok {
				cmd.Println(getKVString(key, v))
			} else {
				cmd.Printf("Configuration key %s does not exist.\n", key)
			}
		}
	}
	return nil
}

func runConfigSetCmd(cmd *cobra.Command, args []string) error {
	appConfig, appFields, acerr := getAppConfigAndMap(cmd)
	if acerr != nil {
		return fmt.Errorf("couldn't get app config: %v", acerr)
	}
	tmConfig, tmFields, tmcerr := getTmConfigAndMap(cmd)
	if tmcerr != nil {
		return fmt.Errorf("couldn't get tendermint config: %v", tmcerr)
	}
	clientConfig, clientFields, ccerr := getClientConfigAndMap(cmd)
	if ccerr != nil {
		return fmt.Errorf("couldn't get client config: %v", ccerr)
	}

	if len(args) == 0 {
		return errors.New("no key/value pairs provided")
	}
	if len(args)%2 != 0 {
		return errors.New("an even number of arguments are required when setting values")
	}
	keyCount := len(args) / 2
	keys := make([]string, keyCount)
	vals := make([]string, keyCount)
	for i := 0; i < keyCount; i++ {
		keys[i] = args[i*2]
		vals[i] = args[i*2+1]
	}
	issueFound := false
	appUpdated, tmUpdated, clientUpdated := false, false, false
	for i, key := range keys {
		v, ok := appFields[key]
		if ok {
			appUpdated = true
		} else {
			v, ok = tmFields[key]
			if ok {
				tmUpdated = true
			} else {
				v, ok = clientFields[key]
				if ok {
					clientUpdated = true
				}
			}
		}
		if !ok {
			cmd.Printf("Configuration key %s does not exist.\n", key)
			issueFound = true
		} else {
			err := setValueFromString(key, v, vals[i])
			if err != nil {
				cmd.Printf("Error setting key %s: %s", key, err.Error())
				issueFound = true
			}
		}
	}
	if issueFound {
		return errors.New("one or more issues encountered; no configuration values have been updated")
	}
	configPath := filepath.Join(client.GetClientContextFromCmd(cmd).HomeDir, "config")
	if appUpdated {
		configFilename := filepath.Join(configPath, "app.toml")
		serverconfig.WriteConfigFile(configFilename, appConfig)
		cmd.Printf("App config updated: %s\n", configFilename)
	}
	if tmUpdated {
		configFilename := filepath.Join(configPath, "config.toml")
		tmconfig.WriteConfigFile(configFilename, tmConfig)
		cmd.Printf("Tendermint config updated: %s\n", configFilename)
	}
	if clientUpdated {
		configFilename := filepath.Join(configPath, "client.toml")
		provconfig.WriteConfigToFile(configFilename, clientConfig)
		cmd.Printf("Client config updated: %s\n", configFilename)
	}
	return nil
}

// getTmConfig gets the tendermint configuration.
func getAppConfigAndMap(cmd *cobra.Command) (*serverconfig.Config, map[string]reflect.Value, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	var conf *serverconfig.Config
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)
	// TODO add special handling for some cosmos config variables.
	// These fields need some special handling.
	// app: global-labels - [][]string (each sub-slice has 2 elements)
	// app: index-events - []string

	return conf, fields, nil
}

// getTmConfig gets the tendermint configuration.
func getTmConfigAndMap(cmd *cobra.Command) (*tmconfig.Config, map[string]reflect.Value, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	var conf *tmconfig.Config
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)
	// TODO add special handling for some tendermint config variables.
	// These fields need some special handling.
	// tm: rpc.cors_allowed_headers - []string
	// tm: rpc.cors_allowed_methods - []string
	// tm: rpc.cors_allowed_origins - []string

	return conf, fields, nil
}

func getClientConfigAndMap(cmd *cobra.Command) (*provconfig.ClientConfig, map[string]reflect.Value, error) {
	v := client.GetClientContextFromCmd(cmd).Viper
	var conf *provconfig.ClientConfig
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)
	return conf, fields, nil
}

// getFieldMapString gets a multi-line string with all the keys and values in the provided map.
func getFieldMapString(m map[string]reflect.Value) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(getKVString(k, m[k]))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// getKVString gets a string associating a key and value.
func getKVString(k string, v reflect.Value) string {
	return fmt.Sprintf("%s=%s", k, getStringFromValue(v))
}

// getStringFromValue gets a string of the given value.
// This creates strings that are more in line with what the values look like in the config files.
// For slices and arrays, it turns into `["a", "b", "c"]`.
// For strings, it turns into `"a"`.
// For anything else, it just uses fmt %v.
// This wasn't designed with the following kinds in mind:
//    Invalid, Chan, Func, Interface, Map, Ptr, Struct, or UnsafePointer.
func getStringFromValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		var sb strings.Builder
		sb.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(getStringFromValue(v.Index(i)))
		}
		sb.WriteByte(']')
		return sb.String()
	case reflect.String:
		return fmt.Sprintf("\"%v\"", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// setValueFromString sets a value from the provided string.
// The string is converted appropriately for the underlying value type.
// Assuming the value came from GetFieldValueMap, this will actually be updating the
// value in the config object provided to that function.
func setValueFromString(fieldName string, fieldVal reflect.Value, strVal string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(strVal)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(strVal)
		if err != nil {
			return err
		}
		fieldVal.SetBool(b)
		return nil
	default:
		return fmt.Errorf("field %s cannot be set because setting values of type %s has not yet been set up", fieldName, fieldVal.Kind())
	}
}
