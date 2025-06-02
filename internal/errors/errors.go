package errors

import (
	"fmt"
	"reflect"
)

type LengthError struct {
	ExpectedLength int
	ActualLength   int
}

func (e LengthError) Error() string {
	return fmt.Sprintf("Expected length: %d got: %d", e.ExpectedLength, e.ActualLength)
}

type TypeError struct {
	ExpectedType reflect.Type
	ActualType   reflect.Type
}

func (e TypeError) Error() string {
	return fmt.Sprintf("Expected type: %T, got: %T", e.ExpectedType, e.ActualType)
}

type NilError struct {
	Name string
}

func (e NilError) Error() string {
	return fmt.Sprintf("%s cannot be nil", e.Name)
}
