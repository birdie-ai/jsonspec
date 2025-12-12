package jsonspec

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// For automatically generates a [Spec] for an object.
func For(o any) (*Spec, error) {
	typ := reflect.TypeOf(o)
	if typ.Kind() == reflect.Pointer {
		return specForType(typ.Elem())
	}
	return specForType(typ)
}

func specForType(typ reflect.Type) (*Spec, error) {
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		return &Spec{Type: Datetime}, nil
	}
	switch typ.Kind() {
	case reflect.Bool:
		return &Spec{Type: Boolean}, nil
	case reflect.String:
		return &Spec{Type: String}, nil
	case reflect.Int:
		return &Spec{Type: Integer}, nil
	case reflect.Float64:
		return &Spec{Type: Number}, nil
	case reflect.Struct:
		return specForObject(typ)
	case reflect.Slice:
		return specForArray(typ)
	case reflect.Map:
		return specForMap(typ)
	}

	return nil, fmt.Errorf("cannot generate spec for type %v", typ)
}

func specForMap(typ reflect.Type) (*Spec, error) {
	// No way to generate spec for an arbitrary map, so in case it's used, we
	// just return that it's an object
	return &Spec{
		Type: Object,
	}, nil
}

func specForObject(typ reflect.Type) (*Spec, error) {
	numFields := typ.NumField()
	fields := make(map[string]Field)
	for i := 0; i < numFields; i++ {
		structField := typ.Field(i)
		name, field, err := fieldFor(structField)
		if err != nil {
			return nil, err
		}
		fields[name] = *field
	}
	spec := &Spec{
		Type:   Object,
		Fields: fields,
	}
	return spec, nil
}

func specForArray(typ reflect.Type) (*Spec, error) {
	elementSpec, err := specForType(typ.Elem())
	if err != nil {
		return nil, err
	}
	spec := &Spec{
		Type:     Array,
		Elements: elementSpec,
	}
	return spec, nil
}

// fieldFor generates a [Field] for a field in an object.
func fieldFor(structField reflect.StructField) (string, *Field, error) {
	name := translateName(structField.Name)
	spec, err := specForType(structField.Type)
	if err != nil {
		return "", nil, fmt.Errorf("field %s: %v", name, err)
	}
	field := &Field{Spec: *spec}
	err = parseTag(field, string(structField.Tag))
	if err != nil {
		return "", nil, err
	}
	return name, field, nil
}

var wordRe = regexp.MustCompile(`[A-Z][a-z]+`)

// translateName translates a name from the convention for Go struct fields to the convention for
// JSON objects, for example "UserID" to "user_id".
func translateName(name string) string {
	// split name into words
	var words [][]byte
	remaining := []byte(name)
	for len(remaining) > 0 {
		pair := wordRe.FindIndex(remaining)
		if pair == nil {
			words = append(words, remaining)
			break
		}
		from, to := pair[0], pair[1]
		if from > 0 {
			words = append(words, remaining[:from])
		}
		words = append(words, remaining[from:to])
		remaining = remaining[to:]
	}

	// join words with underscores
	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = strings.ToLower(string(w))
	}
	return strings.Join(parts, "_")
}

// match the first key:"value" pair in a tag
var tagRe = regexp.MustCompile(`^([a-z]+):("[^"]+")( +.*)?$`)

func parseTag(field *Field, tag string) error {
	remaining := tag
	for {
		remaining = strings.TrimSpace(remaining)
		if len(remaining) == 0 {
			return nil
		}
		matches := tagRe.FindStringSubmatch(remaining)
		if matches == nil {
			return errors.New("invalid tag")
		}
		key := matches[1]
		quotedValue := matches[2]
		remaining = matches[3]
		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			return fmt.Errorf("invalid value: %s", quotedValue)
		}
		err = applyTag(field, key, value)
		if err != nil {
			return err
		}
	}
}

func applyTag(field *Field, key, value string) error {
	switch key {
	case "description":
		field.Spec.Description = value
	case "required":
		b, ok := parseBool(value)
		if !ok {
			return fmt.Errorf("invalid boolean: %s", value)
		}
		field.Required = b
	case "tags":
		field.Tags = strings.Split(value, ",")
	case "default":
		defaultValue, err := parseDefaultValue(field.Spec.Type, value)
		if err != nil {
			return err
		}
		field.Default = defaultValue
	default:
		return fmt.Errorf("unknown tag key: %q", key)
	}
	return nil
}

func parseBool(s string) (result, ok bool) {
	switch s {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return
}

func parseDefaultValue(typ Type, s string) (any, error) {
	switch typ {
	case Boolean:
		return strconv.ParseBool(s)
	case String:
		return s, nil
	case Integer:
		return strconv.Atoi(s)
	case Number:
		return strconv.ParseFloat(s, 64)
	case Datetime:
		return time.Parse(time.RFC3339, s)
	}
	// setting a default for an object or array is not supported
	return nil, fmt.Errorf("cannot set default value for %s", typ)
}
