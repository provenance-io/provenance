package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	provconfig "github.com/provenance-io/provenance/cmd/provenanced/config"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/version"

	tmconfig "github.com/tendermint/tendermint/config"
)

const (
	// entryNotFound is a magic index value that indicates a thing wasn't found (and so no index is applicable).
	entryNotFound = -1
)

var configCmdStart = fmt.Sprintf("%s config", version.AppName)

// updatedField is a struct holding information about a config field that has been updated.
type updatedField struct {
	Key   string
	Was   string
	IsNow string
}

// Update updates the base updatedField given information in the provided newerInfo.
func (u *updatedField) Update(newerInfo updatedField) {
	u.IsNow = newerInfo.IsNow
}

// String converts an updatedField to a string similar to using %#v but a little cleaner.
func (u updatedField) String() string {
	return fmt.Sprintf(`updatedField{Key:%s, Was:%s, IsNow:%s}`, u.Key, u.Was, u.IsNow)
}

// StringAsUpdate creates a string from this updatedField indicating a change has being made.
func (u updatedField) StringAsUpdate() string {
	return fmt.Sprintf("%s Was: %s, Is Now: %s", u.Key, u.Was, u.IsNow)
}

// StringAsDefault creates a string from this updatedField identifying the Was as a default.
func (u updatedField) StringAsDefault() string {
	if !u.HasDiff() {
		return fmt.Sprintf("%s=%s (same as default)", u.Key, u.IsNow)
	}
	return fmt.Sprintf("%s=%s (default=%s)", u.Key, u.IsNow, u.Was)
}

// HasDiff returns true if IsNow and Was have different values.
func (u updatedField) HasDiff() bool {
	return u.IsNow != u.Was
}

// AddToOrUpdateIn adds this updatedField to the provided map if it's not already there (by Key).
// If it's already there, the map's existing entry is updated with info in this updatedField.
func (u updatedField) AddToOrUpdateIn(all map[string]*updatedField) {
	if inf, ok := all[u.Key]; ok {
		inf.Update(u)
	} else {
		all[u.Key] = &u
	}
}

// ConfigCmd returns a CLI command to update config files.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "config",
		Aliases:                    []string{"conf"},
		Short:                      "Manage configuration values",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		ConfigGetCmd(),
		ConfigSetCmd(),
		ConfigChangedCmd(),
		ConfigPackCmd(),
		ConfigUnpackCmd(),
	)
	return cmd
}

func ConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [<key1> [<key2> ...]]",
		Short: "Get configuration values",
		Long: fmt.Sprintf(`Get configuration values.

    The key values can be specific.
        e.g. %[1]s get telemetry.service-name moniker.
    Or they can be parent field names
        e.g. %[1]s get api consensus
    Or they can be a type of config file:
        "cosmos", "app" -> %[2]s configuration values.
            e.g. %[1]s get app
        "tendermint", "tm", "config" -> %[3]s configuration values.
            e.g. %[1]s get tm
        "client" -> %[4]s configuration values.
            e.g. %[1]s get client
    Or they can be the word "all" to get all configuration values.
        e.g. %[1]s get all
    If no keys are provided, all values are retrieved.

`, configCmdStart, provconfig.AppConfFilename, provconfig.TmConfFilename, provconfig.ClientConfFilename),
		RunE: func(cmd *cobra.Command, args []string) error {
			showHelp, err := runConfigGetCmd(cmd, args)
			return handleNoShowHelp(cmd, showHelp, err)
		},
	}
	return cmd
}

func ConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key1> <value1> [<key2> <value2> ...]",
		Short: "Set configuration values",
		Long: fmt.Sprintf(`Set configuration values.

Set a config value: %[1]s set <key> <value>
    The key must be specific, e.g. "telemetry.service-name", or "moniker".
    The value must be provided as a single, separate argument.
    e.g. %[1]s set output json

Set multiple config values %[1]s set <key1> <value1> [<key2> <value2> ...]
    Simply provide multiple key/value pairs as alternating arguments.
    e.g. %[1]s set api.enable true api.swagger true

`, configCmdStart),
		RunE: func(cmd *cobra.Command, args []string) error {
			showHelp, err := runConfigSetCmd(cmd, args)
			return handleNoShowHelp(cmd, showHelp, err)
		},
	}
	return cmd
}

func ConfigChangedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changed [<key1> [<key2>...]",
		Short: "Get configuration values that are different from their default.",
		Long: fmt.Sprintf(`Get configuration values that are different from their default.

Get just the configuration entries that are not default values: %[1]s changed [<key1> [<key2> ...]]
    The key values can be specific.
        e.g. %[1]s get telemetry.service-name moniker.
        Specific keys that are provided that are equal to default values will still be included in output,
            but they will be noted as such.
    Or they can be parent field names
        e.g. %[1]s get api consensus
    Or they can be a type of config file:
        "cosmos", "app" -> %[2]s configuration values.
            e.g. %[1]s get app
        "tendermint", "tm", "config" -> %[3]s configuration values.
            e.g. %[1]s get tm
        "client" -> %[4]s configuration values.
            e.g. %[1]s get client
    Or they can be the word "all" to get all configuration values.
        e.g. %[1]s get all
    Current and default values are both included in the output.
    If no keys are provided, all non-default values are retrieved.

`, configCmdStart, provconfig.AppConfFilename, provconfig.TmConfFilename, provconfig.ClientConfFilename),
		RunE: func(cmd *cobra.Command, args []string) error {
			showHelp, err := runConfigChangedCmd(cmd, args)
			return handleNoShowHelp(cmd, showHelp, err)
		},
	}
	return cmd
}

func ConfigPackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pack",
		Short: "Unpack configuration into a single config file",
		Long: fmt.Sprintf(`Unpack configuration into a single config file

Combines the %[2]s, %[3]s, and %[4]s files into %[1]s.

`, provconfig.PackedConfFilename, provconfig.AppConfFilename, provconfig.TmConfFilename, provconfig.ClientConfFilename),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigPackCmd(cmd)
		},
	}
	return cmd
}

func ConfigUnpackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpack",
		Short: "Unpack configuration into separate config files",
		Long: fmt.Sprintf(`Unpack configuration into separate config files.

Splits the %[1]s file into %[2]s, %[3]s, and %[4]s.
Default values are filled in appropriately.

`, provconfig.PackedConfFilename, provconfig.AppConfFilename, provconfig.TmConfFilename, provconfig.ClientConfFilename),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigUnpackCmd(cmd)
		},
	}
	return cmd
}

func handleNoShowHelp(cmd *cobra.Command, showHelp bool, err error) error {
	// Note: If a RunE returns an error, the usage information is displayed.
	//       That ends up being kind of annoying in most cases in here.
	//       So only return the error when extra help is desired.
	if err != nil {
		if showHelp {
			return err
		}
		cmd.Printf("Error: %v\n", err)
	}
	return nil
}

