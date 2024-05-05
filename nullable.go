package csv

import (
	"errors"
	"reflect"
)

// NullableField allows any type (T) to be nullable,
// the default CSV struct mapper will always use a zero value for a given field for any scalar value.
// This is a wrapper for nullable values to exist and easier to work with that something like sql.Null.
type NullableField[T any] []T

func (n *NullableField[T]) UnmarshalCSV(data string) error {
	var blank T
	if len(data) == 0 {
		return nil
	}
	val, err := getDecoderProvider(reflect.TypeOf(blank), "nullable value", false)(data, false)
	if err != nil {
		return err
	}
	typedVal, ok := val.(T)
	if !ok {
		return errors.New("invalid decoder response for nullable")
	}
	tmp := NullableField[T]{typedVal}
	*n = tmp
	return nil
}

func (n NullableField[T]) MarshalCSV() (string, error) {
	if len(n) == 0 {
		return "", nil
	}
	vOf := reflect.ValueOf(n[0])
	return getEncoderProvider(vOf.Type(), false)(vOf)
}

func (n NullableField[T]) IsNull() bool {
	return len(n) == 0
}

func (n *NullableField[T]) Set(val T) {
	if len(*n) > 0 {
		slc := *n
		slc[0] = val
		return
	}
	tmp := NullableField[T]{val}
	*n = tmp
}
func (n NullableField[T]) Get() (T, bool) {
	var e T
	if len(n) == 0 {
		return e, false
	}
	return n[0], true
}
