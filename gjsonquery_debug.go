package gjsonquery

import "fmt"

var DEBUG = false

func _d(format string, a ...interface{}) {
	if DEBUG {
		fmt.Printf(format, a...)
	}
}
