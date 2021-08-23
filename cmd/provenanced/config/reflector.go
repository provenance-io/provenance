package config

import (
	"reflect"
	"strings"
)

// FieldValueMap maps field names to reflect.Value objects.
type FieldValueMap map[string]reflect.Value

// GetFieldValueMap gets a map of string field names to reflect.Value objects for a given object.
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
func GetFieldValueMap(obj interface{}, fillNilsWithZero bool) FieldValueMap {
	if obj == nil {
		return FieldValueMap{}
	}
	return getFieldValueMapValue(reflect.ValueOf(obj), fillNilsWithZero)
}

// getFieldValueMapValue does all the heavy lifting for getFieldValueMap.
// Most of the time, you'll want to use getFieldValueMap instead of this.
// This operates using reflect.Value objects instead of interface{}.
func getFieldValueMapValue(valIn reflect.Value, fillNilsWithZero bool) FieldValueMap {
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
			for k, v := range getFieldValueMapValue(fVal, fillNilsWithZero) {
				keys[k] = v
			}
		case fKind == reflect.Struct:
			for subkey, subval := range getFieldValueMapValue(fVal, fillNilsWithZero) {
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
// If there is a mapstructure tag, an the name in there isn't "" or "-", then that name is used.
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
