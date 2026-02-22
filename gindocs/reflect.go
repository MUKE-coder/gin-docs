package gindocs

import (
	"reflect"
	"runtime"
	"strings"
)

// handlerFuncName extracts a clean function name from a handler's full path.
func handlerFuncName(handlerName string) string {
	// Handler names look like: "main.createUser" or "github.com/foo/bar.Handler.func1"
	parts := strings.Split(handlerName, "/")
	last := parts[len(parts)-1]

	// Split by dot to get the function name.
	dotParts := strings.Split(last, ".")
	if len(dotParts) > 1 {
		return dotParts[len(dotParts)-1]
	}

	return last
}

// handlerPackageName extracts the package name from a handler's full path.
func handlerPackageName(handlerName string) string {
	parts := strings.Split(handlerName, "/")
	last := parts[len(parts)-1]

	dotParts := strings.Split(last, ".")
	if len(dotParts) > 1 {
		return dotParts[0]
	}

	return last
}

// getFuncName returns the name of a function using reflection.
func getFuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
