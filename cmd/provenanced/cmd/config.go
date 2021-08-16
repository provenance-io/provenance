package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	tmconfig "github.com/tendermint/tendermint/config"
)

const entryNotFound = -1

// Cmd returns a CLI command to interactively create an application CLI
// config file.
func ClientConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [<key> [value]] | get {client|app|config|all|<key1> [<key2> ...]} | set <key1> <value1> [<key2> <value2> ...]",
		Short: "Get or Set configuration values",
		RunE:  runConfigCmd,
	}
	// TODO: add Long: and possibly Example:
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

	configPath := filepath.Join(client.GetClientContextFromCmd(cmd).HomeDir, "config")

	if len(args) == 0 {
		args = append(args, "all")
	}
	toOutput := map[string]reflect.Value{}
	for _, key := range args {
		switch key {
		case "all":
			cmd.Println("App Config: ", filepath.Join(configPath, "app.toml"))
			cmd.Println("---------------")
			cmd.Println(getFieldMapString(appFields))
			cmd.Println("Tendermint Config: ", filepath.Join(configPath, "config.toml"))
			cmd.Println("----------------------")
			cmd.Println(getFieldMapString(tmFields))
			cmd.Println("Client Config: ", filepath.Join(configPath, "client.toml"))
			cmd.Println("------------------")
			cmd.Println(getFieldMapString(clientFields))
		case "app":
			cmd.Println("App Config: ", filepath.Join(configPath, "app.toml"))
			cmd.Println("---------------")
			cmd.Println(getFieldMapString(appFields))
		case "config", "tendermint", "tm":
			cmd.Println("Tendermint Config: ", filepath.Join(configPath, "config.toml"))
			cmd.Println("----------------------")
			cmd.Println(getFieldMapString(tmFields))
		case "client":
			cmd.Println("Client Config: ", filepath.Join(configPath, "client.toml"))
			cmd.Println("------------------")
			cmd.Println(getFieldMapString(clientFields))
		default:
			entries := findEntries(key, appFields, tmFields, clientFields)
			if len(entries) == 0 {
				cmd.Printf("Configuration key %s does not exist.\n", key)
			} else {
				for k, v := range entries {
					toOutput[k] = v
				}
			}
		}
	}
	if len(toOutput) > 0 {
		cmd.Println(getFieldMapString(toOutput))
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
	configPath := filepath.Join(client.GetClientContextFromCmd(cmd).HomeDir, "config")
	for i, key := range keys {
		// Bug: As of Cosmos 0.43 (and 2021-08-16), the app config's index-events configuration value isn't properly marshalled into the config.
		// For example,
		//   appConfig.IndexEvents = []string{"a", "b"}
		//   serverconfig.WriteConfigFile(filename, appConfig)
		// works without error but the configuration file will have
		//   index-events = [a b]
		// instead of what is needed:
		//   index-events = ["a", "b"]
		// This results in that config file being invalid and no longer loadable:
		//   failed to merge configuration: While parsing config: (61, 17): no value can start with a
		// So for now, if someone requests the setting of that field, return an error with some helpful info.
		if key == "index-events" {
			cmd.Printf("The index-events list cannot be set with this command. It can be manually updated in %s\n",
				filepath.Join(configPath, "app.toml"))
			issueFound = true
			continue
		}
		v, foundIn := findEntry(key, appFields, tmFields, clientFields)
		if foundIn == entryNotFound {
			cmd.Printf("Configuration key %s does not exist.\n", key)
			issueFound = true
		} else {
			err := setValueFromString(key, v, vals[i])
			if err != nil {
				cmd.Printf("Error setting key %s: %v\n", key, err)
				issueFound = true
			} else {
				switch foundIn {
				case 0:
					appUpdated = true
				case 1:
					tmUpdated = true
				case 2:
					clientUpdated = true
				}
			}
		}
	}
	if !issueFound {
		if appUpdated {
			if err := appConfig.ValidateBasic(); err != nil {
				cmd.Printf("App config validation error: %v\n", err)
				issueFound = true
			}
		}
		if tmUpdated {
			if err := tmConfig.ValidateBasic(); err != nil {
				cmd.Printf("Tendermint config validation error: %v\n", err)
				issueFound = true
			}
		}
		if clientUpdated {
			if err := clientConfig.ValidateBasic(); err != nil {
				cmd.Printf("Client config validation error: %v\n", err)
				issueFound = true
			}
		}
	}
	if issueFound {
		return errors.New("one or more issues encountered; no configuration values have been updated")
	}
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
	conf := serverconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)

	return conf, fields, nil
}

