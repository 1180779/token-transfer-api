package egeneric

import (
	"fmt"
	"reflect"
)

type LengthError struct {
	ExpectedLength int
	ActualLength   int
}

func (e LengthError) Error() string {
	return fmt.Sprintf("Expected length: %d, got: %d", e.ExpectedLength, e.ActualLength)
}

type TypeError struct {
	ExpectedTypes []reflect.Type
	ActualType    reflect.Type
}

func (e TypeError) Error() string {
	var msg = "Expected one of the following types: "
	for _, t := range e.ExpectedTypes {
		msg += "'" + t.String() + "'; "
	}
	msg += "got: " + e.ActualType.String()
	return msg
}

type NilError struct {
	Name string
}

func (e NilError) Error() string {
	return fmt.Sprintf("%s cannot be nil", e.Name)
}
