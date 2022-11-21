package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ReflectorTestSuit struct {
	suite.Suite
}

func (s *ReflectorTestSuit) SetupTest() {

}

func TestReflectorTestSuit(t *testing.T) {
	suite.Run(t, new(ReflectorTestSuit))
}

type SquashedThing struct {
	SquashedExportedField   string `mapstructure:"exported-field"`
	squashedUnexportedField string
	SquashedSubThing1       SubThing1 `mapstructure:"squashed-sub-thing-1"`
}

type SubThing1 struct {
	AnInt int
	AUint uint
}

type SubThing2 struct {
	AString     string   `mapstructure:"a-string"`
	SomeStrings []string `mapstructure:"some-strings"`
}

type MainThing struct {
	SquashedThing `mapstructure:",squash"`
	SThing1       SubThing1 `mapstructure:"main-sub-thing"`
	PSThing2      *SubThing2
	DeepThing     ******SubThing2
	MainInt       int `mapstructure:"main-int"`
}

const (
	squashedExportedField   = "Now you see me."
	squashedUnexportedField = "Now you don't."
	squashedSubThing1AnInt  = -10
	squashedSubThing1AUint  = uint(10)
	mainSubThingAnInt       = -5
	mainSubThingAUint       = uint(5)
	mainInt                 = 2
)

func DefaultMainThing() MainThing {
	return MainThing{
		SquashedThing: SquashedThing{
			SquashedExportedField:   squashedExportedField,
			squashedUnexportedField: squashedUnexportedField,
			SquashedSubThing1: SubThing1{
				AnInt: squashedSubThing1AnInt,
				AUint: squashedSubThing1AUint,
			},
		},
		SThing1: SubThing1{
			AnInt: mainSubThingAnInt,
			AUint: mainSubThingAUint,
		},
		PSThing2:  nil,
		DeepThing: nil,
		MainInt:   mainInt,
	}
}

func noPanic(f func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		assert.NotPanics(t, func() {
			f(t)
		})
	}
}

