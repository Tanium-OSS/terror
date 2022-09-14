package terror

import (
	"fmt"
)

const errTestError = Const("test error")

func ExampleConst() {
	err := Wrap(errTestError, "in test %q", "ExampleConstError")
	fmt.Println(err.Error())
	// Output:
	// in test "ExampleConstError": test error
}
