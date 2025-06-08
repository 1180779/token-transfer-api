package address

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"token-transfer-api/internal/errors/egeneric"
)

func testAddressValue(t *testing.T, address Address, expectedErrorType interface{}) {
	t.Helper()

	value, err := address.Value()
	if expectedErrorType != nil {
		assert.Error(t, err)
		assert.IsType(t, expectedErrorType, err, "Error should be of type: %s", reflect.TypeOf(expectedErrorType))
	} else {
		assert.NoError(t, err)

		s, ok := value.(string)
		assert.True(t, ok, "Value should be a string")
		assert.Equal(t, AddressHexLength, len(s), "Value should be the address hex length")
		assert.Equal(t, address.String(), s, "Decoded address should match original")
	}
}

func TestAddress_Value_ZeroAddress(t *testing.T) {
	testAddressValue(t, HexToAddress("0x0000000000000000000000000000000000000000"), nil)
}

func TestAddress_Value_NonZeroAddress(t *testing.T) {
	testAddressValue(t, HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), nil)
}

func testAddressScan(t *testing.T, source any, expectedAddress *Address, expectedErrorType interface{}) {
	t.Helper()

	var address Address
	err := address.Scan(source)

	if expectedErrorType != nil {
		assert.Error(t, err)
		assert.IsType(t, expectedErrorType, err, "Error should be of type: %s", reflect.TypeOf(expectedErrorType))
	} else {
		assert.NoError(t, err)
		assert.Equal(t, expectedAddress.String(), address.String(), "Scanned address should match expected")
	}
}

func TestAddress_Scan_NilSource(t *testing.T) {
	testAddressScan(t, nil, nil, egeneric.NilError{})
}

func TestAddress_Scan_InvalidDataSourceType(t *testing.T) {
	testAddressScan(t, []byte{}, nil, egeneric.TypeError{})
}

func TestAddress_Scan_EmptyString(t *testing.T) {
	testAddressScan(t, "", nil, egeneric.LengthError{})
}

func TestAddress_Scan_TooShort(t *testing.T) {
	src := "abdefg"
	testAddressScan(t, src, nil, egeneric.LengthError{})
}

func TestAddress_Scan_TooLong(t *testing.T) {
	src := "abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"
	testAddressScan(t, src, nil, egeneric.LengthError{})
}
