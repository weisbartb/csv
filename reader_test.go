package csv

import (
	"embed"
	"encoding/csv"
	"testing"

	testifyrequire "github.com/stretchr/testify/require"
)

//go:embed testdata
var testData embed.FS

type simpleCSVRecord struct {
	AString string  `csv:"a_string"`
	AFloat  float64 `csv:"a_float"`
	AnInt   int     `csv:"an_int"`
	ABool   bool    `csv:"a_bool"`
}
type simpleCSVRecordStrictFail struct {
	AString string  `csv:"a_string"`
	AFloat  float64 `csv:"a_float"`
	AnInt   int     `csv:"an_int"`
}
type requiredCSVRecordStrictFail struct {
	AString string  `csv:"a_string,required"`
	AFloat  float64 `csv:"a_float,required"`
	AnInt   int     `csv:"an_int,required"`
	ABool   bool    `csv:"a_bool"`
}

func TestNewStructuredCSVReader(t *testing.T) {
	require := testifyrequire.New(t)

	fh, err := testData.Open("testdata/simple.csv")
	require.NoError(err)
	csvReader := csv.NewReader(fh)
	reader, err := NewStructuredCSVReader[simpleCSVRecord](csvReader)
	require.NoError(err)
	require.NotNil(reader)
}

func TestReader_Next(t *testing.T) {
	t.Run("base test", func(t *testing.T) {
		require := testifyrequire.New(t)
		fh, err := testData.Open("testdata/simple.csv")
		require.NoError(err)
		csvReader := csv.NewReader(fh)
		reader, err := NewStructuredCSVReader[simpleCSVRecord](csvReader)
		require.NoError(err)
		record, err := reader.Next()
		require.NoError(err)
		require.NotEmpty(record)
		require.Equal(true, record.ABool)
		require.Equal(11, record.AnInt)
		require.Equal(523.52, record.AFloat)
		require.Equal("string", record.AString)
	})
	t.Run("strict mode", func(t *testing.T) {
		require := testifyrequire.New(t)
		fh, err := testData.Open("testdata/simple.csv")
		require.NoError(err)
		csvReader := csv.NewReader(fh)
		reader, err := NewStructuredCSVReader[simpleCSVRecordStrictFail](csvReader)
		reader.StrictMode = true
		require.NoError(err)
		_, err = reader.Next()
		require.EqualError(err, "a_bool was seen in the csv but not in the record provided")
	})
	t.Run("required mode - ok", func(t *testing.T) {
		require := testifyrequire.New(t)
		fh, err := testData.Open("testdata/simple-required.csv")
		require.NoError(err)
		csvReader := csv.NewReader(fh)
		reader, err := NewStructuredCSVReader[requiredCSVRecordStrictFail](csvReader)
		require.NoError(err)
		record, err := reader.Next()
		require.NoError(err)
		require.Equal(false, record.ABool)
		require.Equal(11, record.AnInt)
		require.Equal(523.52, record.AFloat)
		require.Equal("string", record.AString)
	})
	t.Run("required mode - wrong", func(t *testing.T) {
		require := testifyrequire.New(t)
		fh, err := testData.Open("testdata/simple-required.csv")
		require.NoError(err)
		csvReader := csv.NewReader(fh)
		reader, err := NewStructuredCSVReader[requiredCSVRecordStrictFail](csvReader)
		require.NoError(err)
		_, _ = reader.Next()
		_, err = reader.Next()
		require.EqualError(err, "an_int is a required field")
	})

}