// runConfigGetCmd gets requested values and outputs them.
// The first return value is whether or not to include help with the output of an error.
// This will only ever be true if an error is also returned.
// The second return value is any error encountered.
func runConfigGetCmd(cmd *cobra.Command, args []string) (bool, error) {
	_, appFields, acerr := provconfig.GetAppConfigAndMap(cmd)
	if acerr != nil {
		return false, fmt.Errorf("couldn't get app config: %v", acerr)
	}
	_, tmFields, tmcerr := provconfig.GetTmConfigAndMap(cmd)
	if tmcerr != nil {
		return false, fmt.Errorf("couldn't get tendermint config: %v", tmcerr)
	}
	_, clientFields, ccerr := provconfig.GetClientConfigAndMap(cmd)
	if ccerr != nil {
		return false, fmt.Errorf("couldn't get client config: %v", ccerr)
	}

	if len(args) == 0 {
		args = append(args, "all")
	}

	appToOutput := map[string]reflect.Value{}
	tmToOutput := map[string]reflect.Value{}
	clientToOutput := map[string]reflect.Value{}
	unknownKeyMap := map[string]reflect.Value{}
	for _, key := range args {
		switch key {
		case "all":
			for k, v := range appFields {
				appToOutput[k] = v
			}
			for k, v := range tmFields {
				tmToOutput[k] = v
			}
			for k, v := range clientFields {
				clientToOutput[k] = v
			}
		case "app", "cosmos":
			for k, v := range appFields {
				appToOutput[k] = v
			}
		case "config", "tendermint", "tm":
			for k, v := range tmFields {
				tmToOutput[k] = v
			}
		case "client":
			for k, v := range clientFields {
				clientToOutput[k] = v
			}
		default:
			entries, foundIn := findEntries(key, appFields, tmFields, clientFields)
			switch foundIn {
			case 0:
				for k, v := range entries {
					appToOutput[k] = v
				}
			case 1:
				for k, v := range entries {
					tmToOutput[k] = v
				}
			case 2:
				for k, v := range entries {
					clientToOutput[k] = v
				}
			default:
				unknownKeyMap[key] = reflect.Value{}
			}
		}
	}

	if len(appToOutput) > 0 {
		cmd.Println(makeAppConfigHeader(cmd, ""))
		cmd.Println(makeFieldMapString(appToOutput))
	}
	if len(tmToOutput) > 0 {
		cmd.Println(makeTmConfigHeader(cmd, ""))
		cmd.Println(makeFieldMapString(tmToOutput))
	}
	if len(clientToOutput) > 0 {
		cmd.Println(makeClientConfigHeader(cmd, ""))
		cmd.Println(makeFieldMapString(clientToOutput))
	}
	if len(unknownKeyMap) > 0 {
		unknownKeys := getSortedKeys(unknownKeyMap)
		s := "s"
		if len(unknownKeys) == 1 {
			s = ""
		}
		return false, fmt.Errorf("%d configuration key%s not found: %s", len(unknownKeys), s, strings.Join(unknownKeys, ", "))
	}
	return false, nil
}

