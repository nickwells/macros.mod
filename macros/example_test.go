package macros_test

import (
	"fmt"

	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/macros.mod/macros"
)

// Example_withDirs demonstrates how the macros package might be used with
// macros directories
func Example_withDirs() {
	dirs := []string{
		"testdata/macros1",
		"testdata/macros2",
	}
	m, err := macros.New(macros.Dirs(dirs...), macros.Suffix(".xxx"))
	if err != nil {
		fmt.Printf("Unexpected error creating a new macro cache")
		return
	}

	strs := []string{
		"${f1}",
		"${f2}",
		"${XXX}",
	}
	loc := location.New("strSlice")
	for _, str := range strs {
		loc.Incr()
		newStr, err := m.Substitute(str, loc)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		fmt.Println(newStr)
	}
	// Output:
	// The contents of f1
	// The contents of f2.xxx
	// Error: Macro 'XXX' at strSlice:3 was not found in any of the macro directories: testdata/macros1, testdata/macros2
}

// Example_withoutDirs demonstrates how the macros package might be used
// without any macros directories
func Example_withoutDirs() {
	m, err := macros.New()
	if err != nil {
		fmt.Printf("Unexpected error creating a new macro cache")
		return
	}
	m.AddMacro("f1", "Replaced")
	m.AddMacro("f2", "Changed")
	m.AddMacro("f3", "Substituted")

	strs := []string{
		"Here is the macro to be ${f1}",
		"Whoops - no such macro ${XXX}",
		"${f2} or ${f3}",
		"${f2} and ${f2} again",
		"Whoops: Bad syntax ${f1",
	}
	loc := location.New("strSlice")
	for _, str := range strs {
		loc.Incr()
		newStr, err := m.Substitute(str, loc)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		fmt.Println(newStr)
	}
	// Output:
	// Here is the macro to be Replaced
	// Error: Macro 'XXX' at strSlice:2 was not found
	// Changed or Substituted
	// Changed and Changed again
	// Error: Bad macro at strSlice:5: a macro was started with '${' but not finished with '}'
}
