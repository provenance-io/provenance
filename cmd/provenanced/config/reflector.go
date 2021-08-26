package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FieldValueMap maps field names to reflect.Value objects.
type FieldValueMap map[string]reflect.Value

// MakeFieldValueMap gets a map of string field names to reflect.Value objects for a given object.
// An empty map is returned if the provided obj is nil, or isn't either a struct or pointer chain to a struct.
// Substruct fields will have the parent struct's field name and substruct field name separated by a period.
// E.g. This struct { Foo { Bar: "foobar" } }, will have an entry for "foo.bar".
// Each segment of a name comes from getFieldName. I.e. pays attention to mapstruct and is all lowercase.
// Similarly, if mapstruct says to squash the fields, they won't have the parent field name.
// If fillNilsWithZero is true, nil fields in the obj will be filled in using zero values for that field type.
//    Fields in structs created this way cannot be set later using the map values.
//    But the resulting map will contain information about all possible fields in the object.
//    None of the substructures will not have map entries specifically for the parent field containing the substructure.
//    E.g. With { Foo: { Bar: "foobar" } }, there won't be an entry for "foo", but there will be one for "foo.bar".
// If fillNilsWithZero is false, nil fields in the obj will have an entry for the field and a value where .IsNil() is true.
//    If the provided obj is a pointer, then the resulting map values can be used to set values in the obj.
//    Substructures that are nil will have entries in the resulting map.
//    E.g. With { Foo: (*Bar)(nil) }, there will be an entry for "foo".
//    Substructures that are not nill will still not have an entry though; they'll have entries for sub-fields still.
//    E.g. With { Foo: (*Bar)(nil), Ban: { Ana: "banana" } will have entries for "foo" and "ban.ana", but not "ban" (it isn't nil).
func MakeFieldValueMap(obj interface{}, fillNilsWithZero bool) FieldValueMap {
	if obj == nil {
		return FieldValueMap{}
	}
	return getFieldValueMap(reflect.ValueOf(obj), fillNilsWithZero)
}

// getFieldValueMap does all the heavy lifting for MakeFieldValueMap.
// Most of the time, you'll want to use MakeFieldValueMap instead of this.
// This operates using reflect.Value objects instead of interface{}.
func getFieldValueMap(valIn reflect.Value, fillNilsWithZero bool) FieldValueMap {
	keys := FieldValueMap{}
	objVal := deref(valIn, fillNilsWithZero)
	objType := objVal.Type()
	objBaseKind := objType.Kind()
	if objBaseKind != reflect.Struct {
		return keys
	}
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if len(field.PkgPath) != 0 {
			// PkgPath is the package path that qualifies a lower case (unexported)
			// field name. It is empty for upper case (exported) field names.
			// See https://golang.org/ref/spec#Uniqueness_of_identifiers
			continue
		}
		key, squash := getFieldName(field)
		fVal := deref(objVal.Field(i), fillNilsWithZero)
		fKind := fVal.Kind()
		switch {
		case fKind == reflect.Ptr && fVal.IsNil():
			keys[key] = fVal
		case squash:
			for k, v := range getFieldValueMap(fVal, fillNilsWithZero) {
				keys[k] = v
			}
		case fKind == reflect.Struct:
			for subkey, subval := range getFieldValueMap(fVal, fillNilsWithZero) {
				keys[key+"."+subkey] = subval
			}
		default:
			keys[key] = fVal
		}
	}
	return keys
}

// getFieldName gets the field name and whether or not it needs a squashing.
// The name returned will always be all lowercase.
// If there is a mapstructure tag, and the name in there isn't "" or "-", then that name is used.
// Otherwise, field.Name is used.
// A field needs squashing only if requested in the mapstruct tag for the field.
func getFieldName(field reflect.StructField) (string, bool) {
	name := field.Name
	squash := false
	if tag, ok := field.Tag.Lookup("mapstructure"); ok {
		if index := strings.Index(tag, ","); index != -1 {
			if len(tag[:index]) > 0 {
				name = tag[:index]
			}
			squash = strings.Contains(tag[index+1:], "squash")
		} else {
			name = tag
		}
		if name == "-" {
			name = field.Name
		}
	}
	return strings.ToLower(name), squash
}