// getTmConfig gets the tendermint configuration.
func getTmConfigAndMap(cmd *cobra.Command) (*tmconfig.Config, map[string]reflect.Value, error) {
	v := server.GetServerContextFromCmd(cmd).Viper
	conf := tmconfig.DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)
	// There are several fields in this map that don't correspond to entries in the config files.
	// This info is accurate in Cosmos SDK 0.43 (on 2021-08-16).
	// None of the "home" keys have entries in the config files:
	// "home", "consensus.home", "mempool.home", "p2p.home", "rpc.home"
	// There are several "p2p.test_" fields that should be ignored too.
	// "p2p.test_dial_fail", "p2p.test_fuzz",
	// "p2p.test_fuzz_config.*" ("maxdelay", "mode", "probdropconn", "probdroprw", "probsleep")
	delete(fields, "home")
	for k := range fields {
		if (len(k) > 5 && k[len(k)-5:] == ".home") || (len(k) > 9 && k[:9] == "p2p.test_") {
			delete(fields, k)
		}
	}

	return conf, fields, nil
}

func getClientConfigAndMap(cmd *cobra.Command) (*provconfig.ClientConfig, map[string]reflect.Value, error) {
	v := client.GetClientContextFromCmd(cmd).Viper
	conf := provconfig.DefaultClientConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, nil, err
	}
	fields := provconfig.GetFieldValueMap(conf, true)
	return conf, fields, nil
}

// findEntry gets the entry with the given key in one of the provided maps.
// Maps are searched in the order provided and the first match is returned.
// The second return value is the index of the provided map that the entry was found in (starting with 0).
// If it's equal to entryNotFound, the entry wasn't found.
func findEntry(key string, maps ...map[string]reflect.Value) (reflect.Value, int) {
	for i, m := range maps {
		if v, ok := m[key]; ok {
			return v, i
		}
	}
	return reflect.Value{}, entryNotFound
}

// findEntries gets entries that match a given key from the provided maps.
// If the key doesn't end in a period, findEntry is used first to find an exact match.
// If an exact match isn't found, a period is appended to the key (unless already there) and sub-entry matches are looked for.
// E.g. Providing "filter_peers" will get just the "filter_peers" entry.
// Providing "consensus." will bypass the exact key lookup, and return all fields that start with "consensus.".
// Providing "consensus" will look first for a field specifically called "consensus",
// then, if/when not found, will return all fields that start with "consensus.".
func findEntries(key string, maps ...map[string]reflect.Value) map[string]reflect.Value {
	rv := map[string]reflect.Value{}
	baseKey := key
	if len(key) == 0 {
		return rv
	}
	if key[len(key)-1:] != "." {
		if v, i := findEntry(key, maps...); i != entryNotFound {
			rv[key] = v
			return rv
		}
		baseKey+="."
	}
	baseKeyLen := len(baseKey)
	for _, m := range maps {
		for k, v := range m {
			if len(k) > baseKeyLen && k[:baseKeyLen] == baseKey {
				rv[k] = v
			}
		}
	}
	return rv
}

