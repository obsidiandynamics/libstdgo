/*
Package arity contains helpers for working with variadic functions.
*/
package arity

import (
	"fmt"
	"reflect"
)

// Listify converts a potential array or slice into a slice of empty interfaces, returning an error if 'array'
// is not an array/slice.
func Listify(array interface{}) ([]interface{}, error) {
	kind := reflect.TypeOf(array).Kind()

	if kind != reflect.Array && kind != reflect.Slice {
		return nil, fmt.Errorf("Unsupported type %T", array)
	}

	value := reflect.ValueOf(array)
	length := value.Len()
	copy := make([]interface{}, length)
	for i := 0; i < length; i++ {
		copy[i] = value.Index(i).Interface()
	}
	return copy, nil
}

// ListifyOrPanic functions like Listify, but will panic instead of returning an error.
func ListifyOrPanic(array interface{}) []interface{} {
	a, err := Listify(array)
	if err != nil {
		panic(err)
	}
	return a
}

// SoleUntyped is a variation of Sole that is liberal in the type of argument array that it accepts,
// using ListifyOrPanic to convert any array/slice.
func SoleUntyped(def interface{}, args interface{}) interface{} {
	return Sole(def, ListifyOrPanic(args)...)
}

// Sole extracts the optional sole argument from args if len(args) == 0, returning the specified default
// if args is empty. If args contains multiple elements, the function panics.
func Sole(def interface{}, args ...interface{}) interface{} {
	return Optional(0, 1, def, args...)
}

// OptionalUntyped is a variation of Optional that is liberal in the type of argument array that it accepts,
// using ListifyOrPanic to convert any array/slice.
func OptionalUntyped(offset int, limit int, def interface{}, args interface{}) interface{} {
	return Optional(offset, limit, def, ListifyOrPanic(args)...)
}

const (
	errLimitLessThanOne   = "Limit must be greater than 0"
	errOffsetOutsideRange = "The limit-offset relationship must satisfy 0 <= offset < limit"
)

// Optional extracts the argument from args at the given offset if len(args) > offset, returning the specified default
// if args is not long enough. If args contains more elements than limit, the function panics. Useful
// for implementing optional arguments of variable arity, while enforcing their length.
func Optional(offset int, limit int, def interface{}, args ...interface{}) interface{} {
	switch {
	case limit < 1:
		panic(fmt.Errorf(errLimitLessThanOne))
	case offset < 0 || offset >= limit:
		panic(fmt.Errorf(errOffsetOutsideRange))
	}

	length := len(args)
	switch {
	case length > limit:
		panic(fmt.Errorf("Expected at most %d argument(s), got %d", limit, length))
	case offset < length:
		return args[offset]
	default:
		return def
	}
}
