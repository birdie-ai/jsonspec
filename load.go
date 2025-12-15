package jsonspec

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

// LoadJSON load a JSON value from [source] into [target]. It returns an error in case of invalid
// JSON or in case the data doesn't match the spec for [target].
func LoadJSON(source []byte, target any) error {
	var input any
	err := json.Unmarshal(source, &input)
	if err != nil {
		return err
	}
	return Load(input, target)
}

// Load load [source] into [target]. It returns an error in case the data doesn't match the spec
// for [target].
func Load(source, target any) error {
	pointerType := reflect.TypeOf(target)
	if pointerType.Kind() != reflect.Pointer {
		return errors.New("argument to Load must be a pointer")
	}
	typ := pointerType.Elem()

	// create Spec
	spec, err := specForType(typ)
	if err != nil {
		return err
	}

	// validate source
	err = spec.Validate(source)
	if err != nil {
		return err
	}

	// load into target
	load(spec, source, reflect.ValueOf(target).Elem())

	return nil
}

func load(spec *Spec, input any, target reflect.Value) {
	switch spec.Type {
	case Boolean:
		target.SetBool(input.(bool))
	case String:
		target.SetString(input.(string))
	case Integer:
		switch i := input.(type) {
		case int:
			target.SetInt(int64(i))
		case int64:
			target.SetInt(i)
		case float64:
			target.SetInt(int64(i))
		}
	case Number:
		switch i := input.(type) {
		case int:
			target.SetFloat(float64(i))
		case int64:
			target.SetFloat(float64(i))
		case float64:
			target.SetFloat(i)
		}
	case Datetime:
		switch i := input.(type) {
		case time.Time:
			target.Set(reflect.ValueOf(i))
		case string:
			t, _ := time.Parse(time.RFC3339, i)
			target.Set(reflect.ValueOf(t))
		}
	case Object:
		var inputMap map[string]any
		if input != nil {
			inputMap = input.(map[string]any)
		}
		typ := target.Type()
		switch typ.Kind() {
		case reflect.Map:
			target.Set(reflect.ValueOf(inputMap))
		default:
			for i := 0; i < typ.NumField(); i++ {
				name := typ.Field(i).Name
				key := translateName(name)
				field, ok := spec.Fields[key]
				if ok {
					value := inputMap[key]
					if value == nil {
						value = field.Default
					}
					if value != nil {
						load(&field.Spec, value, target.FieldByName(name))
					}
				}
			}
		}
	case Array:
		var inputSlice []any
		if input != nil {
			inputSlice = input.([]any)
		}
		slice := reflect.MakeSlice(target.Type(), len(inputSlice), len(inputSlice))
		for i, v := range inputSlice {
			load(spec.Elements, v, slice.Index(i))
		}
		target.Set(slice)
	}
}
