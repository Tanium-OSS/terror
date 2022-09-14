// Package terror is an error-handling library.
//
// Errors wrapped by this package include the code location that the error was
// wrapped at. They print detailed stack information when formatted via the "%v"
// printf directive. Otherwise, they print only the message prefix.
//
// That is, given:
//
//	err := terror.New("some error")
//	err = terror.Wrap(err, "loading config")
//	err = terror.Wrap(err, "initializing %s", "system")
//
// Then
//
//	fmt.Println(err)
//	fmt.Println(err.Error())
//	fmt.Printf("%s", err)
//
// will each print
//
//	initializing system: loading config: some error
//
// But
//
//	fmt.Printf("%v", err)
//	fmt.Printf("%+v", err)
//	fmt.Printf("%#v", err)
//
// will print
//
//	initializing system
//	 --- at path/to/my/pkg/system.go:123 (system.Initialize) ---
//	caused by loading config
//	 --- at path/to/my/pkg/config.go:37 (system.LoadConfig) ---
//	caused by some error
//	 --- at path/to/my/pkg/config.go:62 (system.ReadConfig) ---
//
// # Error formatting notes
//
// Most error libraries that include call site information have settled on only
// printing stack details when the error is formatted via "%+v". However, much
// existing code at Tanium only logs error using "%v". We therefore strongly
// encourage users to use "%+v" when logging errors and wanting to show detailed
// stack information, but we will support "%v" until a satisfactory auditing
// mechanism can be achieved.
//
// Using fmt.Errorf to wrap an error commits to using the detailed format.
//
// This library is inspired by github.com/palantir/stacktrace.
package terror
