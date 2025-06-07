package address

import (
	"database/sql/driver"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"reflect"
	"strconv"
	"token-transfer-api/internal/errors/egeneric"
)

const AddressLength = common.AddressLength

type Address common.Address

func (a *Address) Scan(src any) error {
	if src == nil {
		return egeneric.NilError{Name: "src"}
	}

	v, ok := src.([]byte)
	if !ok {
		return egeneric.TypeError{ExpectedTypes: []reflect.Type{reflect.TypeOf([]byte{})}, ActualType: reflect.TypeOf(src)}
	}

	if len(v) != common.AddressLength {
		return egeneric.LengthError{ExpectedLength: common.AddressLength, ActualLength: len(v)}
	}

	copy(a[:], v)
	return nil
}

func (a *Address) Value() (driver.Value, error) {
	return a[:], nil
}

func (Address) DataType() string {
	return "BYTEA"
}

func (a Address) String() string {
	tempAddress := common.Address(a)
	return tempAddress.String()
}

func (a Address) Hex() string {
	return common.Address(a).Hex()
}

func HexToAddress(s string) *Address {
	tempAddress := Address(common.HexToAddress(s))
	return &tempAddress
}

// MarshalGQL implements the graphql.Marshaler interface for the Address type.
// It converts the Address to its hexadecimal string representation for GraphQL output.
func (a Address) MarshalGQL(w io.Writer) {
	quotedString := strconv.Quote(a.Hex())
	_, err := io.WriteString(w, quotedString)
	if err != nil {
		panic(fmt.Errorf("failed to marshal Address to GraphQL: %w", err))
	}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface for the Address type.
// It converts an input GraphQL string to an Address type.
func (a *Address) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return egeneric.TypeError{ExpectedTypes: []reflect.Type{reflect.TypeOf("")}, ActualType: reflect.TypeOf(v)}
	}

	if !common.IsHexAddress(s) {
		return fmt.Errorf("invalid Ethereum address format: %s", s)
	}

	tempAddress := common.HexToAddress(s)
	*a = Address(tempAddress)
	return nil
}
