package csv

import (
	"encoding/csv"
	"io"
	"reflect"

	"github.com/weisbartb/rcache"
	"github.com/weisbartb/stack"
)

// Writer holds the state of the CSV writer
type Writer[Record any] struct {
	headerWritten bool
	instruction   *rcache.FieldCache[csvInstruction]
	w             *csv.Writer
}

// NewWriter makes a new CSV writer
func NewWriter[Record any](writer io.Writer) *Writer[Record] {
	var T Record
	return &Writer[Record]{
		w:           csv.NewWriter(writer),
		instruction: fieldCache.GetTypeDataFor(reflect.TypeOf(T)),
	}
}

// WriteRecord writes record(s) to the underlying file, a flush is automatically called upon finishing.
func (c *Writer[Record]) WriteRecord(items ...Record) error {
	defer func() {
		// Flush the buffered IO from the underlying csv-writer
		c.w.Flush()
	}()
	if !c.headerWritten {
		if err := c.writeHeader(); err != nil {
			return stack.Trace(err)
		}
	}
	for _, item := range items {
		vOf := reflect.ValueOf(item)
		var row []string
		for _, field := range fieldCache.GetTypeDataFor(vOf.Type()).Fields() {
			val, err := field.InstructionData().encoder(vOf.Field(field.Idx))
			if err != nil {
				return stack.Trace(err)
			}
			row = append(row, val)
		}
		if err := c.w.Write(row); err != nil {
			return stack.Trace(err)
		}
	}
	return nil
}

// writeHeader is a helper method to write out the header to the CSV
func (c *Writer[Record]) writeHeader() error {
	var columns []string
	var rec Record
	for _, field := range fieldCache.GetTypeDataFor(reflect.TypeOf(rec)).Fields() {
		columns = append(columns, field.InstructionData().GetCSVHeaderIdentifier())
	}
	if err := c.w.Write(columns); err != nil {
		return stack.Trace(err)
	}
	c.headerWritten = true
	return nil
}
