package csv

import (
	"encoding/csv"
	"reflect"

	"github.com/weisbartb/rcache"
	"github.com/weisbartb/stack"
)

type Writer[Record any] struct {
	headerWritten bool
	instruction   *rcache.FieldCache[csvInstruction]
	w             *csv.Writer
}

func NewWriter[Record any](writer *csv.Writer) *Writer[Record] {
	var T Record
	return &Writer[Record]{
		w:           writer,
		instruction: fieldCache.GetTypeDataFor(reflect.TypeOf(T)),
	}
}

func (c *Writer[Record]) WriteRecord(items ...Record) error {
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
	// Flush the buffered IO from the underlying csv-writer
	c.w.Flush()
	return nil
}
func (c *Writer[Record]) writeHeader() error {
	var columns []string
	var rec Record
	for _, field := range fieldCache.GetTypeDataFor(reflect.TypeOf(rec)).Fields() {
		columns = append(columns, field.InstructionData().GetExportedFieldName())
	}
	if err := c.w.Write(columns); err != nil {
		return stack.Trace(err)
	}
	c.headerWritten = true
	return nil
}
