package decimal

import (
	"fmt"
	dec "github.com/shopspring/decimal"
	"io"
	"reflect"
	"strconv"
	errs "token-transfer-api/internal/errors"
)

type Decimal dec.Decimal

func NewFromString(s string) (Decimal, error) {
	d, err := dec.NewFromString(s)
	if err != nil {
		return Decimal{}, err
	}

	return Decimal(d), nil
}

func NewFromInt64(n int64) Decimal {
	d := dec.NewFromInt(n)
	res := (Decimal)(d)
	return res
}

func NewFromFloat64(f float64) Decimal {
	d := dec.NewFromFloat(f)
	res := (Decimal)(d)
	return res
}

// MarshalGQL implements the graphql.Marshaler interface (used by gqlgen).
// It writes the string representation of the Decimal to the GraphQL response.
func (d Decimal) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(strconv.Quote((dec.Decimal(d)).String())))
	if err != nil {
		panic(fmt.Errorf("failed to marshal Decimal to GraphQL: %w\", ", err))
	}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface (used by gqlgen).
// It parses the GraphQL input value (string, int64, float64) into a Decimal.
func (d *Decimal) UnmarshalGQL(v interface{}) error {
	var val Decimal
	var err error

	switch v := v.(type) {
	case string:
		val, err = NewFromString(v)
	case int64:
		val = NewFromInt64(v)
	case float64:
		val = NewFromFloat64(v)
	default:
		return errs.TypeError{
			ExpectedTypes: []reflect.Type{
				reflect.TypeOf(""),
				reflect.TypeOf(int64(0)),
				reflect.TypeOf(float64(0))},
			ActualType: reflect.TypeOf(v),
		}
	}
	if err != nil {
		return err
	}

	*d = val
	return nil
}

// Value implements the database/sql.Valuer interface for database storage.
// It delegates to the underlying shopspring/decimal.Decimal's Value method.
func (d Decimal) Value() (interface{}, error) {
	return dec.Decimal(d).Value()
}

// Scan implements the database/sql.Scanner interface for database retrieval.
// It delegates to the underlying shopspring/decimal.Decimal's Scan method.
func (d *Decimal) Scan(value interface{}) error {
	temp := new(dec.Decimal)
	err := temp.Scan(value)
	if err != nil {
		return err
	}

	*d = Decimal(*temp)
	return nil
}