// getFieldMapString gets a multi-line string with all the keys and values in the provided map.
func getFieldMapString(m map[string]reflect.Value) string {
	baseKeys := []string{}
	subKeys := []string{}
	for k := range m {
		if strings.Contains(k, ".") {
			subKeys = append(subKeys, k)
		} else {
			baseKeys = append(baseKeys, k)
		}
	}
	sort.Strings(baseKeys)
	sort.Strings(subKeys)
	keys := make([]string, 0, len(m))
	keys = append(keys, baseKeys...)
	keys = append(keys, subKeys...)
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
	case reflect.Int64:
		if v.Type().String() == "time.Duration" {
			return fmt.Sprintf("\"%v\"", v)
		}
		return fmt.Sprintf("%v", v)
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
	case reflect.Int:
		i, err := strconv.Atoi(strVal)
		if err != nil {
			return err
		}
		fieldVal.SetInt(int64(i))
		return nil
	case reflect.Int64:
		if fieldVal.Type().String() == "time.Duration" {
			i, err := time.ParseDuration(strVal)
			if err != nil {
				return err
			}
			fieldVal.SetInt(int64(i))
			return nil
		}
		i, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetInt(i)
		return nil
	case reflect.Int32:
		i, err := strconv.ParseInt(strVal, 10, 32)
		if err != nil {
			return err
		}
		fieldVal.SetInt(i)
		return nil
	case reflect.Int16:
		i, err := strconv.ParseInt(strVal, 10, 16)
		if err != nil {
			return err
		}
		fieldVal.SetInt(i)
		return nil
	case reflect.Int8:
		i, err := strconv.ParseInt(strVal, 10, 8)
		if err != nil {
			return err
		}
		fieldVal.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint64:
		ui, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetUint(ui)
		return nil
	case reflect.Uint32:
		ui, err := strconv.ParseUint(strVal, 10, 32)
		if err != nil {
			return err
		}
		fieldVal.SetUint(ui)
		return nil
	case reflect.Uint16:
		ui, err := strconv.ParseUint(strVal, 10, 16)
		if err != nil {
			return err
		}
		fieldVal.SetUint(ui)
		return nil
	case reflect.Uint8:
		ui, err := strconv.ParseUint(strVal, 10, 8)
		if err != nil {
			return err
		}
		fieldVal.SetUint(ui)
		return nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return err
		}
		fieldVal.SetFloat(f)
		return nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(strVal, 32)
		if err != nil {
			return err
		}
		fieldVal.SetFloat(f)
		return nil
	case reflect.Slice:
		switch fieldVal.Type().Elem().Kind() {
		case reflect.String:
			val := []string{}
			if len(strVal) > 0 {
				err := json.Unmarshal([]byte(strVal), &val)
				if err != nil {
					return err
				}
			}
			fieldVal.Set(reflect.ValueOf(val))
			return nil
		case reflect.Slice:
			if fieldVal.Type().Elem().Elem().Kind() == reflect.String {
				val := [][]string{}
				if len(strVal) > 0 {
					err := json.Unmarshal([]byte(strVal), &val)
					if err != nil {
						return err
					}
				}
				if fieldName == "telemetry.global-labels" {
					// The Cosmos config ValidateBasic doesn't do this checking (as of Cosmos 0.43, 2021-08-16).
					// If the length of a sub-slice is 0 or 1, you get a panic:
					//   panic: template: appConfigFileTemplate:95:26: executing "appConfigFileTemplate" at <index $v 1>: error calling index: reflect: slice index out of range
					// If the length of a sub-slice is greater than 2, everything after the first two ends up getting chopped off.
					// e.g. trying to set it to '[["a","b","c"]]' will actually end up just setting it to '[["a","b"]]'.
					for i, s := range val {
						if len(s) != 2 {
							return fmt.Errorf("invalid %s: sub-arrays must have length 2, but the sub-array at index %d has %d", fieldName, i, len(s))
						}
					}
				}
				fieldVal.Set(reflect.ValueOf(val))
				return nil
			}
		}
	}
	return fmt.Errorf("field %s cannot be set because setting values of type %s has not yet been set up", fieldName, fieldVal.Type())
}
