package decimal

import (
	"fmt"
	dec "github.com/shopspring/decimal"
	"io"
	"reflect"
	"strconv"
	"token-transfer-api/internal/errors/egeneric"
)

type Decimal dec.Decimal

var Zero = Decimal(dec.Zero)

const (
	NumericPrecision = 78
	NumericScale     = 0
)

func (Decimal) GormDataType() string {
	return fmt.Sprintf("numeric(%d,%d)", NumericPrecision, NumericScale)
}

func (d Decimal) String() string {
	return (dec.Decimal(d)).String()
}

// Int64 returns the decimal as int64.
// If the value cannot be represented in an int64 the result is undefined.
func (d Decimal) Int64() int64 {
	return (dec.Decimal(d)).BigInt().Int64()
}

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

func (d Decimal) IsInteger() bool {
	return (dec.Decimal(d)).IsInteger()
}

func (d Decimal) Cmp(o Decimal) int {
	return dec.Decimal(d).Cmp(dec.Decimal(o))
}

func (d Decimal) Equal(o Decimal) bool {
	return d.Cmp(o) == 0
}

func (d Decimal) GreaterThan(o Decimal) bool {
	return d.Cmp(o) > 0
}

func (d Decimal) GreaterThanOrEqual(o Decimal) bool {
	return d.Cmp(o) >= 0
}

func (d Decimal) LessThan(o Decimal) bool {
	return d.Cmp(o) < 0
}

func (d Decimal) LessThanOrEqual(o Decimal) bool {
	return d.Cmp(o) <= 0
}

func (d Decimal) IsZero() bool {
	return d.Cmp(Zero) == 0
}

func (d Decimal) Add(o Decimal) Decimal {
	return Decimal((dec.Decimal(d)).Add(dec.Decimal(o)))
}

func (d Decimal) Sub(o Decimal) Decimal {
	return Decimal((dec.Decimal(d)).Sub(dec.Decimal(o)))
}

// MarshalGQL implements the graphql.Marshaler interface (used by gqlgen).
// It writes the string representation of the Decimal to the GraphQL response.
func (d Decimal) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(strconv.Quote((dec.Decimal(d)).String())))
	if err != nil {
		panic(fmt.Errorf("failed to marshal Decimal to GraphQL: %w", err))
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
		return egeneric.TypeError{
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
