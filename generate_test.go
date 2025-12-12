package jsonspec

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSpecFor(t *testing.T) {
	object := struct{ Admin bool }{}
	objectSpec := Spec{Type: Object, Fields: map[string]Field{"admin": {Spec: Spec{Type: Boolean}}}}
	array := []string{}
	arraySpec := Spec{Type: Array, Elements: &Spec{Type: String}}
	mapObj := map[string]any{"admin": true}
	mapSpec := Spec{Type: Object}
	cases := []struct {
		o    any
		want Spec
	}{
		{false, Spec{Type: Boolean}},
		{"", Spec{Type: String}},
		{0, Spec{Type: Integer}},
		{0.0, Spec{Type: Number}},
		{time.Time{}, Spec{Type: Datetime}},
		{object, objectSpec},
		{&object, objectSpec},
		{array, arraySpec},
		{&array, arraySpec},
		{mapObj, mapSpec},
		{&mapObj, mapSpec},
	}
	for _, c := range cases {
		got, err := For(c.o)
		if err != nil {
			t.Errorf("For(%v) returned error: %v", c.o, err)
		}
		if diff := cmp.Diff(c.want, *got); diff != "" {
			t.Errorf("For(%v) result mismatch (-want +got):\n%s", c.o, diff)
		}
	}
}

