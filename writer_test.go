package csv

import (
	"bytes"
	"encoding/csv"
	"testing"

	testifyrequire "github.com/stretchr/testify/require"
)

type testWriterStruct struct {
	Email      string  `csv:"email"`
	Age        int     `csv:"age"`
	Owed       float64 `csv:"owed"`
	ShouldBill bool    `cst:"should_bill"`
}

type testWriterOmitEmptyStruct struct {
	Email      string  `csv:"email,omitempty"`
	Age        int     `csv:"age,omitempty"`
	Owed       float64 `csv:"owed,omitempty"`
	ShouldBill bool    `cst:"should_bill,omitempty"`
}

func TestNewWriter(t *testing.T) {
	t.Run("Basic writer", func(t *testing.T) {
		require := testifyrequire.New(t)
		buf := bytes.Buffer{}
		csvWriter := csv.NewWriter(&buf)
		writer := NewWriter[testWriterStruct](csvWriter)
		err := writer.WriteRecord(testWriterStruct{
			Email:      "test@example.com",
			Age:        32,
			Owed:       6512.23,
			ShouldBill: true,
		})
		require.NoError(err)
		require.Equal("test@example.com,32,6512.23,TRUE\n", buf.String())
	})
	t.Run("Omit Empty", func(t *testing.T) {
		require := testifyrequire.New(t)
		buf := bytes.Buffer{}
		csvWriter := csv.NewWriter(&buf)
		writer := NewWriter[testWriterStruct](csvWriter)
		err := writer.WriteRecord(testWriterStruct{
			Email:      "test@example.com",
			Age:        32,
			Owed:       6512.23,
			ShouldBill: true,
		})
		require.NoError(err)
		require.Equal("test@example.com,32,6512.23,TRUE\n", buf.String())
	})
}
