package csv

import (
	"errors"
	"reflect"
)

// NullableField allows any type (T) to be nullable,
// the default CSV struct mapper will always use a zero value for a given field for any scalar value.
// This is a wrapper for nullable values to exist and easier to work with that something like sql.Null.
type NullableField[T any] []T

// UnmarshalCSV allows for a NullableField to be unmarshalled and its underlying type to be resolved if not null.
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

// MarshalCSV marshals the underlying type for a CSV.
func (n NullableField[T]) MarshalCSV() (string, error) {
	if len(n) == 0 {
		return "", nil
	}
	vOf := reflect.ValueOf(n[0])
	return getEncoderProvider(vOf.Type(), false)(vOf)
}

// IsNull checks to see if the field is null or a value was set.
func (n NullableField[T]) IsNull() bool {
	return len(n) == 0
}

// Set updates the nullable field value.
func (n *NullableField[T]) Set(val T) {
	if len(*n) > 0 {
		slc := *n
		slc[0] = val
		return
	}
	tmp := NullableField[T]{val}
	*n = tmp
}

// Unset sets the field to null.
func (n *NullableField[T]) Unset() {
	if len(*n) == 0 {
		return
	}
	tmp := NullableField[T]{}
	*n = tmp
}

// Get returns the value and if it was set or not.
func (n NullableField[T]) Get() (T, bool) {
	var e T
	if len(n) == 0 {
		return e, false
	}
	return n[0], true
}
