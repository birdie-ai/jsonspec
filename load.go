package jsonspec

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

// Load a JSON value from [data] into [o]. It returns an error if the data is invalid.
func Load(data []byte, o any) error {
	pointerType := reflect.TypeOf(o)
	if pointerType.Kind() != reflect.Pointer {
		return errors.New("argument to Load must be a pointer")
	}
	typ := pointerType.Elem()

	// create Spec
	spec, err := specForType(typ)
	if err != nil {
		return err
	}

	// parse JSON
	var input any
	err = json.Unmarshal(data, &input)
	if err != nil {
		return err
	}

	// validate input
	err = spec.Validate(input)
	if err != nil {
		return err
	}

	// load into o
	load(spec, input, reflect.ValueOf(o).Elem())

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
