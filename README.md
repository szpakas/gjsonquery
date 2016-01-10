# gjsonquery

An implementation of the [JSON Query Language](https://github.com/clue/json-query-language) specification in Go.

The [JSON Query Language](https://github.com/clue/json-query-language) specification is currently in draft mode.
This library implements the [v0.4 draft](https://github.com/clue/json-query-language/releases/tag/v0.4.0).

> Note: This project is in beta stage! Feel free to report any issues you encounter.

Go implementation is a, somewhat naive, port of an original PHP implementation.

## Missing features
Comparator "$contains" is not implemented.

Multiple comparators/matchers at one level are not supported.

## Implementation detail

Comparators "$gt", "$gte", "$lt", "$lte" will try to perform int <-> float64 casting when necessary.
Float to int casting does not round the values.
Base type is always taken from expected value (from query).

Reflection is used only for reporting errors and in tests.

## Dependencies

There are no external dependencies outside the Go standard library.

## License

Apache 2.0