func (s *ReflectorTestSuit) TestGetFieldValueMapWithFill() {
	thing := DefaultMainThing()
	thingMap := MakeFieldValueMap(&thing, true)

	s.T().Run("main-sub-thing does not exist", noPanic(func(t *testing.T) {
		// main-sub-thing should not exist because it's a struct, only it's sub-fields should have entries.
		key := "main-sub-thing"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("psthing2 does not exist", noPanic(func(t *testing.T) {
		// psthing2 should not exist because it's a pointer to a struct.
		// Since fillNilsWithZero was true, only its sub-fields should have entries.
		key := "psthing2"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("psthing2.a-string exists", noPanic(func(t *testing.T) {
		key := "psthing2.a-string"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.False(t, value.CanAddr(), "%s value.CanAddr()", key)
			assert.False(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("psthing2.some-strings exists", noPanic(func(t *testing.T) {
		key := "psthing2.some-strings"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.False(t, value.CanAddr(), "%s value.CanAddr()", key)
			assert.False(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("deepthing does not exist", noPanic(func(t *testing.T) {
		// deepthing should not exist because it's a pointer to a struct.
		// Since fillNilsWithZero was true, only its sub-fields should have entries.
		key := "deepthing"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("deepthing.a-string exists", noPanic(func(t *testing.T) {
		key := "deepthing.a-string"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.False(t, value.CanAddr(), "%s value.CanAddr()", key)
			assert.False(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("deepthing.some-strings exists", noPanic(func(t *testing.T) {
		key := "deepthing.some-strings"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.False(t, value.CanAddr(), "%s value.CanAddr()", key)
			assert.False(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("main-int exists", noPanic(func(t *testing.T) {
		key := "main-int"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actual := int(value.Int())
			assert.Equal(t, mainInt, actual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))
}

func (s *ReflectorTestSuit) TestGetFieldValueMapNoFillWithValue() {
	aString := "This is a string!"
	someStrings := []string{"one", "two", "three"}
	thing := DefaultMainThing()
	thing.PSThing2 = &SubThing2{
		AString:     aString,
		SomeStrings: someStrings,
	}
	thingMap := MakeFieldValueMap(&thing, true)

	s.T().Run("psthing2.a-string exists", noPanic(func(t *testing.T) {
		key := "psthing2.a-string"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actual := value.String()
			assert.Equal(t, aString, actual, "%s value", key)
			require.True(t, value.CanSet(), "%s value.CanSet()", key)
			fromObjExpected := "New string value"
			value.SetString(fromObjExpected)
			fromObjActual := thing.PSThing2.AString
			assert.Equal(t, fromObjExpected, fromObjActual, "%s from obj after set")
		}
	}))

	s.T().Run("psthing2.some-strings exists", noPanic(func(t *testing.T) {
		key := "psthing2.some-strings"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actualLen := value.Len()
			actual := make([]string, actualLen)
			for i := 0; i < actualLen; i++ {
				actual[i] = value.Index(i).String()
			}
			assert.Equal(t, someStrings, actual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))
}

func (s *ReflectorTestSuit) TestGetFieldValueMapBaseFieldsNoFill() {
	thing := DefaultMainThing()
	thingMap := MakeFieldValueMap(&thing, false)

	s.T().Run("main-sub-thing does not exist", noPanic(func(t *testing.T) {
		// main-sub-thing shouldn't exist because it's a struct, only it's sub-fields should have entries.
		key := "main-sub-thing"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("psthing2 exists", noPanic(func(t *testing.T) {
		// psthing2 should exist because it's a Ptr, it's value is nil, and fillNilsWithZero was false.
		key := "psthing2"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.Equal(t, value.Kind(), reflect.Ptr, "%s kind", key)
			assert.True(t, value.IsNil(), "%s value.IsNil()", key)
			assert.Equal(t, "*config.SubThing2", value.Type().String(), "%s type", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("deepthing exists", noPanic(func(t *testing.T) {
		// deepthing should exist because it's a Ptr, it's value is nil, and fillNilsWithZero was false.
		key := "deepthing"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			assert.Equal(t, value.Kind(), reflect.Ptr, "%s kind", key)
			assert.True(t, value.IsNil(), "%s value.IsNil()", key)
			assert.Equal(t, "******config.SubThing2", value.Type().String(), "%s type", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("main-int exists", noPanic(func(t *testing.T) {
		key := "main-int"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actual := int(value.Int())
			assert.Equal(t, mainInt, actual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))
}

func (s *ReflectorTestSuit) TestGetFieldValueMapSquashedFields() {
	thing := DefaultMainThing()
	thingMap := MakeFieldValueMap(&thing, false)

	s.T().Run("squashedthing does not exist", noPanic(func(t *testing.T) {
		// squashedthing shouldn't exist because it's squashed (fields look like main fields)
		key := "squashedthing"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("exported-field exists", noPanic(func(t *testing.T) {
		key := "exported-field"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			expFieldActual := value.String()
			assert.Equal(t, squashedExportedField, expFieldActual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("squashedunexportedfield does not exist", noPanic(func(t *testing.T) {
		key := "squashedunexportedfield"
		value, ok := thingMap[key]
		if assert.False(t, ok, "%s found", key) {
			assert.False(t, value.IsValid(), "%s value.IsValid()", key)
		}
	}))

	s.T().Run("squashed-sub-thing-1.anint exists", noPanic(func(t *testing.T) {
		key := "squashed-sub-thing-1.anint"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			sst1AnIntActual := int(value.Int())
			assert.Equal(t, squashedSubThing1AnInt, sst1AnIntActual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("squashed-sub-thing-1.auint exists", noPanic(func(t *testing.T) {
		key := "squashed-sub-thing-1.auint"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			sst1AUintActual := uint(value.Uint())
			assert.Equal(t, squashedSubThing1AUint, sst1AUintActual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))
}

func (s *ReflectorTestSuit) TestGetFieldValueMapSubFields() {
	thing := DefaultMainThing()
	thingMap := MakeFieldValueMap(&thing, false)

	s.T().Run("main-sub-thing.anint exists", noPanic(func(t *testing.T) {
		key := "main-sub-thing.anint"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actual := int(value.Int())
			assert.Equal(t, mainSubThingAnInt, actual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))

	s.T().Run("main-sub-thing.auint exists", noPanic(func(t *testing.T) {
		key := "main-sub-thing.auint"
		value, ok := thingMap[key]
		if assert.True(t, ok, "%s found", key) {
			actual := uint(value.Uint())
			assert.Equal(t, mainSubThingAUint, actual, "%s value", key)
			assert.True(t, value.CanSet(), "%s value.CanSet()", key)
		}
	}))
}

func (s *ReflectorTestSuit) TestAsConfigMap() {
	// Note: Because the AsConfigMap function loops over a map,
	// for the ones expecting errors, there's no telling which error it will hit.
	// I.e. we don't know if it ends up adding the section first or the key first.
	// So we can only make sure that the error is one of a possible set.
	tests := []struct {
		name     string
		fvm      FieldValueMap
		exp      map[string]interface{}
		errOneOf []string
	}{
		{
			name: "empty",
			fvm:  FieldValueMap{},
			exp:  make(map[string]interface{}),
		},
		{
			name: "two keys",
			fvm: FieldValueMap{
				"thing1": reflect.ValueOf("value of thing 1"),
				"thing2": reflect.ValueOf(99),
			},
			exp: map[string]interface{}{
				"thing1": "value of thing 1",
				"thing2": 99,
			},
		},
		{
			name: "some deep things",
			fvm: FieldValueMap{
				"rootthing":        reflect.ValueOf("root thing"),
				"sub1.sub2.thing1": reflect.ValueOf("I am thing 1"),
				"sub1.nothing":     reflect.ValueOf("do I exist?"),
				"sub1.sub2.thing2": reflect.ValueOf(2),
				"sub1.sub3.thing3": reflect.ValueOf([]string{"a slice!"}),
			},
			exp: map[string]interface{}{
				"rootthing": "root thing",
				"sub1": map[string]interface{}{
					"nothing": "do I exist?",
					"sub2": map[string]interface{}{
						"thing1": "I am thing 1",
						"thing2": 2,
					},
					"sub3": map[string]interface{}{
						"thing3": []string{"a slice!"},
					},
				},
			},
		},
		{
			name: "reused section name",
			fvm: FieldValueMap{
				"mysect":       reflect.ValueOf("oops"),
				"mysect.value": reflect.ValueOf("a deeper value"),
			},
			errOneOf: []string{
				`error at key "mysect": should not already exist but has type map[string]interface {}`,
				`error at key "mysect.value" at section "mysect": sub-section should have type map[string]interface{}, got string`,
			},
		},
		{
			name: "deep reused section name",
			fvm: FieldValueMap{
				"rootthing":        reflect.ValueOf("a root thing"),
				"sub1.sub2.value1": reflect.ValueOf("value 1"),
				"sub1.sub2":        reflect.ValueOf("sub 2 oops"),
			},
			errOneOf: []string{
				`error at key "sub1.sub2.value1" at section "sub2": sub-section should have type map[string]interface{}, got string`,
				`error at key "sub1.sub2": key "sub2" should not already exist in sub-section but has type map[string]interface {}`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual, err := tc.fvm.AsConfigMap()
			if len(tc.errOneOf) > 0 {
				if s.Assert().Error(err, "AsConfigMap error: actual = %+v", actual) {
					s.Assert().Contains(tc.errOneOf, err.Error(), "AsConfigMap error")
				}
			} else {
				if s.Assert().NoError(err, "AsConfigMap error") {
					s.Assert().Equal(tc.exp, actual, "AsConfigMap result")
				}
			}
		})
	}
}

func (s *ReflectorTestSuit) TestAsConfigMapLotsMoreTimes() {
	// Because of the non-deterministic nature of map loops, there's some
	// uncertainty in what happens in AsConfigMap (i.e. which error gets triggered).
	// To be more certain that it's all good, run it 1,000 more times.
	// Hopefully in that time, the map looping order changes and things happen
	// differently, but the tests still pass.

	for i := 1; i <= 1000; i++ {
		s.Run(fmt.Sprintf("%04d", i), s.TestAsConfigMap)
		if s.T().Failed() {
			break
		}
	}
}
