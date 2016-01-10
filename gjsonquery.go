package gjsonquery

import (
	"errors"
	"strings"
)

const COLUMN_LEVEL_SEPARATOR string = "."

func DoesMatch(query interface{}, data map[string]interface{}) (bool, error) {
	// first level match is always AND
	return matcherAnd(query, data)
}

// -- matchers

func matcherAnd(query interface{}, data map[string]interface{}) (bool, error) {

	switch v := interface{}(query).(type) {
	// -- query is a key->value map
	case map[string]interface{}:
		for column, expectedValue := range v {
			matched, err := matchValue(column, expectedValue, data)
			if err != nil {
				return false, err
			}
			if !matched {
				// first mismatch determine result -> no match
				return false, nil
			}
		}
		// defaults to match if nothing failed first
		return true, nil
	// -- query is a list
	case []interface{}:
		for _, expectedValue := range v {
			// list match is always and
			matched, err := matcherAnd(expectedValue, data)
			if err != nil {
				return false, err
			}
			if !matched {
				// first mismatch determine result -> no match
				return false, nil
			}
		}
		// defaults to match if nothing failed first
		return true, nil
	}

	// unknown type
	return false, errors.New("matcherAnd: unknown query type")
}

func matcherOr(query interface{}, data map[string]interface{}) (bool, error) {

	switch v := interface{}(query).(type) {
	// -- query is a key->value map
	case map[string]interface{}:
		for column, expectedValue := range v {
			matched, err := matchValue(column, expectedValue, data)
			if err != nil {
				return false, err
			}
			if matched {
				// one passed match is enough
				return true, nil
			}
		}
		// defaults to match if nothing failed first
		return false, nil
	// -- query is a list
	case []interface{}:
		for _, expectedValue := range v {
			// list match is always and
			matched, err := matcherAnd(expectedValue, data)
			if err != nil {
				return false, err
			}
			if matched {
				// one passed match is enough
				return true, nil
			}
		}
		// defaults to NO match if nothing matched first
		return false, nil
	}

	// unknown type
	return false, errors.New("matcherAnd: unknown query type")
}

// -- logic/helpers

type comparatorType int

const (
	COMPARATOR_NOT comparatorType = iota + 1
	COMPARATOR_IS
	COMPARATOR_IN
	COMPARATOR_GT
	COMPARATOR_GTE
	COMPARATOR_LT
	COMPARATOR_LTE
)

type comparator struct {
	cType   comparatorType
	negated bool
}

func matchValue(column string, expectation interface{}, data map[string]interface{}) (bool, error) {

	_d("[matchValue]\n\tcolumn: %#v\n\texpectation: %#v\n\tdata: %#v\n", column, expectation, data)

	// -- direct detection based on column
	if string(column[0]) == "!" {
		_d("[matchValue] DIRECT: !\n")
		matched, err := matchValue(column[1:], expectation, data)
		if err != nil {
			return false, err
		}
		return !matched, nil
	}

	switch column {
	case "$and":
		_d("[matchValue] DIRECT: matcherAnd\n")
		return matcherAnd(expectation, data)
	case "$or":
		_d("[matchValue] DIRECT: matcherOr\n")
		return matcherOr(expectation, data)
	case "$not":
		_d("[matchValue] DIRECT: matcherNotAnd\n")
		matched, err := matcherAnd(expectation, data)
		if err != nil {
			return false, err
		}
		return !matched, nil
	}

	// direct comparator which is not matcher
	if string(column[0]) == "$" {
		_d("[matchValue] DIRECT: triggerComparator\n")
		comparator := detectComparator(column)
		return matchComparator(comparator, data, expectation)
	}

	// -- if still undetermined fall-back to expectation type based detection
	var cmp comparator

	switch interface{}(expectation).(type) {
	// -- expectation is logical scalar (not list/slice, nor map)
	// TODO: switch to less naive solution
	case string, int, float32, float64:
		cmp = comparator{cType: COMPARATOR_IS, negated: false}
	// -- expectation is a list -> wrap in "$in" comparator
	case []interface{}:
		cmp = comparator{cType: COMPARATOR_IN, negated: false}
	}

	// -- still undetermined -> pull comparator/expectation from expectation
	if cmp.cType == 0 {
		expectationAsMap, ok := expectation.(map[string]interface{})
		if !ok {
			_d("[matchValue] ERROR: NOT_A_MAP\n")
			return false, errors.New("matchValue: not a map")
		}

		if len(expectationAsMap) > 1 {
			// TODO: implement properly
			_d("[matchValue] ERROR: MULTIPLE_EXPECTATIONS\n")
			return false, errors.New("matchValue: multiple expectations")
		}
		for expKey, expValue := range expectationAsMap {
			cmp = detectComparator(expKey)
			expectation = expValue
			_d("[matchValue] unpacking comparator\n\tcomparator name: %#v,\n\tcomparator type: %#v,\n\texpectation: %#v\n", expKey, cmp, expectation)
			// we only allow one iteration (check on length should ensure this either way)
			break
		}
	}

	// -- obtain value
	valueInData, existsInData := fetchValue(data, column)
	_d("[matchValue]\n\tvalueInData: %#v\n\texistsInData: %#v\n", valueInData, existsInData)
	_d("[matchValue] comparator: %#v\n", cmp)

	// -- perform comparison
	return matchComparator(cmp, valueInData, expectation)
}