func TestSpecForObject(t *testing.T) {
	var object struct {
		Username string
		Password string `tags:"secret,important"`
	}
	want := &Spec{
		Type: Object,
		Fields: map[string]Field{
			"username": {Spec: Spec{Type: String}},
			"password": {Spec: Spec{Type: String}, Tags: []string{"secret", "important"}},
		},
	}
	typ := reflect.TypeOf(object)
	got, err := specForObject(typ)
	if err != nil {
		t.Fatalf("specForObject(%v) returned error: %v", typ, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("specForObject(%v) result mismatch (-want +got):\n%s", typ, diff)
	}
}

func TestSpecForArray(t *testing.T) {
	cases := []struct {
		array any
		want  Spec
	}{
		{
			[]string{},
			Spec{Type: Array, Elements: &Spec{Type: String}},
		},
		{
			[]time.Time{},
			Spec{Type: Array, Elements: &Spec{Type: Datetime}},
		},
		{
			[][]float64{},
			Spec{Type: Array, Elements: &Spec{Type: Array, Elements: &Spec{Type: Number}}},
		},
		{
			[]struct {
				FirstName string `required:"true"`
			}{},
			Spec{Type: Array, Elements: &Spec{
				Type: Object,
				Fields: map[string]Field{
					"first_name": {Spec: Spec{Type: String}, Required: true},
				},
			}},
		},
	}
	for _, c := range cases {
		typ := reflect.TypeOf(c.array)
		got, err := specForArray(typ)
		if err != nil {
			t.Errorf("specForArray(%v) returned error: %v", typ, err)
			continue
		}
		if diff := cmp.Diff(c.want, *got); diff != "" {
			t.Errorf("specForArray(%v) result mismatch (-want +got):\n%s", typ, diff)
		}
	}
}

func TestFieldFor(t *testing.T) {
	cases := []struct {
		o     any
		name  string
		field Field
	}{
		{
			struct{ Admin bool }{},
			"admin",
			Field{Spec: Spec{Type: Boolean}},
		},
		{
			struct{ FirstName string }{},
			"first_name",
			Field{Spec: Spec{Type: String}},
		},
		{
			struct {
				ID int `required:"true"`
			}{},
			"id",
			Field{Spec: Spec{Type: Integer}, Required: true},
		},
		{
			struct {
				ID int `required:"true" description:"Unique identifier"`
			}{},
			"id",
			Field{
				Spec:     Spec{Type: Integer, Description: "Unique identifier"},
				Required: true,
			},
		},
		{
			struct {
				UserPWD string `tags:"secret"`
			}{},
			"user_pwd",
			Field{Spec: Spec{Type: String}, Tags: []string{"secret"}},
		},
		{
			struct{ Words []string }{},
			"words",
			Field{Spec: Spec{Type: Array, Elements: &Spec{Type: String}}},
		},
		{
			struct {
				IntercomAPI struct {
					AccessToken string `tags:"secret"`
				}
			}{},
			"intercom_api",
			Field{Spec: Spec{
				Type: Object,
				Fields: map[string]Field{
					"access_token": {Spec: Spec{Type: String}, Tags: []string{"secret"}},
				},
			}},
		},
	}
	for _, c := range cases {
		structField := reflect.TypeOf(c.o).Field(0)
		name, field, err := fieldFor(structField)
		if err != nil {
			t.Errorf("fieldFor(%v) returned error: %v", structField, err)
			continue
		}
		if name != c.name {
			t.Errorf("fieldFor(%v) returned name %q, want %q", structField, name, c.name)
		}
		if diff := cmp.Diff(c.field, *field); diff != "" {
			t.Errorf("fieldFor(%v) result mismatch (-want +got):\n%s", structField, diff)
		}
	}
}

func TestTranslateName(t *testing.T) {
	cases := []struct {
		name, want string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"AdminUserIDValidationRuleError", "admin_user_id_validation_rule_error"},
		{"ID", "id"},
		{"UserID", "user_id"},
		{"URLPrefix", "url_prefix"},
		{"PlanB", "plan_b"},
		{"XChromosome", "x_chromosome"},
		{"FromAToB", "from_a_to_b"},
	}
	for _, c := range cases {
		got := translateName(c.name)
		if got != c.want {
			t.Errorf("translateName(%q) == %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseTag(t *testing.T) {
	cases := []struct {
		tag  string
		want *Field
	}{
		{
			``,
			&Field{},
		},
		{
			`required:"true"`,
			&Field{Required: true},
		},
		{
			`required:"true" description:"How's it going?"`,
			&Field{Required: true, Spec: Spec{Description: "How's it going?"}},
		},
		{
			`required:"true" tags:"secret"`,
			&Field{Required: true, Tags: []string{"secret"}},
		},
		{
			`required:"true"     tags:"secret"`,
			&Field{Required: true, Tags: []string{"secret"}},
		},
	}
	for _, c := range cases {
		field := new(Field)
		err := parseTag(field, c.tag)
		if err != nil {
			t.Errorf("parseTag(field, %v) returned error: %v", c.tag, err)
			continue
		}
		if diff := cmp.Diff(c.want, field); diff != "" {
			t.Errorf("parseTag(field, %v) result mismatch (-want +got):\n%s", c.tag, diff)
		}
	}
}

func TestParseTagError(t *testing.T) {
	cases := []string{
		`required`,
		`"required:true"`,
	}
	for _, c := range cases {
		field := new(Field)
		err := parseTag(field, c)
		if err == nil {
			t.Errorf("parseTag(field, %v) did not return error", c)
		}
	}
}

func TestApplyTag(t *testing.T) {
	cases := []struct {
		key, value  string
		field, want *Field
	}{
		{"description", "Hello there", new(Field), &Field{Spec: Spec{Description: "Hello there"}}},
		{"required", "false", new(Field), new(Field)},
		{"required", "true", new(Field), &Field{Required: true}},
		{"tags", "secret,important,spicy", new(Field), &Field{Tags: []string{"secret", "important", "spicy"}}},
		{
			"default",
			"123",
			&Field{Spec: Spec{Type: Integer}},
			&Field{Spec: Spec{Type: Integer}, Default: 123},
		},
	}
	for _, c := range cases {
		err := applyTag(c.field, c.key, c.value)
		if err != nil {
			t.Errorf("applyTag(%v, %q, %q) returned error: %v", c.field, c.key, c.value, err)
			continue
		}
		if diff := cmp.Diff(c.want, c.field); diff != "" {
			t.Errorf("applyTag(%v, %q, %q) result mismatch (-want +got):\n%s", c.field, c.key, c.value, diff)
		}
	}
}

func TestApplyTagError(t *testing.T) {
	cases := []struct {
		field      *Field
		key, value string
	}{
		{new(Field), "hello", "true"},
		{new(Field), "required", "hello"},
		{
			&Field{Spec: Spec{Type: Integer}},
			"default",
			"hello",
		},
	}
	for _, c := range cases {
		err := applyTag(c.field, c.key, c.value)
		if err == nil {
			t.Errorf("applyTag(%v, %q, %q) returned err == nil", c.field, c.key, c.value)
		}
	}
}

func TestParseDefaultValue(t *testing.T) {
	cases := []struct {
		typ   Type
		value string
		want  any
	}{
		{Boolean, "true", true},
		{String, "hello", "hello"},
		{Integer, "123", 123},
		{Number, "123.456", 123.456},
		{Datetime, "2024-03-07T11:38:47Z", time.Date(2024, 3, 7, 11, 38, 47, 0, time.UTC)},
	}
	for _, c := range cases {
		got, err := parseDefaultValue(c.typ, c.value)
		if err != nil {
			t.Errorf("parseDefaultValue(%v, %v) returned error: %v", c.typ, c.value, err)
			continue
		}
		if diff := cmp.Diff(c.want, got); diff != "" {
			t.Errorf("parseDefaultValue(%v, %v) result mismatch (-want +got):\n%s", c.typ, c.value, diff)
		}
	}
}