// runConfigSetCmd sets values as provided.
// The first return value is whether or not to include help with the output of an error.
// This will only ever be true if an error is also returned.
// The second return value is any error encountered.
func runConfigSetCmd(cmd *cobra.Command, args []string) (bool, error) {
	appConfig, appFields, acerr := provconfig.GetAppConfigAndMap(cmd)
	if acerr != nil {
		return false, fmt.Errorf("couldn't get app config: %v", acerr)
	}
	tmConfig, tmFields, tmcerr := provconfig.GetTmConfigAndMap(cmd)
	if tmcerr != nil {
		return false, fmt.Errorf("couldn't get tendermint config: %v", tmcerr)
	}
	clientConfig, clientFields, ccerr := provconfig.GetClientConfigAndMap(cmd)
	if ccerr != nil {
		return false, fmt.Errorf("couldn't get client config: %v", ccerr)
	}

	if len(args) == 0 {
		return true, errors.New("no key/value pairs provided")
	}
	if len(args)%2 != 0 {
		return true, errors.New("an even number of arguments are required when setting values")
	}
	keyCount := len(args) / 2
	keys := make([]string, keyCount)
	vals := make([]string, keyCount)
	for i := 0; i < keyCount; i++ {
		keys[i] = args[i*2]
		vals[i] = args[i*2+1]
	}
	issueFound := false
	appUpdates := map[string]*updatedField{}
	tmUpdates := map[string]*updatedField{}
	clientUpdates := map[string]*updatedField{}
	for i, key := range keys {
		// Bug: As of Cosmos 0.43 (and 2021-08-16), the app config's index-events configuration value isn't properly marshaled into the config.
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
				provconfig.GetFullPathToAppConf(cmd))
			issueFound = true
			continue
		}
		v, foundIn := findEntry(key, appFields, tmFields, clientFields)
		if foundIn == entryNotFound {
			cmd.Printf("Configuration key %s does not exist.\n", key)
			issueFound = true
			continue
		}
		was := getStringFromValue(v)
		err := setValueFromString(key, v, vals[i])
		if err != nil {
			cmd.Printf("Error setting key %s: %v\n", key, err)
			issueFound = true
			continue
		}
		info := updatedField{
			Key:   key,
			Was:   was,
			IsNow: getStringFromValue(v),
		}
		switch foundIn {
		case 0:
			info.AddToOrUpdateIn(appUpdates)
		case 1:
			info.AddToOrUpdateIn(tmUpdates)
		case 2:
			info.AddToOrUpdateIn(clientUpdates)
		}
	}
	if !issueFound {
		if len(appUpdates) > 0 {
			if err := appConfig.ValidateBasic(); err != nil {
				cmd.Printf("App config validation error: %v\n", err)
				issueFound = true
			}
		}
		if len(tmUpdates) > 0 {
			if err := tmConfig.ValidateBasic(); err != nil {
				cmd.Printf("Tendermint config validation error: %v\n", err)
				issueFound = true
			}
		}
		if len(clientUpdates) > 0 {
			if err := clientConfig.ValidateBasic(); err != nil {
				cmd.Printf("Client config validation error: %v\n", err)
				issueFound = true
			}
		}
	}
	if issueFound {
		return false, errors.New("one or more issues encountered; no configuration values have been updated")
	}
	if len(appUpdates) > 0 {
		serverconfig.WriteConfigFile(provconfig.GetFullPathToAppConf(cmd), appConfig)
		cmd.Println(makeAppConfigHeader(cmd, "Updated"))
		cmd.Println(makeUpdatedFieldMapString(appUpdates, updatedField.StringAsUpdate))
	}
	if len(tmUpdates) > 0 {
		tmconfig.WriteConfigFile(provconfig.GetFullPathToTmConf(cmd), tmConfig)
		cmd.Println(makeTmConfigHeader(cmd, "Updated"))
		cmd.Println(makeUpdatedFieldMapString(tmUpdates, updatedField.StringAsUpdate))
	}
	if len(clientUpdates) > 0 {
		provconfig.WriteConfigToFile(provconfig.GetFullPathToClientConf(cmd), clientConfig)
		cmd.Println(makeClientConfigHeader(cmd, "Updated"))
		cmd.Println(makeUpdatedFieldMapString(clientUpdates, updatedField.StringAsUpdate))
	}
	return false, nil
}

