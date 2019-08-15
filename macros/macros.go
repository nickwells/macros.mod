package macros

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/nickwells/check.mod/check"
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

// M records the information needed to substitute macros
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
type M struct {
	mMap     map[string]string
	mDirs    []string
	suffixes []string
	mStart   string
	mEnd     string
}

type OptFunc func(m *M) error

// New creates a new M object.
func New(opts ...OptFunc) (*M, error) {
	m := &M{
		mMap:     make(map[string]string),
		mDirs:    make([]string, 0),
		suffixes: []string{""},
		mStart:   DfltMStart,
		mEnd:     DfltMEnd,
	}

	for _, o := range opts {
		if err := o(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Dirs returns an OptFunc that will add the directory names to the,
// initially empty, set of directories to be searched. Each of the passed
// values must be a directory, an error will be returned if not and none of
// the passed values will be added.
func Dirs(dirs ...string) OptFunc {
	return func(m *M) error {
		if len(dirs) == 0 {
			return fmt.Errorf("at least one macros directory must be passed")
		}

		es := filecheck.Provisos{
			Checks:    []check.FileInfo{check.FileInfoIsDir},
			Existence: filecheck.MustExist,
		}
		for _, dir := range dirs {
			err := es.StatusCheck(dir)
			if err != nil {
				return err
			}
		}

		m.mDirs = append(m.mDirs, dirs...)
		return nil
	}
}

// Suffix returns an OptFunc that will add a suffix to the list of strings to
// be tried as suffixes Any suffix must be complete and include the separator
// (if any). For instance ".sql". The suffixes are tried in the order they
// are added and there is always a first, empty suffix so that a macro name
// will always match a file with the exact same name.
func Suffix(suffix string) OptFunc {
	return func(m *M) error {
		m.suffixes = append(m.suffixes, suffix)

		return nil
	}
}

// StartEndStr returns an OptFunc that will change the strings that are used
// to bracket a macro in the line. The values given will be used in
// Substitute to find the macro. The default values are given by DfltMStart
// and DfltMEnd
func StartEndStr(start, end string) OptFunc {
	return func(m *M) error {
		m.mStart = start
		m.mEnd = end

		return nil
	}
}

// AddMacro will add a named macro to the macro map which can
// subsequently be used to substitute into a string
func (m *M) AddMacro(name, value string) {
	m.mMap[name] = value
}

// Find searches for the macro name in the cache. If it is not found and
// there are macro directories to be searched then it will search for a
// matching file name and returns the contents if it finds it. If no matching
// macro is found an error is returned
func (m *M) Find(mName string, loc *location.L) (string, error) {
	if macro, ok := m.mMap[mName]; ok {
		return macro, nil
	}

	for _, fd := range m.mDirs {
		for _, suffix := range m.suffixes {
			macro, err := ioutil.ReadFile(filepath.Join(fd, mName+suffix))
			if err == nil {
				m.mMap[mName] = string(macro)
				return m.mMap[mName], nil
			}
		}
	}

	errStr := fmt.Sprintf("Macro '%s' at %s was not found", mName, loc)
	if len(m.mDirs) == 1 {
		errStr += " in the macro directory: " + m.mDirs[0]
	} else if len(m.mDirs) > 1 {
		errStr += " in any of the macro directories: " +
			strings.Join(m.mDirs, ", ")
	}

	return "", errors.New(errStr)
}

// Substitute searches the line for macros and replaces them with the
// corresponding text. If a macro is not well-formed (is not terminated
// properly) or cannot be found in the cache or any of the macro directories
// then an error is returned. A macro is a string between the macro start and
// end strings (see DfltMStart and DfltMEnd). Macros do not nest. There can
// be any number of macros on a line
func (m *M) Substitute(line string, loc *location.L) (string, error) {
	parts := strings.SplitN(line, m.mStart, 2)
	line = parts[0]
	for len(parts) == 2 {
		parts = strings.SplitN(parts[1], m.mEnd, 2)

		if len(parts) != 2 {
			err :=
				fmt.Errorf("Bad macro at %s:"+
					" a macro was started with '%s'"+
					" but not finished with '%s'",
					loc, m.mStart, m.mEnd)
			return "", err
		}
		macro, err := m.Find(parts[0], loc)
		if err != nil {
			return "", err
		}
		line += macro

		parts = strings.SplitN(parts[1], m.mStart, 2)
		line += parts[0]
	}
	return line, nil
}