// deref follows any pointers until it gets to something else.
// If not a pointer, the provided v is returned.
// If it's a pointer (or pointer chain) to something other than nil, that something is returned.
// If fillNilsWithZero is true, and the pointer chain ends at a nil, a zero value of the pointers type is created and returned.
// If fillNilsWithZero is false, and the pointer chain ends at a nil, the originally provided v is returned.
func deref(v reflect.Value, fillNilsWithZero bool) reflect.Value {
	vOrig := v
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !fillNilsWithZero {
				return vOrig
			}
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = v.Elem()
		}
	}
	return v
}

// Has checks if the provided key exists in this FieldValueMap.
func (m FieldValueMap) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// GetSortedKeys gets the keys of this FieldValueMap and sorts them using sortKeys.
func (m FieldValueMap) GetSortedKeys() []string {
	rv := make([]string, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	return sortKeys(rv)
}

// FindEntries looks for entries in this map that match the provided key.
// First, if the key doesn't end in a period, an exact entry match is looked for.
// If such an entry is found, only that entry will be returned.
// If no such entry is found, a period is added to the end of the key (if not already there).
// E.g. Providing "filter_peers" will get just the "filter_peers" entry.
// Providing "consensus." will bypass the exact key lookup, and return all fields that start with "consensus.".
// Providing "consensus" will look first for a field specifically called "consensus",
// then, if/when not found, will return all fields that start with "consensus.".
// Then, all entries with keys that start with the desired key are returned.
// The second return value indicates whether or not anything was found.
func (m FieldValueMap) FindEntries(key string) (FieldValueMap, bool) {
	rv := FieldValueMap{}
	if len(key) == 0 {
		return rv, false
	}
	if key[len(key)-1:] != "." {
		if val, ok := m[key]; ok {
			rv[key] = val
			return rv, true
		}
		key += "."
	}
	keylen := len(key)
	for k, v := range m {
		if len(k) > keylen && k[:keylen] == key {
			rv[k] = v
		}
	}
	return rv, len(rv) > 0
}

// GetStringOf gets a string representation of the value with the given key.
// If the key doesn't exist in this FieldValueMap, an empty string is returned.
func (m FieldValueMap) GetStringOf(key string) string {
	if v, ok := m[key]; ok {
		return GetStringFromValue(v)
	}
	return ""
}

// GetStringFromValue gets a string of the given value.
// This creates strings that are more in line with what the values look like in the config files.
// For slices and arrays, it turns into `["a", "b", "c"]`.
// For strings, it turns into `"a"`.
// For anything else, it just uses fmt %v.
// This wasn't designed with the following kinds in mind:
//    Invalid, Chan, Func, Interface, Map, Ptr, Struct, or UnsafePointer.
func GetStringFromValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		var sb strings.Builder
		sb.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(GetStringFromValue(v.Index(i)))
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

// SetFromString sets a value from the provided string.
// The string is converted appropriately for the underlying value type.
// Assuming the value came from MakeFieldValueMap, this will actually be updating the
// value in the config object provided to that function.
func (m FieldValueMap) SetFromString(key, valueStr string) error {
	if v, ok := m[key]; ok {
		return setValueFromString(key, v, valueStr)
	}
	return fmt.Errorf("no field found for key: %s", key)
}

// setValueFromString sets a value from the provided string.
// The string is converted appropriately for the underlying value type.
// Assuming the value came from MakeFieldValueMap, this will actually be updating the
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
							return fmt.Errorf("invalid %s: sub-arrays must have length 2, but the sub-array at index %d has length %d", fieldName, i, len(s))
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

// SetToNil sets the given key to a nil value.
func (m FieldValueMap) SetToNil(key string) {
	m[key] = reflect.Value{}
}

// AddEntriesFrom Add all entries from the provided map into this map.
// If the same key exists in both maps, the value from the one provided will overwrite the one in this map.
func (m FieldValueMap) AddEntriesFrom(maps ...FieldValueMap) {
	for _, m2 := range maps {
		for k, v := range m2 {
			m[k] = v
		}
	}
}

// UpdatedField is a struct holding information about a config field that has been updated.
type UpdatedField struct {
	Key   string
	Was   string
	IsNow string
}

// MakeUpdatedField creates an UpdateField with the given key getting the values from each provided map.
// If either one of the maps doesn't have the key, the second returned value will be false.
func MakeUpdatedField(key string, wasMap, isNowMap FieldValueMap) (UpdatedField, bool) {
	rv := UpdatedField{
		Key: key,
	}
	wasVal, wasFound := wasMap[key]
	if wasFound {
		rv.Was = GetStringFromValue(wasVal)
	}
	isNowVal, isNowFound := isNowMap[key]
	if isNowFound {
		rv.IsNow = GetStringFromValue(isNowVal)
	}
	return rv, isNowFound && wasFound
}

// Update updates the base UpdatedField given information in the provided newerInfo.
func (u *UpdatedField) Update(newerInfo UpdatedField) {
	u.IsNow = newerInfo.IsNow
}

// String converts an UpdatedField to a string similar to using %#v but a little cleaner.
func (u UpdatedField) String() string {
	return fmt.Sprintf(`UpdatedField{Key:%s, Was:%s, IsNow:%s}`, u.Key, u.Was, u.IsNow)
}

// StringAsUpdate creates a string from this UpdatedField indicating a change has being made.
func (u UpdatedField) StringAsUpdate() string {
	return fmt.Sprintf("%s Was: %s, Is Now: %s", u.Key, u.Was, u.IsNow)
}

// StringAsDefault creates a string from this UpdatedField identifying the Was as a default.
func (u UpdatedField) StringAsDefault() string {
	if !u.HasDiff() {
		return fmt.Sprintf("%s=%s (same as default)", u.Key, u.IsNow)
	}
	return fmt.Sprintf("%s=%s (default=%s)", u.Key, u.IsNow, u.Was)
}

// HasDiff returns true if IsNow and Was have different values.
func (u UpdatedField) HasDiff() bool {
	return u.IsNow != u.Was
}

// UpdatedFieldMap maps field names to UpdatedField references.
type UpdatedFieldMap map[string]*UpdatedField

// MakeUpdatedFieldMap creates an UpdatedFieldMap with fields that exist in both provided FieldValueMap objects.
// Set onlyChanged to true to further limit fields to only those with differences.
func MakeUpdatedFieldMap(wasMap, isNowMap FieldValueMap, onlyChanged bool) UpdatedFieldMap {
	rv := UpdatedFieldMap{}
	for key := range isNowMap {
		uf, ok := MakeUpdatedField(key, wasMap, isNowMap)
		if ok && (!onlyChanged || uf.HasDiff()) {
			rv[key] = &uf
		}
	}
	return rv
}

// AddOrUpdate adds or updates an entry in this map using the provided info.
// If the key isn't in this map yet, the entry is added.
// Otherwise UpdatedField.Update us called on the existing entry.
func (m UpdatedFieldMap) AddOrUpdate(key, was, isNow string) {
	info := UpdatedField{
		Key:   key,
		Was:   was,
		IsNow: isNow,
	}
	m.AddOrUpdateEntry(&info)
}

// AddOrUpdateEntry adds or updates an entry in this map using the provided info.
// If the key isn't in this map yet, the entry is added.
// Otherwise UpdatedField.Update us called on the existing entry.
func (m UpdatedFieldMap) AddOrUpdateEntry(info *UpdatedField) {
	if uf, ok := m[info.Key]; ok {
		uf.Update(*info)
	} else {
		m[info.Key] = info
	}
}

// AddOrUpdateEntriesFrom applies AddOrUpdateEntry on this map using each entry in the provided map.
func (m UpdatedFieldMap) AddOrUpdateEntriesFrom(maps ...UpdatedFieldMap) {
	for _, m2 := range maps {
		for _, info := range m2 {
			m.AddOrUpdateEntry(info)
		}
	}
}

// GetSortedKeys gets the keys of this UpdatedFieldMap and sorts them using sortKeys.
func (m UpdatedFieldMap) GetSortedKeys() []string {
	rv := make([]string, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	return sortKeys(rv)
}

// sortKeys sorts the provided keys slice.
// Base keys are put first and sorted alphabetically
// followed by keys in sub-configs sorted alphabetically.
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
