package jsonspec

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func testLoad[T any](t *testing.T, input string, want T) {
	t.Helper()

	var got T
	err := Load([]byte(input), &got)
	if err != nil {
		t.Errorf("Load(%q) returned error: %v", input, err)
		return
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Load(%q) result mismatch (-want +got):\n%s", input, diff)
	}
}

func TestLoad(t *testing.T) {
	testLoad(t, `true`, true)
	testLoad(t, `"hello"`, "hello")
	testLoad(t, `123`, 123)
	testLoad(t, `123.0`, 123)
	testLoad(t, `123`, 123.0)
	testLoad(t, `123.0`, 123.0)
	testLoad(t, `123.456`, 123.456)
	testLoad(t, `"2024-03-07T11:38:47Z"`, time.Date(2024, 3, 7, 11, 38, 47, 0, time.UTC))
	testLoad(t, `"2024-03-07T11:38:47.123456789Z"`, time.Date(2024, 3, 7, 11, 38, 47, 123456789, time.UTC))

	testLoad(t,
		`{"first_name": "Jane", "last_name": "Doe"}`,
		struct {
			FirstName string
			LastName  string
		}{"Jane", "Doe"},
	)
	testLoad(t,
		`{"first_name": "Jane"}`,
		struct {
			FirstName string
			LastName  string
		}{"Jane", ""},
	)
	testLoad(t,
		`{"first_name": "Jane"}`,
		struct {
			FirstName string `default:"Jack"`
			LastName  string `default:"Smith"`
		}{"Jane", "Smith"},
	)

	testLoad(t, `["jane", "joe", "julia"]`, []string{"jane", "joe", "julia"})
	testLoad(t, `[1, 2, 3]`, []int{1, 2, 3})
	testLoad(t, `[[1], [2, 3]]`, [][]int{{1}, {2, 3}})
}
