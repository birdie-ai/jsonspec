package jsonspec

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSpecAsJSON(t *testing.T) {
	// marshal and unmarshal a Spec to make sure it works with encoding/json
	spec := &Spec{
		Description: "Object describing a person",
		Fields: map[string]Field{
			"first_name": {
				Spec: Spec{
					Type:        String,
					Description: "First (given) name",
				},
				Required: true,
			},
		},
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	got := new(Spec)
	err = json.Unmarshal(data, got)
	if err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if diff := cmp.Diff(spec, got); diff != "" {
		t.Errorf("spec changed after marshaling & unmarshaling (-want +got):\n%s", diff)
	}
}

type NestedStruct struct {
	Customers []struct {
		Name           string `required:"true"`
		ContactDetails struct {
			PhoneNumbers []struct {
				CountryCode string
				Number      string `required:"true"`
			}
		}
	}
}

func TestSpecValidateSuccess(t *testing.T) {
	cases := []struct {
		o, value any
	}{
		{false, true},
		{"", "hello"},
		{0, 4},
		{0, 4.0},
		{0.0, 4},
		{0.0, 4.1},
		{time.Time{}, "2024-03-06T12:30:23Z"},
		{
			struct{ FirstName string }{},
			map[string]any{"first_name": "Jane"},
		},
		{
			struct {
				UserID   int    `required:"true"`
				Password string `required:"true" tags:"secret"`
			}{},
			map[string]any{"user_id": 4, "password": "Password1"},
		},
		{
			[]int{},
			[]any{1, 2, 3},
		},
		{
			[]struct {
				FirstName string
				LastName  string `required:"true"`
			}{},
			[]any{
				map[string]any{"first_name": "Jane", "last_name": "Doe"},
				map[string]any{"last_name": "Doe"},
			},
		},
		{
			NestedStruct{},
			map[string]any{
				"customers": []any{
					map[string]any{
						"name": "Jane Doe",
						"contact_details": map[string]any{
							"phone_numbers": []any{
								map[string]any{
									"country_code": "001",
									"number":       "123 555 1234",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		spec, err := For(c.o)
		if err != nil {
			t.Fatalf("For(%v) returned error: %v", c.o, err)
		}
		err = spec.Validate(c.value)
		if err != nil {
			t.Errorf("spec.Validate(%v) returned error: %v", c.value, err)
		}
	}
}

func TestSpecValidateError(t *testing.T) {
	// The test cases include the actual error messages because having good error messages is one of
	// the goals of the library.
	cases := []struct {
		o, value any
		want     string
	}{
		{false, 123, "expected boolean value"},
		{"", 123, "expected a string"},
		{123, "hello", "expected an integer"},
		{123, 123.456, "expected an integer"},
		{0.0, "hello", "expected a number"},
		{time.Time{}, "hello", "expected a datetime in RFC3339 format"},
		{
			struct{ FirstName string }{},
			"hello",
			"expected an object",
		},
		{
			struct{ FirstName string }{},
			map[string]any{"first_name": 123},
			"first_name: expected a string",
		},
		{
			struct {
				FirstName string `required:"true"`
			}{},
			map[string]any{},
			"first_name is required",
		},
		{
			[]int{},
			"hello",
			"expected an array",
		},
		{
			[]int{},
			[]any{"hello"},
			"element 0: expected an integer",
		},
		{
			[]int{},
			[]any{0, 1, 2, true, 4},
			"element 3: expected an integer",
		},
		{
			NestedStruct{},
			map[string]any{
				"customers": []any{
					map[string]any{
						"name": "Jane Doe",
						"contact_details": map[string]any{
							"phone_numbers": []any{
								map[string]any{
									"country_code": "001",
								},
							},
						},
					},
				},
			},
			"customers: element 0: contact_details: phone_numbers: element 0: number is required",
		},
	}
	for _, c := range cases {
		spec, err := For(c.o)
		if err != nil {
			t.Fatalf("For(%v) returned error: %v", c.o, err)
		}
		err = spec.Validate(c.value)
		if err == nil {
			t.Errorf("spec.Validate(%v) did not return error", c.value)
			continue
		}
		got := err.Error()
		if got != c.want {
			t.Errorf("spec.Validate(%v) returned %q, want %q", c.value, got, c.want)
		}
	}
}
