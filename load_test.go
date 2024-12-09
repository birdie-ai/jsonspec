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

	testLoadJSON(t, `["jane", "joe", "julia"]`, []string{"jane", "joe", "julia"})
	testLoadJSON(t, `[1, 2, 3]`, []int{1, 2, 3})
	testLoadJSON(t, `[[1], [2, 3]]`, [][]int{{1}, {2, 3}})
}
