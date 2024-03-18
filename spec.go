// Package jsonspec implements a way to validate JSON objects.
package jsonspec

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

// A Type is one of the valid types for fields.
type Type string

// Constants for types.
const (
	Boolean  Type = "boolean"
	String   Type = "string"
	Integer  Type = "integer"
	Number   Type = "number"
	Datetime Type = "datetime"
	Object   Type = "object"
	Array    Type = "array"
)

// A Spec defines a specification for JSON values. It can be used to validate data.
type Spec struct {
	// Type is the type of the value.
	Type Type `json:"type"`

	// Description contains an optional description of the spec.
	Description string `json:"description,omitempty"`

	// Fields defines the fields of an object. Only relevant if Type is [Object].
	Fields map[string]Field `json:"fields,omitempty"`

	// Elements defines the elements of an array. Only relevant if Type is [Array].
	Elements *Spec `json:"elements,omitempty"`
}

// A Field defines one field in a JSON object.
type Field struct {
	Spec

	// Required is true if the field has to be set for the object to be valid.
	Required bool `json:"required,omitempty"`

	// Default sets a default value for the field.
	Default any `json:"default,omitempty"`

	// Tags is a list of custom tags.
	Tags []string `json:"tags,omitempty"`
}

// ValidateJSON returns an error if [value] doesn't match the spec.
func (s *Spec) ValidateJSON(data []byte) error {
	var input any
	err := json.Unmarshal(data, &input)
	if err != nil {
		return err
	}
	return s.Validate(input)
}

// Validate returns an error if [value] doesn't match the spec.
func (s *Spec) Validate(value any) error {
	switch s.Type {
	case Boolean:
		switch value.(type) {
		case bool:
		default:
			return errors.New("expected boolean value")
		}
	case String:
		switch value.(type) {
		case string:
		default:
			return errors.New("expected a string")
		}
	case Integer:
		switch v := value.(type) {
		case int, int64:
		case float64:
			if !(math.Floor(v) == v) {
				return errors.New("expected an integer")
			}
		default:
			return errors.New("expected an integer")
		}
	case Number:
		switch value.(type) {
		case int, int64, float64:
		default:
			return errors.New("expected a number")
		}
	case Datetime:
		switch value := value.(type) {
		case time.Time:
		case string:
			_, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return errors.New("expected a datetime in RFC3339 format")
			}
		default:
			return errors.New("expected a datetime in RFC3339 format")
		}
	case Object:
		object, ok := value.(map[string]any)
		if !ok {
			return errors.New("expected an object")
		}
		if s.Fields == nil {
			return nil
		}
		for name, field := range s.Fields {
			v := object[name]
			if v == nil {
				if field.Required {
					return fmt.Errorf("%s is required", name)
				}
			} else {
				if err := field.Spec.Validate(v); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
			}
		}
	case Array:
		array, ok := value.([]any)
		if !ok {
			return errors.New("expected an array")
		}
		if s.Elements == nil {
			return nil
		}
		for index, element := range array {
			if err := s.Elements.Validate(element); err != nil {
				return fmt.Errorf("element %d: %v", index, err)
			}
		}
	}
	return nil
}
