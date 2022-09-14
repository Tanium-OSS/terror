package terror_test

import (
	"fmt"
	"os"

	"github.com/Tanium-OSS/terror"
)

type Config struct{}

func ReadConfig(filename string) (Config, error) {
	_, err := os.ReadFile(filename)
	return Config{}, terror.WrapWithCode(err, 123, "loading config")
}

type System struct{}

func InitializeSystem(configPath string) (System, error) {
	_, err := ReadConfig(configPath)
	return System{}, terror.Wrap(err, "initializing system")
}

func Example() {
	_, err := InitializeSystem("oops where mah bucket?")
	fmt.Println(err)         // prints stack details
	fmt.Println(err.Error()) // prints short message only
	fmt.Printf("%s\n", err)  // prints short message only
	fmt.Printf("%+v\n", err) // prints stack details
	fmt.Printf("Error code: %d\n", terror.GetCode(err))
	// Output:
	// initializing system
	//  --- at terror/example_test.go:21 (InitializeSystem) ---
	// caused by loading config
	//  --- at terror/example_test.go:14 (ReadConfig) ---
	// caused by open oops where mah bucket?: no such file or directory
	// initializing system: loading config: open oops where mah bucket?: no such file or directory
	// initializing system: loading config: open oops where mah bucket?: no such file or directory
	// initializing system
	//  --- at terror/example_test.go:21 (InitializeSystem) ---
	// caused by loading config
	//  --- at terror/example_test.go:14 (ReadConfig) ---
	// caused by open oops where mah bucket?: no such file or directory
	// Error code: 123
}
