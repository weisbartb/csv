package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"

	"github.com/weisbartb/stack"
)

type HeaderKey string
type Field struct {
	id     int
	ignore bool
	t      reflect.Kind
}
type Reader[Record any] struct {
	// StrictMode will error on any unhandled fields seen in the CSV
	StrictMode bool
	reader     *csv.Reader
	currentRow int
	headerRead bool
	headerMap  map[string]int
	headers    []string
}

var ErrMissingRecord = errors.New("a valid record type is required")

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

func (r *Reader[Record]) initialize(emptyRecord any) error {
	err := r.readHeader()
	if err != nil {
		return stack.Trace(err)
	}
	if emptyRecord == nil {
		return stack.Trace(ErrMissingRecord)
	}
	tOf := reflect.TypeOf(emptyRecord)
	// Warm up cache
	instructions := fieldCache.GetTypeDataFor(tOf)
	if r.StrictMode {
		// See if any fields that
		for _, v := range r.headers {
			if instructions.GetFieldByName(v) == nil {
				return stack.Trace(fmt.Errorf("%v was seen in the csv but not in the record provided", v))
			}
		}
	}
	return nil
}

func (r *Reader[Record]) nextRow() (record []string, err error) {
	record, err = r.reader.Read()
	if err == nil {
		r.currentRow++
	} else {
		err = stack.Wrap(err, fmt.Sprintf("on row %v", r.currentRow))
	}
	return
}

func (r *Reader[Record]) Next() (Record, error) {
	var out Record
	if !r.headerRead {
		if err := r.initialize(out); err != nil {
			return out, stack.Trace(err)
		}
	}
	row, err := r.nextRow()
	if err != nil {
		return out, stack.Trace(err)
	}
	vOf := reflect.ValueOf(&out)
	tData := vOf.Elem()
	instructionSet := fieldCache.GetTypeDataFor(reflect.TypeOf(out))
	for cellOffset, cell := range row {
		fieldData := instructionSet.GetFieldByName(r.headers[cellOffset])
		if fieldData == nil {
			continue
		}
		var isNull bool
		if len(cell) == 0 {
			isNull = true
		}
		endOfBlock := len(cell) - 1
		if len(cell) > 2 && (cell[0] == '"' && cell[endOfBlock] == '"') {
			cell = cell[1:endOfBlock]
		}
		val, err := fieldData.InstructionData().GetDecoder()(cell, isNull)
		if err != nil {
			return out, stack.Trace(err)
		}
		tData.Field(fieldData.Idx).Set(reflect.ValueOf(val))
	}
	return out, nil
}

func NewStructuredCSVReader[Record any](csvFileHandler *csv.Reader) (*Reader[Record], error) {
	wrapper := &Reader[Record]{
		reader: csvFileHandler,
	}
	return wrapper, nil
}