func detectComparator(comparatorName string) (cmp comparator) {
	_d("[detectComparator] comparatorName: %#v\n", comparatorName)
	// -- detect negation directly in comparator via "!"
	negate := false
	var comparatorNameFiltered string = comparatorName
	for i, chr := range comparatorName {
		if string(chr) == "!" {
			negate = !negate
			comparatorNameFiltered = comparatorName[i+1:]
			_d("[detectComparator] negation => negate: %t, comparatorNameFiltered: %#v\n", negate, comparatorNameFiltered)
		}
	}

	switch comparatorNameFiltered {
	case "$not":
		cmp = comparator{cType: COMPARATOR_NOT, negated: negate}
	case "$is":
		cmp = comparator{cType: COMPARATOR_IS, negated: negate}
	case "$in":
		cmp = comparator{cType: COMPARATOR_IN, negated: negate}
	case "$gt":
		cmp = comparator{cType: COMPARATOR_GT, negated: negate}
	case "$gte":
		cmp = comparator{cType: COMPARATOR_GTE, negated: negate}
	case "$lt":
		cmp = comparator{cType: COMPARATOR_LT, negated: negate}
	case "$lte":
		cmp = comparator{cType: COMPARATOR_LTE, negated: negate}
	}

	_d("[detectComparator] RETURN: %#v\n", cmp)
	return
}

func matchComparator(cmp comparator, valueInData, expectation interface{}) (out bool, err error) {
	_d("[matchComparator]\n\tcomparator: %#v\n\tvalueInData: %#v\n\texpectation: %#v\n", cmp, valueInData, expectation)

	var cmpResult bool

	switch cmp.cType {
	case COMPARATOR_IN:
		cmpResult, err = comparatorIn(valueInData, expectation)
	case COMPARATOR_IS:
		cmpResult = comparatorIs(valueInData, expectation)
	case COMPARATOR_NOT:
		cmpResult, err = comparatorNot(valueInData, expectation)
	case COMPARATOR_GT, COMPARATOR_GTE, COMPARATOR_LT, COMPARATOR_LTE:
		cmpResult, err = comparatorGeneric(cmp.cType, valueInData, expectation)
	default:
		// unknown comparator -> failure
		_d("[matchComparator] ERROR: UNKNOWN_COMPARATOR\n")
		cmpResult, err = false, errors.New("matchComparator: unknown comparator")
	}

	if err != nil {
		return false, err
	}

	out = cmpResult != cmp.negated
	_d("[matchComparator] RETURN: %t (negated: %t, cmpResult: %t)\n", out, cmp.negated, cmpResult)

	return
}

func fetchValue(data map[string]interface{}, column string) (result interface{}, found bool) {
	pathElements := strings.Split(column, COLUMN_LEVEL_SEPARATOR)

	var isMap bool
	_d("[fetchValue] enter\n\tpathElements: %#v\n\tdata: %#v\n\tcolumn: %#v\n", pathElements, data, column)

	for i, singlePathElement := range pathElements {
		dataNext, existsInData := data[singlePathElement]

		// case: no key on current level -> failure
		if !existsInData {
			_d("[fetchValue] RETURN: False (KEY_MISSING, iteration: %d)\n", i)
			break
		}

		// case: last iteration -> pass the result
		if i >= len(pathElements)-1 {
			_d("[fetchValue] RETURN: %#v\n", dataNext)
			result = dataNext
			found = true
			break
		}
		// case: is NOT a last iteration -> remap data
		data, isMap = dataNext.(map[string]interface{})

		// case: no value yet, and next level is not a map -> failure
		if !isMap {
			_d("[fetchValue] RETURN: False (NOT_A_MAP, iteration: %d)\n", i)
			break
		}
	}
	return
}
