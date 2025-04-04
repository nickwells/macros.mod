package macros

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
)

// DfltMStart is the default start string for a macro
// DfltMEnd is the default end string for a macro
//
// They are used by Substitute to find macro names in the string to be
// substituted
const (
	DfltMStart = "${"
	DfltMEnd   = "}"
)

// Cache records the information needed to substitute macros
//
// You should create a new Macros object with New and then if you want to
// read any macros from files in macro directories then add the
// macro directories before substituting values.
//
// You can add any suffixes that should be tried when searching for a
// macro in the macro directories using the AddSuffix method.
//
// You can then use Find to get the text representing the macro or
// use Substitute to replace any macro names in the passed line
type Cache struct {
	mMap      map[string]string
	badMacros map[string]bool
	mDirs     []string
	suffixes  []string
	mStart    string
	mEnd      string
}

// OptFunc is the type of a function used to set options on a macro Cache
type OptFunc func(c *Cache) error

// NewCache creates a new Cache object.
func NewCache(opts ...OptFunc) (*Cache, error) {
	c := &Cache{
		mMap:      make(map[string]string),
		badMacros: make(map[string]bool),
		mDirs:     make([]string, 0),
		suffixes:  []string{""},
		mStart:    DfltMStart,
		mEnd:      DfltMEnd,
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Dirs returns an OptFunc that will add the directory names to the,
// initially empty, set of directories to be searched. Each of the passed
// values must be a directory, an error will be returned if not and none of
// the passed values will be added.
func Dirs(dirs ...string) OptFunc {
	return func(c *Cache) error {
		if len(dirs) == 0 {
			return fmt.Errorf("at least one macros directory must be passed")
		}

		es := filecheck.DirExists()
		for _, dir := range dirs {
			err := es.StatusCheck(dir)
			if err != nil {
				return err
			}
		}

		c.mDirs = append(c.mDirs, dirs...)

		return nil
	}
}

// Suffix returns an OptFunc that will add a suffix to the list of strings to
// be tried as suffixes Any suffix must be complete and include the separator
// (if any). For instance ".sql". The suffixes are tried in the order they
// are added and there is always a first, empty suffix so that a macro name
// will always match a file with the exact same name.
func Suffix(suffix string) OptFunc {
	return func(c *Cache) error {
		c.suffixes = append(c.suffixes, suffix)

		return nil
	}
}

// StartEndStr returns an OptFunc that will change the strings that are used
// to bracket a macro in the line. The values given will be used in
// Substitute to find the macro. The default values are given by DfltMStart
// and DfltMEnd
func StartEndStr(start, end string) OptFunc {
	return func(c *Cache) error {
		c.mStart = start
		c.mEnd = end

		return nil
	}
}

// AddMacro will add a named macro to the macro map which can
// subsequently be used to substitute into a string
func (c *Cache) AddMacro(name, value string) {
	c.mMap[name] = value
}

// Find searches for the macro name in the cache. If it is not found and
// there are macro directories to be searched then it will search for a
// matching file name and returns the contents if it finds it. If no matching
// macro is found an error is returned
func (c *Cache) Find(mName string, loc *location.L) (string, error) {
	if macro, ok := c.mMap[mName]; ok {
		return macro, nil
	}

	if c.badMacros[mName] {
		return "", fmt.Errorf("macro %q at %s was not found", mName, loc)
	}

	for _, fd := range c.mDirs {
		for _, suffix := range c.suffixes {
			macro, err := os.ReadFile( //nolint:gosec
				filepath.Join(fd, mName+suffix))
			if err == nil {
				c.mMap[mName] = string(macro)
				return c.mMap[mName], nil
			}
		}
	}

	c.badMacros[mName] = true

	errStr := fmt.Sprintf("macro %q at %s was not found", mName, loc)
	if len(c.mDirs) == 1 {
		errStr += " in the macro directory: " + c.mDirs[0]
	} else if len(c.mDirs) > 1 {
		errStr += " in any of the macro directories: " +
			strings.Join(c.mDirs, ", ")
	}

	return "", errors.New(errStr)
}

// Substitute searches the line for macros and replaces them with the
// corresponding text. If a macro is not well-formed (is not terminated
// properly) or cannot be found in the cache or any of the macro directories
// then an error is returned. A macro is a string between the macro start and
// end strings (see DfltMStart and DfltMEnd). Macros do not nest. There can
// be any number of macros on a line
func (c *Cache) Substitute(line string, loc *location.L) (string, error) {
	plainText, macroEtc, macroFound := strings.Cut(line, c.mStart)
	expandedLine := plainText

	for macroFound {
		macroName, remainder, macroTerminated := strings.Cut(macroEtc, c.mEnd)
		if !macroTerminated {
			err := fmt.Errorf(
				"%s: a macro was started with %q but not finished with %q",
				loc, c.mStart, c.mEnd)

			return "", err
		}

		macro, err := c.Find(macroName, loc)
		if err != nil {
			return "", err
		}

		expandedLine += macro
		plainText, macroEtc, macroFound = strings.Cut(remainder, c.mStart)
		expandedLine += plainText
	}

	return expandedLine, nil
}

// GetStartEndStrings returns the start and end strings (which bracket the
// macro name)
func (c Cache) GetStartEndStrings() (string, string) {
	return c.mStart, c.mEnd
}
