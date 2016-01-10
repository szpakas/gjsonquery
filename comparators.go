package gjsonquery

import (
	"errors"
	"fmt"
	"reflect"
)

func comparatorIs(actual, expected interface{}) bool {
	_d("[comparatorIs]\n\tactual: %#v \n\texpected: %#v\n", actual, expected)
	if actual != expected {
		_d("[comparatorIs] RETURN: False\n")
		return false
	}
	_d("[comparatorIs] RETURN: True\n")
	return true
}

func comparatorIn(actual, expected interface{}) (bool, error) {
	eCasted, ok := expected.([]interface{})

	_d("[comparatorIn]\n\tactual: %#v (%t|%#v)\n\texpected: %#v\n", actual, ok, eCasted, expected)

	if !ok {
		_d("[comparatorIn] ERROR: unknown expected type\n")
		return false, errors.New("comparatorIn: unknown expected type")
	}

	for _, e := range eCasted {
		if e == actual {
			_d("[comparatorIn] RETURN: True\n")
			return true, nil
		}
	}
	_d("[comparatorIn] RETURN: False\n")
	return false, nil
}

func comparatorNot(actual, expected interface{}) (matched bool, err error) {
	_d("[comparatorNot]\n\tactual: %#v \n\texpected: %#v\n", actual, expected)
	// -- determine actual comparator
	var cmp comparator
	switch interface{}(expected).(type) {
	case []interface{}:
		cmp = comparator{cType: COMPARATOR_IN, negated: false}
	default:
		cmp = comparator{cType: COMPARATOR_IS, negated: false}
	}

	_d("[comparatorNot] triggered comparator: %#v\n", cmp)
	matched, err = matchComparator(cmp, actual, expected)
	if err != nil {
		return false, err
	}
	return !matched, nil
}

func comparatorGeneric(cType comparatorType, actual, expected interface{}) (matched bool, err error) {
	_d("[comparator]\n\tcType: %#v\n\tactual: %#v \n\texpected: %#v\n", cType, actual, expected)
	isInt, aI, eI, aF, eF, err := castArguments(actual, expected)
	if err != nil {
		return
	}

	switch {
	case cType == COMPARATOR_GT && isInt:
		matched = aI > eI
	case cType == COMPARATOR_GT && !isInt:
		matched = aF > eF
	case cType == COMPARATOR_GTE && isInt:
		matched = aI >= eI
	case cType == COMPARATOR_GTE && !isInt:
		matched = aF >= eF
	case cType == COMPARATOR_LT && isInt:
		matched = aI < eI
	case cType == COMPARATOR_LT && !isInt:
		matched = aF < eF
	case cType == COMPARATOR_LTE && isInt:
		matched = aI <= eI
	case cType == COMPARATOR_LTE && !isInt:
		matched = aF <= eF
	default:
		return false, errors.New("comparator: unknown comparator type")
	}
	return
}

func castArguments(actual, expected interface{}) (isInt bool, aI, eI int, aF, eF float64, err error) {
	// We are using the "expected" type as required type and casting actual to it if necessary
	switch expectedCasted := expected.(type) {
	case int:
		actualCastedToInt, ok := actual.(int)
		if !ok {
			actualCastedToFloat64, ok := actual.(float64)
			if !ok {
				err = errors.New("comparator: casting actual to Int failed.")
				return
			}
			actualCastedToInt = int(actualCastedToFloat64)
		}
		// prepare output
		isInt = true
		aI = actualCastedToInt
		eI = expectedCasted
		return
	case float64:
		actualCastedToFloat64, ok := actual.(float64)
		if !ok {
			actualCastedToInt, ok := actual.(int)
			if !ok {
				err = errors.New("comparator: casting actual to Float64 failed.")
				return
			}
			actualCastedToFloat64 = float64(actualCastedToInt)
		}
		// prepare output
		isInt = false
		aF = actualCastedToFloat64
		eF = expectedCasted
		return
	}

	err = errors.New(fmt.Sprintf("comparator: unknown type (type: %s)", reflect.TypeOf(expected)))
	return
}
