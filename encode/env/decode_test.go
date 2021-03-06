package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecoder_Unmarshal(t *testing.T) {
	type Level3 struct {
		FirstField *string
		SecondField string `env:"second_field"`
	}

	type level2 struct {
		FirstField *string
		SecondField string `env:"second_field"`

		privateField string // Should not be populated.

		Level3 Level3 `env:"LEVEL3"` // Should not matter if struct type is public or private. Only the field name.
		Level3Pointer *Level3 // Should initialize Level3 and store the pointer type and not panic.
	}

	type aField struct {
		Field1 string
	}

	type withprefix struct {
		WithPrefix aField
	}

	type omitprefix struct {
		NoPrefix aField `env:"omitprefix"`
	}

	type Embedded struct {
		EmbeddedField string
	}

	type EmbeddedPointer struct {
		EmbeddedPointerField string
	}

	type EmbeddedWPrefix struct {
		EmbeddedWPrefixField string
	}

	type EmbeddedCustomPrefix struct {
		EmbeddedCustomField string
	}

	type SliceStruct struct {
		SliceField string
	}

	type CustomIntType int

	type level1 struct {
		// note: private embedded structs are not accessible.
		Embedded `env:"omitprefix"` // if omitprefix not provided then the prefix is "Embedded"
		*EmbeddedPointer `env:"omitprefix"` // pointer also is valid
		EmbeddedWPrefix // keep prefix
		EmbeddedCustomPrefix `env:"E"`

		DurField time.Duration
		TimeField time.Time // default format is RFC3339
		TimeCustomField time.Time `fmt:"2006/01/02"` // custom format
		TimePointerField *time.Time // *time.Time supported but be careful with referencing time.Time!
		FirstField *string `env:"first_field"`
		SecondField string
		IntField int
		CustomIntField CustomIntType // just treated as an int
		IntPointerField *int
		BoolField bool
		BoolFieldFalse bool // default is true but set env is false so final value should be false.
		BoolPointerField *bool
		ArrayField [3]int
		SliceStringField []string
		SliceIntField []int
		SliceIntFieldWSpaces []int // input env value should be able to be '1, 2, 3'
		SliceIntFieldWQuotes1 []int // input env value should be able to be '"1","2","3"'
		SliceIntFieldWQuotes2 []int // input env value should be able to be "'1','2','3''
		SliceIntFieldSquareBrackets []int // input env value should be able to be "[1,2,3]"
		SliceFloatField []float32
		SliceStructField []SliceStruct // slice of structs is ignored
		MapField map[string]string // maps are ignored.
		IgnoreField string `env:"-"` // ignore field
		IgnoreStruct level2 `env:"-"` // ignore struct
		IgnorePointerStruct *level2 `env:"-"` // ignore struct pointer (will not even be initialized)

		// omitprefix
		// this level omits prefix but the next one does not.
		OmitPrefix withprefix `env:"omitprefix"`
		OmitPrefixPointer *withprefix `env:"omitprefix"`

		// omitprefix fallthrough
		// prefix at this level but next level prefix omitted.
		WithPrefixInherited omitprefix
		WithPrefixInheritedPointer *omitprefix

		Level2 level2 `env:"LEVEL2"`

		privateField string // should not be populated
		privateFieldWTag string `env:"private_field_w_tag"` // Will not populate private field even with tag.
	}

	cfg := level1{
		BoolFieldFalse: true,
	}

	// error if struct is not a pointer
	d := &Decoder{}
	err := d.Unmarshal(cfg)
	assert.NotNil(t, err)

	// set env vars
	os.Setenv("EMBEDDED_FIELD", "vEMBEDDED_FIELD")
	os.Setenv("EMBEDDED_POINTER_FIELD", "vEMBEDDED_POINTER_FIELD")
	os.Setenv("EMBEDDED_W_PREFIX_EMBEDDED_W_PREFIX_FIELD", "vEMBEDDED_W_PREFIX_EMBEDDED_W_PREFIX_FIELD")
	os.Setenv("E_EMBEDDED_CUSTOM_FIELD", "vE_EMBEDDED_CUSTOM_FIELD")
	os.Setenv("DUR_FIELD", "64s")
	os.Setenv("TIME_FIELD", "2020-01-01T11:04:01Z") // RFC3339
	os.Setenv("TIME_CUSTOM_FIELD", "2020/01/02") // custom format
	os.Setenv("TIME_POINTER_FIELD", "2020-01-01T11:04:01Z") // RFC3339
	os.Setenv("first_field", "vfirst_field")
	os.Setenv("SECOND_FIELD", "vSECOND_FIELD")
	os.Setenv("INT_FIELD", "1")
	os.Setenv("CUSTOM_INT_FIELD", "3")
	os.Setenv("INT_POINTER_FIELD", "2")
	os.Setenv("BOOL_FIELD", "true")
	os.Setenv("BOOL_FIELD_FALSE", "false") // false should overwrite default of true
	os.Setenv("BOOL_POINTER_FIELD", "true")
	os.Setenv("ARRAY_FIELD", "1,2,3")
	os.Setenv("SLICE_STRING_FIELD", "part1,part2")
	os.Setenv("SLICE_INT_FIELD", "1,2,3")
	os.Setenv("SLICE_INT_FIELD_W_SPACES", "1, 2, 3")
	os.Setenv("SLICE_INT_FIELD_W_QUOTES_1", `"1","2","3"`)
	os.Setenv("SLICE_INT_FIELD_W_QUOTES_2", `'1','2','3'`)
	os.Setenv("SLICE_INT_FIELD_SQUARE_BRACKETS", "[1,2,3]")
	os.Setenv("SLICE_FLOAT_FIELD", "1.1,2.2,3.3")
	os.Setenv("SLICE_STRUCT_FIELD", "ignored") // structs as single values are ignored.
	os.Setenv("IGNORE_FIELD", "vIGNORE_FIELD") // should not get populated
	os.Setenv("-", "vIGNORE_FIELD") // make sure it doesn't look for a '-' env variable.
	os.Setenv("IGNORE_STRUCT", "vIGNORE_STRUCT") // should not get populated
	os.Setenv("IGNORE_POINTER_STRUCT", "vIGNORE_POINTER_STRUCT") // should not get populated
	os.Setenv("WITH_PREFIX_FIELD_1", "vWITH_PREFIX_FIELD_1") // field should have this name (top level prefix omitted but next level retained).
	os.Setenv("WITH_PREFIX_INHERITED_FIELD_1", "vWITH_PREFIX_INHERITED_FIELD_1") // top level has prefix but next level ignores it.
	os.Setenv("WITH_PREFIX_INHERITED_POINTER_FIELD_1", "vWITH_PREFIX_INHERITED_POINTER_FIELD_1")
	os.Setenv("PRIVATE_FIELD", "vPRIVATE_FIELD") // should not get set
	os.Setenv("private_field_w_tag", "vprivate_field_w_tag") // should not get set
	os.Setenv("PRIVATE_FIELD_W_TAG", "vPRIVATE_FIELD_W_TAG") // just checking this variation in case a logic slip.
	os.Setenv("LEVEL2_FIRST_FIELD", "vLEVEL2_FIRST_FIELD")
	os.Setenv("LEVEL2_second_field", "vLEVEL2_second_field")
	os.Setenv("LEVEL2_PRIVATE_FIELD", "vLEVEL2_PRIVATE_FIELD") // should not get set
	os.Setenv("LEVEL2_LEVEL3_FIRST_FIELD", "vLEVEL2_LEVEL3_FIRST_FIELD")
	os.Setenv("LEVEL2_LEVEL3_second_field", "vLEVEL2_LEVEL3_second_field")

	d = &Decoder{}
	cfg = level1{}
	err = d.Unmarshal(&cfg)
	assert.Nil(t, err)

	// make sure each field is populated as expected.
	drtn, _ := time.ParseDuration("64s")
	dte := time.Date(2020, 01, 01, 11, 04, 01, 0,time.UTC)
	assert.Equal(t, cfg.EmbeddedField, "vEMBEDDED_FIELD")
	assert.Equal(t, cfg.EmbeddedPointerField, "vEMBEDDED_POINTER_FIELD")
	assert.Equal(t, cfg.EmbeddedWPrefixField, "vEMBEDDED_W_PREFIX_EMBEDDED_W_PREFIX_FIELD")
	assert.Equal(t, cfg.EmbeddedCustomField, "vE_EMBEDDED_CUSTOM_FIELD")
	assert.Equal(t, cfg.DurField, drtn)
	assert.Equal(t, cfg.TimeField, dte)
	assert.Equal(t, cfg.TimeCustomField, time.Date(2020,01,02,0,0,0,0,time.UTC))
	assert.Equal(t, *cfg.TimePointerField, dte)
	assert.Equal(t, *cfg.FirstField, "vfirst_field")
	assert.Equal(t, cfg.SecondField, "vSECOND_FIELD")
	assert.Equal(t, cfg.IntField, 1)
	assert.Equal(t, int(cfg.CustomIntField), 3) // custom int type just treated as an int.
	assert.Equal(t, *cfg.IntPointerField, 2)
	assert.Equal(t, cfg.BoolField, true)
	assert.Equal(t, cfg.BoolFieldFalse, false)
	assert.Equal(t, *cfg.BoolPointerField, true)
	//assert.Equal(t, cfg.ArrayField, [3]int{1,2,3})
	assert.Equal(t, cfg.SliceStringField, []string{"part1", "part2"})
	assert.Equal(t, cfg.SliceIntField, []int{1,2,3})
	assert.Equal(t, cfg.SliceIntFieldWSpaces, []int{1,2,3})
	assert.Equal(t, cfg.SliceIntFieldWQuotes1, []int{1,2,3})
	assert.Equal(t, cfg.SliceIntFieldSquareBrackets, []int{1,2,3})
	assert.Equal(t, cfg.SliceFloatField, []float32{1.1,2.2,3.3})
	assert.Empty(t, cfg.IgnoreField)
	assert.Empty(t, cfg.IgnoreStruct)
	assert.Empty(t, cfg.IgnorePointerStruct)
	assert.Equal(t, cfg.OmitPrefix.WithPrefix.Field1, "vWITH_PREFIX_FIELD_1")
	assert.Equal(t, cfg.OmitPrefixPointer.WithPrefix.Field1, "vWITH_PREFIX_FIELD_1")
	assert.Equal(t, cfg.WithPrefixInherited.NoPrefix.Field1, "vWITH_PREFIX_INHERITED_FIELD_1")
	assert.Equal(t, cfg.WithPrefixInheritedPointer.NoPrefix.Field1, "vWITH_PREFIX_INHERITED_POINTER_FIELD_1")
	assert.Empty(t, cfg.privateField)
	assert.Empty(t, cfg.privateFieldWTag)
	assert.Equal(t, *cfg.Level2.FirstField, "vLEVEL2_FIRST_FIELD")
	assert.Equal(t, cfg.Level2.SecondField, "vLEVEL2_second_field")
	assert.Empty(t, cfg.Level2.privateField)
	assert.Equal(t, *cfg.Level2.Level3.FirstField, "vLEVEL2_LEVEL3_FIRST_FIELD")
	assert.Equal(t, cfg.Level2.Level3.SecondField, "vLEVEL2_LEVEL3_second_field")

	// misc tests
	// Test: 'omitprefix' on non-struct and pointer non-struct
	type omitprefixNonStruct struct {
		OmitPrefixField string `env:"omitprefix"` // not allowed returns error.
	}

	d = &Decoder{}
	cfgErr := omitprefixNonStruct{}
	err = d.Unmarshal(&cfgErr)
	assert.EqualError(t, err, "'omitprefix' cannot be used on non-struct field types")

	// Test: a comma in the env tag value gets translated directly as an env field
	// same as everything else. While it doesn't return an error the user is unlikely
	// to set an env variable with a comma. Regardless, the behavior is defined.
	type envComma struct {
		CommaField string `env:"commafield,"`
	}

	os.Setenv("commafield,", "vcommafield,")

	d = &Decoder{}
	cfgComma := envComma{}
	err = d.Unmarshal(&cfgComma)
	assert.Nil(t, err)
	assert.Equal(t, "vcommafield,", cfgComma.CommaField)

	// Test: incorrect formatting - tag value is omitted. only 'env' is provided.
	type envNoValue struct {
		NoTagValueField string `env:""` // does not return error but has no effect.
		NoTagValueField2 string `env` // not even the ':""' provided.
	}

	os.Setenv("NO_TAG_VALUE_FIELD", "vNO_TAG_VALUE_FIELD")
	os.Setenv("NO_TAG_VALUE_FIELD_2", "vNO_TAG_VALUE_FIELD_2")

	d = &Decoder{}
	cfgEnvNoValue := envNoValue{}
	err = d.Unmarshal(&cfgEnvNoValue)
	assert.Nil(t, err)
	assert.Equal(t, "vNO_TAG_VALUE_FIELD", cfgEnvNoValue.NoTagValueField)
	assert.Equal(t, "vNO_TAG_VALUE_FIELD_2", cfgEnvNoValue.NoTagValueField2)

	// Test: default values are overwritten.
	// If a default value is provided but no env is found, the default is retained.
	type withDefaults struct {
		DefaultField1 string
		DefaultField2 string
	}

	os.Setenv("DEFAULT_FIELD_1", "vDEFAULT_FIELD_1")

	d = &Decoder{}
	cfgWithDefaults := withDefaults{
		DefaultField1: "default1", // should be overwritten.
		DefaultField2: "default2", // should persist with no env set.
	}
	err = d.Unmarshal(&cfgWithDefaults)
	assert.Nil(t, err)
	assert.Equal(t, "vDEFAULT_FIELD_1", cfgWithDefaults.DefaultField1)
	assert.Equal(t, "default2", cfgWithDefaults.DefaultField2)

	// can only assign "true", "false" or "" to type bool
	type badBool struct {
		BadBoolField bool
	}

	os.Setenv("BAD_BOOL_FIELD", "badvalue") // must be "true", "false", ""

	d = &Decoder{}
	cfgBadBool := badBool{}
	err = d.Unmarshal(&cfgBadBool)
	assert.EqualError(t, err, "'badvalue' from 'BAD_BOOL_FIELD' cannot be set to BadBoolField (bool)")

	// can only assign proper int value to int type.
	type badInt struct {
		BadIntField int
	}

	os.Setenv("BAD_INT_FIELD", "badvalue")

	d = &Decoder{}
	cfgBadInt := badInt{}
	err = d.Unmarshal(&cfgBadInt)
	assert.EqualError(t, err, "'badvalue' from 'BAD_INT_FIELD' cannot be set to BadIntField (int)")

	// test bad uint field
	type badUint struct {
		BadUintField uint
	}

	os.Setenv("BAD_UINT_FIELD", "badvalue")

	d = &Decoder{}
	cfgBadUint := badUint{}
	err = d.Unmarshal(&cfgBadUint)
	assert.EqualError(t, err, "'badvalue' from 'BAD_UINT_FIELD' cannot be set to BadUintField (uint)")

	// Test: pass in pointer of non-struct
	otherPtr := 5

	d = &Decoder{}
	err = d.Unmarshal(&otherPtr)
	assert.EqualError(t, err, "'*int' must be a non-nil pointer struct")

	// teardown: unset envs
	os.Clearenv()
}

