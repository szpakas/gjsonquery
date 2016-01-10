package gjsonquery_test

import (
	"os"
	"reflect"
	"testing"

	. "github.com/szpakas/gjsonquery"
)

func TestFetchValue(t *testing.T) {
	if os.Getenv("TESTS_SKIP_FETCH_VALUE") == "yes" {
		t.Skip("Skiping test as ordered")
	}

	type subTestCase struct {
		symbol          string
		column          string
		expectedValue   interface{}
		expectedSuccess bool
	}

	type testCase struct {
		symbol string
		data   map[string]interface{}
		tests  []subTestCase
	}

	var tests = []testCase{
		{
			symbol: "AA",
			data:   map[string]interface{}{"a": 100, "b": "101", "c": 102.0, "d": 0, "e": "", "f": 0.0},
			tests: []subTestCase{
				{"aa", "a", 100, true},
				{"ab", "b", "101", true},
				{"ac", "c", 102.0, true},
				{"ad", "d", 0, true},
				{"ae", "e", "", true},
				{"af", "f", 0.0, true},
				{"bb", "x", nil, false},
			},
		},
		// nested
		{
			symbol: "BA",
			data:   map[string]interface{}{"l1_a": 100, "l1_b": map[string]interface{}{"l2_a": 101}},
			tests: []subTestCase{
				{"aa", "l1_a", 100, true},
				{"ab", "l1_b", map[string]interface{}{"l2_a": 101}, true},
				{"ac", "l1_b.l2_a", 101, true},
				{"ad", "l1_a.l2_b", nil, false},
			},
		},
	}

	for _, tDef := range tests {
		for _, tCase := range tDef.tests {
			value, success := FetchValue(tDef.data, tCase.column)

			if !reflect.DeepEqual(value, tCase.expectedValue) || (success != tCase.expectedSuccess) {
				t.Errorf("[%s|%s] Mismatch\nexpected => success: % 5v, value: %#+v\n    have => success: % 5v, value: %#+v",
					tDef.symbol, tCase.symbol, tCase.expectedSuccess, tCase.expectedValue, success, value)
			}
		}
	}
}