// runConfigChangedCmd gets values that have changed from their defaults.
// The first return value is whether or not to include help with the output of an error.
// This will only ever be true if an error is also returned.
// The second return value is any error encountered.
func runConfigChangedCmd(cmd *cobra.Command, args []string) (bool, error) {
	_, appFields, acerr := provconfig.GetAppConfigAndMap(cmd)
	if acerr != nil {
		return false, fmt.Errorf("couldn't get app config: %v", acerr)
	}
	_, tmFields, tmcerr := provconfig.GetTmConfigAndMap(cmd)
	if tmcerr != nil {
		return false, fmt.Errorf("couldn't get tendermint config: %v", tmcerr)
	}
	_, clientFields, ccerr := provconfig.GetClientConfigAndMap(cmd)
	if ccerr != nil {
		return false, fmt.Errorf("couldn't get client config: %v", ccerr)
	}

	if len(args) == 0 {
		args = append(args, "all")
	}

	allDefaults := provconfig.GetAllConfigDefaults()
	showApp, showTm, showClient := false, false, false
	appDiffs := map[string]*updatedField{}
	tmDiffs := map[string]*updatedField{}
	clientDiffs := map[string]*updatedField{}
	unknownKeyMap := map[string]reflect.Value{}
	for _, key := range args {
		switch key {
		case "all":
			showApp, showTm, showClient = true, true, true
			for k, v := range getFieldMapChanges(appFields, allDefaults) {
				appDiffs[k] = v
			}
			for k, v := range getFieldMapChanges(tmFields, allDefaults) {
				tmDiffs[k] = v
			}
			for k, v := range getFieldMapChanges(clientFields, allDefaults) {
				clientDiffs[k] = v
			}
		case "app", "cosmos":
			showApp = true
			for k, v := range getFieldMapChanges(appFields, allDefaults) {
				appDiffs[k] = v
			}
		case "config", "tendermint", "tm":
			showTm = true
			for k, v := range getFieldMapChanges(tmFields, allDefaults) {
				tmDiffs[k] = v
			}
		case "client":
			showClient = true
			for k, v := range getFieldMapChanges(clientFields, allDefaults) {
				clientDiffs[k] = v
			}
		default:
			entries, foundIn := findEntries(key, appFields, tmFields, clientFields)
			switch foundIn {
			case 0:
				showApp = true
				for k, v := range entries {
					if uf, ok := makeUpdatedField(k, v, allDefaults); ok {
						appDiffs[k] = &uf
					}
				}
			case 1:
				showTm = true
				for k, v := range entries {
					if uf, ok := makeUpdatedField(k, v, allDefaults); ok {
						tmDiffs[k] = &uf
					}
				}
			case 2:
				showClient = true
				for k, v := range entries {
					if uf, ok := makeUpdatedField(k, v, allDefaults); ok {
						clientDiffs[k] = &uf
					}
				}
			default:
				unknownKeyMap[key] = reflect.Value{}
			}
		}
	}

	if showApp {
		cmd.Println(makeAppConfigHeader(cmd, "Differences from Defaults"))
		if len(appDiffs) > 0 {
			cmd.Println(makeUpdatedFieldMapString(appDiffs, updatedField.StringAsDefault))
		} else {
			cmd.Println("All app config values equal the default config values.")
			cmd.Println("")
		}
	}

	if showTm {
		cmd.Println(makeTmConfigHeader(cmd, "Differences from Defaults"))
		if len(tmDiffs) > 0 {
			cmd.Println(makeUpdatedFieldMapString(tmDiffs, updatedField.StringAsDefault))
		} else {
			cmd.Println("All tendermint config values equal the default config values.")
			cmd.Println("")
		}
	}

	if showClient {
		cmd.Println(makeClientConfigHeader(cmd, "Differences from Defaults"))
		if len(clientDiffs) > 0 {
			cmd.Println(makeUpdatedFieldMapString(clientDiffs, updatedField.StringAsDefault))
		} else {
			cmd.Println("All client config values equal the default config values.")
			cmd.Println("")
		}
	}

	if len(unknownKeyMap) > 0 {
		unknownKeys := getSortedKeys(unknownKeyMap)
		s := "s"
		if len(unknownKeys) == 1 {
			s = ""
		}
		return false, fmt.Errorf("%d configuration key%s not found: %s", len(unknownKeys), s, strings.Join(unknownKeys, ", "))
	}
	return false, nil
}

func runConfigPackCmd(cmd *cobra.Command) error {
	// TODO: Write runConfigPackCmd
	return fmt.Errorf("not implemented")
}

