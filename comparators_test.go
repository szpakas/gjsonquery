package gjsonquery_test

import (
	"testing"

	. "github.com/szpakas/gjsonquery"
)

func TestComparatorGeneric(t *testing.T) {
	_, err := ComparatorGeneric(-999, 123, 456)
	if err == nil && err.Error() != "comparator: unknown comparator type" {
		t.Error("Generic comparator should bail on unknown comparator type.")
	}
}
