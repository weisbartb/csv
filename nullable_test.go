package csv

import (
	"encoding/csv"
	"testing"

	testifyrequire "github.com/stretchr/testify/require"
)

func TestNullable(t *testing.T) {
	require := testifyrequire.New(t)
	var nullableBool NullableField[bool]
	nullableBool.Set(true)
	set, ok := nullableBool.Get()
	require.True(set)
	require.True(ok)
}

type simpleNullableCSVRecord struct {
	AString NullableField[string]  `csv:"a_string"`
	AFloat  NullableField[float64] `csv:"a_float"`
	AnInt   NullableField[int]     `csv:"an_int"`
	ABool   NullableField[bool]    `csv:"a_bool"`
}

func selectNullVal[T any](val T, _ bool) T {
	return val
}

func TestNullableCSV(t *testing.T) {
	require := testifyrequire.New(t)
	fh, err := testData.Open("testdata/simple.csv")
	require.NoError(err)
	csvReader := csv.NewReader(fh)
	reader, err := NewStructuredCSVReader[simpleNullableCSVRecord](csvReader)
	record, err := reader.Next()
	require.NoError(err)
	require.NotEmpty(record)
	require.Equal(true, selectNullVal(record.ABool.Get()))
	require.Equal(11, selectNullVal(record.AnInt.Get()))
	require.Equal(523.52, selectNullVal(record.AFloat.Get()))
	require.Equal("string", selectNullVal(record.AString.Get()))
}