func runConfigUnpackCmd(cmd *cobra.Command) error {
	// TODO: Write runConfigUnpackCmd
	return fmt.Errorf("not implemented")
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
// The second return value is the index of the provided map that the entries were found in (starting with 0).
// If it's equal to entryNotFound, no entries were found.
func findEntries(key string, maps ...map[string]reflect.Value) (map[string]reflect.Value, int) {
	rv := map[string]reflect.Value{}
	baseKey := key
	if len(key) == 0 {
		return rv, entryNotFound
	}
	if key[len(key)-1:] != "." {
		if v, i := findEntry(key, maps...); i != entryNotFound {
			rv[key] = v
			return rv, i
		}
		baseKey += "."
	}
	baseKeyLen := len(baseKey)
	for i, m := range maps {
		for k, v := range m {
			if len(k) > baseKeyLen && k[:baseKeyLen] == baseKey {
				rv[k] = v
			}
		}
		if len(rv) > 0 {
			return rv, i
		}
	}
	return rv, entryNotFound
}

// getFieldMapChanges gets an updated field map with changes between two field maps.
// If the key doesn't exist in both maps, the entry is ignored.
func getFieldMapChanges(isNowMap map[string]reflect.Value, wasMap map[string]reflect.Value) map[string]*updatedField {
	changes := map[string]*updatedField{}
	for key, isNowVal := range isNowMap {
		uf, ok := makeUpdatedField(key, isNowVal, wasMap)
		if ok && uf.HasDiff() {
			changes[key] = &uf
		}
	}
	return changes
}

// makeUpdatedField makes an updatedField with available information.
// The new updatedField will have its key and IsNow set from the provided arguments.
// If the wasMap contains the key, the Was value will be set and the second return argument will be true.
// If the wasMap does not contain the key, the second return argument will be false.
func makeUpdatedField(key string, isNowVal reflect.Value, wasMap map[string]reflect.Value) (updatedField, bool) {
	rv := updatedField{
		Key:   key,
		IsNow: getStringFromValue(isNowVal),
	}
	if wasVal, ok := wasMap[key]; ok {
		rv.Was = getStringFromValue(wasVal)
		return rv, true
	}
	return rv, false
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
			var val []string
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
				var val [][]string
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

// makeFieldMapString makes a multi-line string with all the keys and values in the provided map.
func makeFieldMapString(m map[string]reflect.Value) string {
	keys := getSortedKeys(m)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(getStringFromValue(m[k]))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// makeUpdatedFieldMapString makes a multi-line string of the given updated field map.
// The provided stringer function is used to convert each map value to a string.
func makeUpdatedFieldMapString(m map[string]*updatedField, stringer func(v updatedField) string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	keys = sortKeys(keys)
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(stringer(*m[key]))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// getSortedKeys gets the keys of the provided map and sorts them using sortKeys.
func getSortedKeys(m map[string]reflect.Value) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return sortKeys(keys)
}

// sortKeys sorts the provided keys slice.
// Base keys are put first and sorted alphabetically followed by keys in sub-configs sorted alphabetically.
func sortKeys(keys []string) []string {
	var baseKeys []string
	var subKeys []string
	for _, k := range keys {
		if strings.Contains(k, ".") {
			subKeys = append(subKeys, k)
		} else {
			baseKeys = append(baseKeys, k)
		}
	}
	sort.Strings(baseKeys)
	sort.Strings(subKeys)
	copy(keys, baseKeys)
	for i, k := range subKeys {
		keys[i+len(baseKeys)] = k
	}
	return keys
}

// makeSectionHeaderString creates a string to use as a section header in output.
func makeSectionHeaderString(lead, addedLead, filename string) string {
	var sb strings.Builder
	sb.WriteString(lead)
	if len(addedLead) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(addedLead)
	}
	sb.WriteByte(':')
	hr := strings.Repeat("-", sb.Len())
	if len(filename) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(filename)
		hr += "-----"
	}
	sb.WriteByte('\n')
	sb.WriteString(hr)
	return sb.String()
}

// makeAppConfigHeader creates a section header string for app config stuff.
func makeAppConfigHeader(c *cobra.Command, addedLead string) string {
	return makeSectionHeaderString("App Config", addedLead, provconfig.GetFullPathToAppConf(c))
}

// makeTmConfigHeader creates a section header string for tendermint config stuff.
func makeTmConfigHeader(c *cobra.Command, addedLead string) string {
	return makeSectionHeaderString("Tendermint Config", addedLead, provconfig.GetFullPathToTmConf(c))
}

// makeClientConfigHeader creates a section header string for client config stuff.
func makeClientConfigHeader(c *cobra.Command, addedLead string) string {
	return makeSectionHeaderString("Client Config", addedLead, provconfig.GetFullPathToClientConf(c))
}
