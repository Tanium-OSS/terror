package terror

import (
	"fmt"
)

func ExampleWrapInto() {
	checkComputer := func() error {
		return nil
	}

	getQuestionForAnswer := func(answer string) (question string, err error) {
		// Set up error handling for this function.
		defer WrapInto(&err, "getQuestionForAnswer(%q)", answer)

		err = checkComputer()
		if err != nil {
			// Can also call Wrap here, if there is additional context to provide for this one error.
			return "", err
		}

		// Can defer another WrapInto, if there is additional context to provide for all subsequent errors.

		if answer == "42" {
			return "What is the answer to the ultimate question of life, the universe, and everything?", nil
		}
		return "", New("insufficient cheese")
	}

	_, errExample := getQuestionForAnswer("five tons of flax")

	fmt.Println(errExample.Error())
	// Output: getQuestionForAnswer("five tons of flax"): insufficient cheese
}
