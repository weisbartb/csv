package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"

	"github.com/weisbartb/rcache"
	"github.com/weisbartb/stack"
)

// Reader holds the state of a CSV reader and binds it to a given record type.
// Record can be any struct
type Reader[Record any] struct {
	// StrictMode will error on any unhandled fields seen in the CSV
	StrictMode bool
	// reader holds the underlying CSV reader
	reader *csv.Reader
	// currentRow holds the current row for reporting problematic rows
	currentRow int
	// headerRead activates after the header gets parsed the first time
	headerRead bool
	// headerMap stores the position of all header keys
	headerMap map[string]int
	// headers contain a list of all header values.
	headers []string
	// instruction holds a cached copy of the record instruction
	instruction *rcache.FieldCache[csvInstruction]
}

func (r *Reader[Record]) readHeader() error {
	if r.headerRead {
		return nil
	}
	row, err := r.reader.Read()
	if err != nil {
		return stack.Wrap(err, "reading csv header")
	}
	r.headerMap = map[string]int{}
	for k, v := range row {
		r.headers = append(r.headers, v)
		r.headerMap[v] = k
	}
	r.headerRead = true
	r.currentRow++
	return nil
}

// initialize initializes the reader
func (r *Reader[Record]) initialize() error {
	var t Record
	err := r.readHeader()
	if err != nil {
		return stack.Trace(err)
	}
	tOf := reflect.TypeOf(t)
	// Warm up cache
	instructions := fieldCache.GetTypeDataFor(tOf)
	if r.StrictMode {
		// StrictMode will error for any field that can't be found in the struct
		for _, v := range r.headers {
			if instructions.GetFieldByName(v) == nil {
				return stack.Trace(fmt.Errorf("%v was seen in the csv but not in the record provided", v))
			}
		}
	}
	return nil
}

// nextRow is a helper method to get the next row and wrap errors with the row that failed
func (r *Reader[Record]) nextRow() (record []string, err error) {
	record, err = r.reader.Read()
	if err == nil {
		r.currentRow++
	} else {
		err = stack.Wrap(err, fmt.Sprintf("on row %v", r.currentRow))
	}
	return
}

// Next gets the next Record in the file.
// This will return a valid Record or an error.
// This can return io.EOF which is a valid control signal to stop the loop.
func (r *Reader[Record]) Next() (Record, error) {
	var out Record
	if !r.headerRead {
		if err := r.initialize(); err != nil {
			return out, stack.Trace(err)
		}
	}
	// Load the row
	row, err := r.nextRow()
	if err != nil {
		return out, stack.Trace(err)
	}
	vOf := reflect.ValueOf(&out)
	tData := vOf.Elem()

	for cellOffset, cell := range row {
		fieldData := r.instruction.GetFieldByName(r.headers[cellOffset])
		// fieldData is nil if the field is ignored or unrecognized.
		if fieldData == nil {
			continue
		}
		var isNull bool
		if len(cell) == 0 {
			// Set the isNull flag for the decoder
			isNull = true
		}
		val, err := fieldData.InstructionData().GetDecoder()(cell, isNull)
		if err != nil {
			return out, stack.Trace(err)
		}
		// Set the value on the field
		tData.Field(fieldData.Idx).Set(reflect.ValueOf(val))
	}
	return out, nil
}

// NewStructuredCSVReader sets up a new reader for a given file handle.
func NewStructuredCSVReader[Record any](fileHandle io.Reader) *Reader[Record] {
	var T Record
	wrapper := &Reader[Record]{
		reader:      csv.NewReader(fileHandle),
		instruction: fieldCache.GetTypeDataFor(reflect.TypeOf(T)),
	}
	return wrapper
}
