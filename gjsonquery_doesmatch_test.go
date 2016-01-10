package gjsonquery_test

import (
	"errors"
	"testing"

	. "github.com/szpakas/gjsonquery"
)

func TestDoesMatch(t *testing.T) {
	type subTestCase struct {
		symbol   string
		data     map[string]interface{}
		expected bool
		err      error
	}

	type testCase struct {
		symbol string
		query  interface{}
		tests  []subTestCase
		debug  bool
	}

	var tests = []testCase{
		// match on single
		{
			symbol: "AA",
			query:  map[string]interface{}{"a": 100},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 100}, true, nil},  // straight match
				{"ab", map[string]interface{}{"a": 201}, false, nil}, // incorrect value
				{"ac", map[string]interface{}{"b": 200}, false, nil}, // missing key, the same value on different key
				{"ad", map[string]interface{}{}, false, nil},         // empty data
				{"ae", map[string]interface{}{"a": 100, "b": 101, "c": 102}, true, nil},
				{"af", map[string]interface{}{"a": 200, "b": 201, "c": 202}, false, nil},
				// match but on incorrect types
				{"ba", map[string]interface{}{"a": "100"}, false, nil},
				{"bb", map[string]interface{}{"a": 100.0}, false, nil},
				{"bc", map[string]interface{}{"a": map[string]interface{}{"b": 100}}, false, nil},
			},
		},
		// match on multiple
		{
			symbol: "AB",
			query:  map[string]interface{}{"a": 100, "b": 101.0, "c": "102"},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 100, "b": 101.0, "c": "102"}, true, nil},
				{"ab", map[string]interface{}{"a": "100", "b": 101.0, "c": "102"}, false, nil},
				{"ac", map[string]interface{}{"a": 100, "b": 101, "c": "102"}, false, nil},
				{"ad", map[string]interface{}{"a": 100, "b": 101.0, "c": 102}, false, nil},
				{"ae", map[string]interface{}{"a": 100, "b": 101.0, "c": "102", "d": 501, "l1_e.l2_a": "502"}, true, nil},
			},
		},
		// match on default values
		{
			symbol: "AC",
			query:  map[string]interface{}{"a": 0, "b": 0.0, "c": ""},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 0, "b": 0.0, "c": ""}, true, nil},
				{"ab", map[string]interface{}{"b": 0.0, "c": ""}, false, nil},
				{"ac", map[string]interface{}{"a": 0, "c": ""}, false, nil},
				{"ad", map[string]interface{}{"a": 0, "b": 0.0}, false, nil},
			},
		},
		// match on empty AND query -> should always match
		{
			symbol: "AD",
			query:  map[string]interface{}{},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 0, "b": 0.0, "c": ""}, true, nil},
				{"ab", map[string]interface{}{"b": 0.0, "c": ""}, true, nil},
				{"ac", map[string]interface{}{}, true, nil},
			},
		},
		// match on nested
		{
			symbol: "BA",
			query:  map[string]interface{}{"l1_a.l2_a": 100},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": 100}}, true, nil},
				{"ab", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": nil}}, false, nil},
				{"aa", map[string]interface{}{"l1_a": map[string]interface{}{"l2_b": 100}}, false, nil},
			},
		},

		// $in: match on lists
		{
			symbol: "CA",
			query:  map[string]interface{}{"a": []interface{}{100, 101, 102}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 100}, true, nil},
				{"ab", map[string]interface{}{"a": 101}, true, nil},
				{"ac", map[string]interface{}{"a": 102}, true, nil},
				{"ba", map[string]interface{}{"a": 200}, false, nil},
				{"bb", map[string]interface{}{"b": 100}, false, nil},
			},
		},
		// $in: match on lists (deep)
		{
			symbol: "CB",
			query:  map[string]interface{}{"l1_a.l2_a.l3_a": []interface{}{100, 101, 102}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 101}}}, true, nil},
				{"ba", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 0}}}, false, nil},
				{"bb", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": map[string]interface{}{"l4_a": 101}}}}, false, nil},
			},
		},
		// $in: match via keyword
		{
			symbol: "CC",
			query:  map[string]interface{}{"a": map[string]interface{}{"$in": []interface{}{100, 101, 102}}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 100}, true, nil},
				{"ab", map[string]interface{}{"a": 101}, true, nil},
				{"ac", map[string]interface{}{"a": 102}, true, nil},
				{"ba", map[string]interface{}{"a": 200}, false, nil},
				{"bb", map[string]interface{}{"b": 100}, false, nil},
			},
		},
		// $in: match via keyword (deep)
		{
			symbol: "CD",
			query:  map[string]interface{}{"l1_a.l2_a.l3_a": map[string]interface{}{"$in": []interface{}{100, 101, 102}}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 100}}}, true, nil},
				{"ab", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 101}}}, true, nil},
				{"ac", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 102}}}, true, nil},
				{"ba", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": 0}}}, false, nil},
				{"bb", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": map[string]interface{}{"l4_a": 101}}}}, false, nil},
				{"bc", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": map[string]interface{}{}}}}, false, nil},
				{"bd", map[string]interface{}{"l1_a": map[string]interface{}{"l2_a": map[string]interface{}{"l3_a": []interface{}{}}}}, false, nil},
			},
		},
		// !$in: match via keyword
		{
			symbol: "CE",
			query:  map[string]interface{}{"a": map[string]interface{}{"!$in": []interface{}{100, 101, 102}}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 100}, false, nil},
				{"ab", map[string]interface{}{"a": 101}, false, nil},
				{"ac", map[string]interface{}{"a": 102}, false, nil},
				{"ba", map[string]interface{}{"a": 200}, true, nil},
				{"bb", map[string]interface{}{"b": 100}, true, nil},
			},
		},
		// $and: match via keyword
		{
			symbol: "DA",
			query: map[string]interface{}{
				"$and": []interface{}{
					map[string]interface{}{"a": 101}, map[string]interface{}{"b": 102},
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101, "b": 102}, true, nil},
				{"ab", map[string]interface{}{"a": 101, "b": 102, "c": 103}, true, nil},
				{"ba", map[string]interface{}{"a": 201, "b": 102}, false, nil},
				{"bb", map[string]interface{}{"a": 101, "b": 202}, false, nil},
				{"ca", map[string]interface{}{"a": 101}, false, nil},
			},
		},
		// $or: match on list
		{
			symbol: "EA",
			query: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{"a": 101},
					map[string]interface{}{"b": 102},
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, true, nil},
				{"ab", map[string]interface{}{"b": 102}, true, nil},
				{"ba", map[string]interface{}{"a": 201, "b": 102}, true, nil},
				{"bb", map[string]interface{}{"a": 101, "b": 202}, true, nil},
				{"ca", map[string]interface{}{"a": 201, "b": 202}, false, nil},
				{"cb", map[string]interface{}{}, false, nil},
			},
		},
		// $or: match on map
		{
			symbol: "EB",
			query: map[string]interface{}{
				"$or": map[string]interface{}{"a": 101, "b": 102},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, true, nil},
				{"ab", map[string]interface{}{"b": 102}, true, nil},
				{"ba", map[string]interface{}{"a": 201, "b": 102}, true, nil},
				{"bb", map[string]interface{}{"a": 101, "b": 202}, true, nil},
				{"ca", map[string]interface{}{"a": 201, "b": 202}, false, nil},
				{"cb", map[string]interface{}{}, false, nil},
			},
		},
		// $or: match on empty list -> should never match
		{
			symbol: "EC",
			query: map[string]interface{}{
				"$or": map[string]interface{}{},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101, "b": 102.0, "c": "103"}, false, nil},
				{"ab", map[string]interface{}{"b": 0.0, "c": ""}, false, nil},
				{"ac", map[string]interface{}{}, false, nil},
			},
		},
		// $notAnd
		{
			symbol: "FA",
			query:  map[string]interface{}{"$not": map[string]interface{}{"a": 100}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 200}, true, nil},
				{"ab", map[string]interface{}{"a": 100}, false, nil},
				{"ac", map[string]interface{}{}, true, nil},
				{"ad", map[string]interface{}{"b": 100}, true, nil},
			},
		},
		// $andNot
		{
			symbol: "FB",
			query:  map[string]interface{}{"a": map[string]interface{}{"$not": 100}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 200}, true, nil},
				{"ab", map[string]interface{}{"a": 100}, false, nil},
				{"ac", map[string]interface{}{}, true, nil},
			},
		},
		// $not in list
		{
			symbol: "FC",
			query:  map[string]interface{}{"a": map[string]interface{}{"$not": []interface{}{100, 101, 102}}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 200}, true, nil},
				{"ab", map[string]interface{}{"a": 100}, false, nil},
				{"ac", map[string]interface{}{}, true, nil},
				{"ad", map[string]interface{}{"b": 100}, true, nil},
			},
		},
		// !$and
		{
			symbol: "GA",
			query: map[string]interface{}{
				"!$and": []interface{}{
					map[string]interface{}{"a": 101}, map[string]interface{}{"b": 102},
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101, "b": 102}, false, nil},
				{"ab", map[string]interface{}{"a": 101, "b": 102, "c": 103}, false, nil},
				{"ba", map[string]interface{}{"a": 201, "b": 102}, true, nil},
				{"bb", map[string]interface{}{"a": 101, "b": 202}, true, nil},
				{"ca", map[string]interface{}{"a": 101}, true, nil},
			},
		},
		// !!$and (double negation)
		{
			symbol: "GB",
			query: map[string]interface{}{
				"!!$and": []interface{}{
					map[string]interface{}{"a": 101}, map[string]interface{}{"b": 102},
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101, "b": 102}, true, nil},
				{"ab", map[string]interface{}{"a": 101, "b": 102, "c": 103}, true, nil},
				{"ba", map[string]interface{}{"a": 201, "b": 102}, false, nil},
				{"bb", map[string]interface{}{"a": 101, "b": 202}, false, nil},
				{"ca", map[string]interface{}{"a": 101}, false, nil},
			},
		},
		// $and $is
		{
			symbol: "GC",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$is": 101},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, true, nil},
				{"ab", map[string]interface{}{"a": 201}, false, nil},
				{"ac", map[string]interface{}{"b": 101}, false, nil},
				{"ad", map[string]interface{}{}, false, nil},
			},
		},
		// $and !$is
		{
			symbol: "GD",
			query: map[string]interface{}{
				"a": map[string]interface{}{"!$is": 101},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, nil},
				{"ab", map[string]interface{}{"a": 201}, true, nil},
				{"ac", map[string]interface{}{"b": 101}, true, nil},
				{"ad", map[string]interface{}{}, true, nil},
			},
		},
		// $and !!$is
		{
			symbol: "GE",
			query: map[string]interface{}{
				"a": map[string]interface{}{"!!$is": 101},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, true, nil},
			},
		},
		// multiple matchers/comparator on one level
		// TODO: change after implementing
		{
			symbol: "HA",
			query: map[string]interface{}{
				"a": map[string]interface{}{
					"$in":  []interface{}{101, 102},
					"$not": 201,
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchValue: multiple expectations")},
			},
		},
		// TODO: change after implementing
		{
			symbol: "HB",
			query: map[string]interface{}{
				"a": map[string]interface{}{
					"$in":  []interface{}{101, 102},
					"$not": 101,
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchValue: multiple expectations")},
			},
			//			debug: true,
		},
		// arithmetic comparators
		// $gt as Int
		{
			symbol: "IA",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$gt": 105},
			},
			tests: []subTestCase{
				// native type
				{"aa", map[string]interface{}{"a": 206}, true, nil},
				{"ab", map[string]interface{}{"a": 105}, false, nil},
				{"ac", map[string]interface{}{"a": 100}, false, nil},
				{"ad", map[string]interface{}{"a": 0}, false, nil},
				{"ae", map[string]interface{}{"a": -100}, false, nil},
				{"af", map[string]interface{}{"a": -105}, false, nil},
				{"ag", map[string]interface{}{"a": -106}, false, nil},
				// automatic casting to native type
				{"ba", map[string]interface{}{"a": 206.0}, true, nil},
				{"bb", map[string]interface{}{"a": 105.0}, false, nil},
				{"bc", map[string]interface{}{"a": 100.0}, false, nil},
				{"bd", map[string]interface{}{"a": 0.0}, false, nil},
				{"be", map[string]interface{}{"a": -100.0}, false, nil},
				{"bf", map[string]interface{}{"a": -105.0}, false, nil},
				{"bg", map[string]interface{}{"a": -106.0}, false, nil},
				// casting impossible
				{"ca", map[string]interface{}{"a": "206"}, false, errors.New("comparator: casting actual to Int failed.")},
				{"cb", map[string]interface{}{"a": nil}, false, errors.New("comparator: casting actual to Int failed.")},
			},
		},
		// $gt as float64
		{
			symbol: "IB",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$gt": 105.3},
			},
			tests: []subTestCase{
				// native type
				{"aa", map[string]interface{}{"a": 206.0}, true, nil},
				{"ab", map[string]interface{}{"a": 105.0}, false, nil},
				{"ac", map[string]interface{}{"a": 100.0}, false, nil},
				{"ad", map[string]interface{}{"a": 0.0}, false, nil},
				{"ae", map[string]interface{}{"a": -100.0}, false, nil},
				{"af", map[string]interface{}{"a": -105.0}, false, nil},
				{"ag", map[string]interface{}{"a": -106.0}, false, nil},
				// automatic casting to native type
				{"ba", map[string]interface{}{"a": 206}, true, nil},
				{"bb", map[string]interface{}{"a": 100}, false, nil},
				{"bc", map[string]interface{}{"a": 0}, false, nil},
				{"be", map[string]interface{}{"a": -100}, false, nil},
				{"bf", map[string]interface{}{"a": -105}, false, nil},
				{"bg", map[string]interface{}{"a": -106}, false, nil},
				// casting impossible
				{"ca", map[string]interface{}{"a": "206"}, false, errors.New("comparator: casting actual to Float64 failed.")},
				{"cb", map[string]interface{}{"a": nil}, false, errors.New("comparator: casting actual to Float64 failed.")},
			},
		},
		// $gte as int
		{
			symbol: "IC",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$gte": 105},
			},
			tests: []subTestCase{
				// native type
				{"aa", map[string]interface{}{"a": 206}, true, nil},
				{"ab", map[string]interface{}{"a": 105}, true, nil},
				{"ac", map[string]interface{}{"a": 100}, false, nil},
				// automatic casting to native type
				{"ba", map[string]interface{}{"a": 206.0}, true, nil},
				{"bb", map[string]interface{}{"a": 105.0}, true, nil},
				{"bc", map[string]interface{}{"a": 100.0}, false, nil},
				// casting impossible
				{"ca", map[string]interface{}{"a": "206"}, false, errors.New("comparator: casting actual to Int failed.")},
				{"cb", map[string]interface{}{"a": nil}, false, errors.New("comparator: casting actual to Int failed.")},
			},
		},
		// $gte as float64
		{
			symbol: "ID",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$gte": 105.0},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 206.0}, true, nil},
				{"ab", map[string]interface{}{"a": 105.0}, true, nil},
				{"ac", map[string]interface{}{"a": 100.0}, false, nil},
				{"ad", map[string]interface{}{"a": 100}, false, nil},
			},
		},
		// $lt as int
		{
			symbol: "IE",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$lt": 105},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 104}, true, nil},
				{"ab", map[string]interface{}{"a": 105}, false, nil},
				{"ac", map[string]interface{}{"a": 106}, false, nil},
				{"ac", map[string]interface{}{"a": 106.0}, false, nil},
			},
		},
		// $lt as float
		{
			symbol: "IF",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$lt": 105.0},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 104.0}, true, nil},
				{"ab", map[string]interface{}{"a": 105.0}, false, nil},
				{"ac", map[string]interface{}{"a": 106}, false, nil},
				{"ac", map[string]interface{}{"a": 106.0}, false, nil},
			},
		},
		// $lte as int
		{
			symbol: "IG",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$lte": 105},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 104}, true, nil},
				{"ab", map[string]interface{}{"a": 105}, true, nil},
				{"ac", map[string]interface{}{"a": 106}, false, nil},
				{"ac", map[string]interface{}{"a": 106.0}, false, nil},
			},
		},
		// $lte as int
		{
			symbol: "IH",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$lte": 105.0},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 104.0}, true, nil},
				{"ab", map[string]interface{}{"a": 105.0}, true, nil},
				{"ac", map[string]interface{}{"a": 106}, false, nil},
				{"ac", map[string]interface{}{"a": 106.0}, false, nil},
			},
		},
		// $unknownComparator
		{
			symbol: "ZA",
			query:  map[string]interface{}{"$unknownComparator": 101},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchComparator: unknown comparator")},
			},
		},
		// $and with unknown structure -> failed due to unknown structure
		{
			symbol: "ZB",
			query:  map[string]interface{}{"$and": 101},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matcherAnd: unknown query type")},
			},
		},
		// $or with unknown structure -> failed due to unknown structure
		{
			symbol: "ZC",
			query:  map[string]interface{}{"$or": 101},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matcherAnd: unknown query type")},
			},
		},
		// $in with unknown structure
		{
			symbol: "ZC",
			query:  map[string]interface{}{"$in": 101},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("comparatorIn: unknown expected type")},
			},
		},
		{
			symbol: "ZD",
			query:  map[string]interface{}{"!$unknown": 101},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchComparator: unknown comparator")},
			},
		},
		{
			symbol: "ZE",
			query:  map[string]interface{}{"$not": map[string]interface{}{"$unknown": 101}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchComparator: unknown comparator")},
			},
		},
		{
			symbol: "ZF",
			query:  map[string]interface{}{"a": map[int]interface{}{123: 101}},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchValue: not a map")},
			},
		},
		{
			symbol: "ZG",
			query: []interface{}{
				map[string]interface{}{
					"$unknown": 101,
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchComparator: unknown comparator")},
			},
		},
		{
			symbol: "ZH",
			query: map[string]interface{}{
				"$or": map[string]interface{}{
					"$unknown": 101,
				},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matchComparator: unknown comparator")},
			},
		},
		{
			symbol: "ZI",
			query: map[string]interface{}{
				"$or": []interface{}{101},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("matcherAnd: unknown query type")},
			},
		},
		{
			symbol: "ZJ",
			query: map[string]interface{}{
				"a": map[string]interface{}{"$lt": "105"},
			},
			tests: []subTestCase{
				{"aa", map[string]interface{}{"a": 101}, false, errors.New("comparator: unknown type (type: string)")},
			},
			debug: true,
		},

		// TODO: $contains, $and $contains
		// TODO: full structure tests
	}

	for _, tDef := range tests {
		DEBUG = tDef.debug
		for _, tCase := range tDef.tests {
			result, err := DoesMatch(tDef.query, tCase.data)

			if (err != nil) || (tCase.err != nil) {
				var (
					expectedString string
					gotString      string
				)
				if tCase.err != nil {
					expectedString = tCase.err.Error()
				}
				if err != nil {
					gotString = err.Error()
				}
				if expectedString != gotString {
					t.Errorf("[%s|%s] Mismatch on error => expected: %#+v, have: %#+v", tDef.symbol, tCase.symbol, tCase.err, err)
				}
			}

			if result != tCase.expected {
				t.Errorf("[%s|%s] Mismatch => expected: %#+v, have: %#+v", tDef.symbol, tCase.symbol, tCase.expected, result)
			}
		}
	}
}
