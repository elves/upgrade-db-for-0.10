package eval

import (
	"errors"
	"os"
	"strings"
	"sync"
)

// Errors
var (
	ErrCanOnlyAssignList          = errors.New("can only assign compatible values")
	ErrPathMustBeString           = errors.New("path must be string")
	ErrPathCannotContainColonZero = errors.New(`path cannot contain colon or \0`)
)

// EnvPathList is a variable whose value is constructed from an environment
// variable by splitting at colons. Changes to it are also propagated to the
// corresponding environment variable. Its elements cannot contain colons or
// \0; attempting to put colon or \0 in its elements will result in an error.
//
// EnvPathList implements both Value and Variable interfaces. It also satisfied
// ListLike.
type EnvPathList struct {
	sync.RWMutex
	envName     string
	cachedValue string
	cachedPaths []string
}

var (
	_ Variable = (*EnvPathList)(nil)
	_ Value    = (*EnvPathList)(nil)
	_ ListLike = (*EnvPathList)(nil)
)

// Get returns a Value for an EnvPathList.
func (epl *EnvPathList) Get() Value {
	return epl
}

// Set sets an EnvPathList. The underlying environment variable is set.
func (epl *EnvPathList) Set(v Value) {
	iterator, ok := v.(Iterable)
	if !ok {
		throw(ErrCanOnlyAssignList)
	}
	var paths []string
	iterator.Iterate(func(v Value) bool {
		s, ok := v.(String)
		if !ok {
			throw(ErrPathMustBeString)
		}
		path := string(s)
		if strings.ContainsAny(path, ":\x00") {
			throw(ErrPathCannotContainColonZero)
		}
		paths = append(paths, string(s))
		return true
	})
	epl.set(paths)
}

// Kind returns "list".
func (epl *EnvPathList) Kind() string {
	return "list"
}

func (epl *EnvPathList) Eq(a interface{}) bool {
	return epl == a || eqListLike(epl, a)
}

// Repr returns the representation of an EnvPathList, as if it were an ordinary
// list.
func (epl *EnvPathList) Repr(indent int) string {
	var b ListReprBuilder
	b.Indent = indent
	for _, path := range epl.get() {
		b.WriteElem(quote(path))
	}
	return b.String()
}

// Len returns the length of an EnvPathList.
func (epl *EnvPathList) Len() int {
	return len(epl.get())
}

// Iterate iterates an EnvPathList.
func (epl *EnvPathList) Iterate(f func(Value) bool) {
	for _, p := range epl.get() {
		if !f(String(p)) {
			break
		}
	}
}

// IndexOne returns the result of one indexing operation.
func (epl *EnvPathList) IndexOne(idx Value) Value {
	paths := epl.get()
	slice, i, j := ParseAndFixListIndex(ToString(idx), len(paths))
	if slice {
		sliced := paths[i:j]
		values := make([]Value, len(sliced))
		for i, p := range sliced {
			values[i] = String(p)
		}
		return NewList(values...)
	}
	return String(paths[i])
}

// IndexSet sets one value in an EnvPathList.
func (epl *EnvPathList) IndexSet(idx, v Value) {
	s, ok := v.(String)
	if !ok {
		throw(ErrPathMustBeString)
	}

	paths := epl.get()
	slice, i, _ := ParseAndFixListIndex(ToString(idx), len(paths))
	if slice {
		throw(errors.New("slice set unimplemented"))
	}

	epl.Lock()
	defer epl.Unlock()
	paths[i] = string(s)
	epl.syncFromPaths()
}

func (epl *EnvPathList) get() []string {
	epl.Lock()
	defer epl.Unlock()

	value := os.Getenv(epl.envName)
	if value == epl.cachedValue {
		return epl.cachedPaths
	}
	epl.cachedValue = value
	epl.cachedPaths = strings.Split(value, ":")
	return epl.cachedPaths
}

func (epl *EnvPathList) set(paths []string) {
	epl.Lock()
	defer epl.Unlock()

	epl.cachedPaths = paths
	epl.syncFromPaths()
}

func (epl *EnvPathList) syncFromPaths() {
	epl.cachedValue = strings.Join(epl.cachedPaths, ":")
	err := os.Setenv(epl.envName, epl.cachedValue)
	maybeThrow(err)
}
