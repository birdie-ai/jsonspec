package jsonspec

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func testLoadJSON[T any](t *testing.T, input string, want T) {
	t.Helper()

	var got T
	err := LoadJSON([]byte(input), &got)
	if err != nil {
		t.Errorf("Load(%q) returned error: %v", input, err)
		return
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Load(%q) result mismatch (-want +got):\n%s", input, diff)
	}
}

func TestLoadJSON(t *testing.T) {
	testLoadJSON(t, `true`, true)
	testLoadJSON(t, `"hello"`, "hello")
	testLoadJSON(t, `123`, 123)
	testLoadJSON(t, `123.0`, 123)
	testLoadJSON(t, `123`, 123.0)
	testLoadJSON(t, `123.0`, 123.0)
	testLoadJSON(t, `123.456`, 123.456)
	testLoadJSON(t, `"2024-03-07T11:38:47Z"`, time.Date(2024, 3, 7, 11, 38, 47, 0, time.UTC))
	testLoadJSON(t, `"2024-03-07T11:38:47.123456789Z"`, time.Date(2024, 3, 7, 11, 38, 47, 123456789, time.UTC))

	testLoadJSON(t,
		`{"first_name": "Jane", "last_name": "Doe"}`,
		struct {
			FirstName string
			LastName  string
		}{"Jane", "Doe"},
	)
	testLoadJSON(t,
		`{"first_name": "Jane"}`,
		struct {
			FirstName string
			LastName  string
		}{"Jane", ""},
	)
	testLoadJSON(t,
		`{"first_name": "Jane"}`,
		struct {
			FirstName string `default:"Jack"`
			LastName  string `default:"Smith"`
		}{"Jane", "Smith"},
	)

	testLoadJSON(t,
		`{"first_name": "Jane", "args":{"a":"1", "b":2, "c":{"set":true}}}`,
		struct {
			FirstName string
			Args      map[string]any
		}{FirstName: "Jane", Args: map[string]any{"a": "1", "b": 2.0, "c": map[string]any{"set": true}}},
	)

	testLoadJSON(t, `["jane", "joe", "julia"]`, []string{"jane", "joe", "julia"})
	testLoadJSON(t, `[1, 2, 3]`, []int{1, 2, 3})
	testLoadJSON(t, `[[1], [2, 3]]`, [][]int{{1}, {2, 3}})
}

func TestSpecLoad(t *testing.T) {
	type args struct {
		Host  string `config:"credential" required:"true"`
		Auth  string `config:"credential" default:"oauth"`
		Table string `config:"connection" required:"true"`
	}

	cases := []struct {
		name    string
		source  map[string]any
		want    args
		wantErr bool
	}{
		{
			name:   "loads the fields the spec keeps",
			source: map[string]any{"host": "https://source.test"},
			want:   args{Host: "https://source.test", Auth: "oauth"},
		},
		{
			name:   "ignores the fields the spec drops",
			source: map[string]any{"host": "https://source.test", "table": "events"},
			want:   args{Host: "https://source.test", Auth: "oauth"},
		},
		{
			name:    "still requires the fields the spec keeps",
			source:  map[string]any{"table": "events"},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spec, err := For(args{})
			if err != nil {
				t.Fatalf("For() returned error: %v", err)
			}
			delete(spec.Fields, "table")

			var got args
			err = spec.Load(c.source, &got)
			if c.wantErr {
				if err == nil {
					t.Fatalf("Load(%v) = nil, want error", c.source)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%v) returned error: %v", c.source, err)
			}
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("Load(%v) result mismatch (-want +got):\n%s", c.source, diff)
			}
		})
	}
}
