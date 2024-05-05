package csv

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/weisbartb/rcache"
)

var tOfMarshalCSV = reflect.TypeFor[MarshalCSV]()
var tOfTextMarshaller = reflect.TypeFor[encoding.TextMarshaler]()
var tOfStringer = reflect.TypeFor[Stringer]()
var tOfUnmarshalCSV = reflect.TypeFor[UnmarshalCSV]()
var tOfTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

type encoderFunction func(val reflect.Value) (string, error)

type decoderFunction func(val string, isNull bool) (any, error)

func getEncoderProvider(fieldType reflect.Type, omitEmpty bool) encoderFunction {
	if fieldType.Implements(tOfMarshalCSV) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			return val.Interface().(MarshalCSV).MarshalCSV()
		}
	} else if fieldType.Implements(tOfTextMarshaller) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			out, err := val.Interface().(encoding.TextMarshaler).MarshalText()
			return string(out), err
		}
	} else if fieldType.Implements(tOfStringer) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			out, err := val.Interface().(encoding.TextMarshaler).MarshalText()
			return string(out), err
		}
	}
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	switch fieldType.Kind() {
	case reflect.String:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			// This value should not be pre-quoted, the go CSV writer will automatically quote this.
			// Empty quotes are determined to be "less useful" than a null field.
			// See <stdlib>/src/encoding/csv/writer.go:148
			return val.String(), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			return strconv.FormatInt(val.Int(), 10), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			return strconv.FormatUint(val.Uint(), 10), nil
		}
	case reflect.Float32, reflect.Float64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			return strconv.FormatFloat(val.Float(), 'f', -1, 64), nil
		}
	case reflect.Bool:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && val.IsZero() {
				return "", nil
			}
			if val.Bool() {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
	default:
		return func(val reflect.Value) (string, error) {
			return "", fmt.Errorf("can not serialize type %v", fieldType.Kind())
		}
	}
}

func getDecoderProvider(fieldType reflect.Type, fieldName string, required bool) decoderFunction {
	var errFieldRequired = fmt.Errorf("%v is a required field", fieldName)
	if fieldType.Kind() != reflect.Ptr {
		// Create a pointer for a value type to assert if an interface can be applied
		fieldType = reflect.New(fieldType).Type()
	}
	if fieldType.Implements(tOfUnmarshalCSV) {
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			data := reflect.New(fieldType.Elem()).Interface()
			err := data.(UnmarshalCSV).UnmarshalCSV(s)
			ref := reflect.ValueOf(data)
			if ref.Kind() == reflect.Ptr {
				// Unmarshalling creates a new value for store the result (data)
				// When data is unmarshalled it is always the addressable version
				// This (and above) reflect the value again to get it to a value rather than a pointer.
				// Failure to do this result in errors such as
				//	reflect.Set: value of type *csv.NullableField[int] is not assignable to type csv.NullableField[int]
				ref = ref.Elem()
			}
			return ref.Interface(), err
		}
	} else if fieldType.Implements(tOfTextUnmarshaler) {
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			data := reflect.New(fieldType.Elem()).Interface()
			err := data.(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
			return data, err
		}
	}
	switch fieldType.Elem().Kind() {
	case reflect.String:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			return s, nil
		}
	case reflect.Int:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseInt(s, 10, strconv.IntSize)
			return int(val), err
		}
	case reflect.Int8:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseInt(s, 10, 8)
			return int8(val), err
		}
	case reflect.Int16:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseInt(s, 10, 16)
			return int16(val), err
		}
	case reflect.Int32:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseInt(s, 10, 32)
			return int32(val), err
		}
	case reflect.Int64:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			return strconv.ParseInt(s, 10, 64)
		}
	case reflect.Uint:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseUint(s, 10, strconv.IntSize)
			return uint(val), err
		}
	case reflect.Uint8:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseUint(s, 10, 8)
			return uint8(val), err
		}
	case reflect.Uint16:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseUint(s, 10, 16)
			return uint16(val), err
		}
	case reflect.Uint32:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			val, err := strconv.ParseUint(s, 10, 32)
			return uint32(val), err
		}
	case reflect.Uint64:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			return strconv.ParseUint(s, 10, 64)
		}
	case reflect.Float32:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			f, err := strconv.ParseFloat(s, 32)
			return float32(f), err
		}
	case reflect.Float64:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return 0, nil
			}
			f, err := strconv.ParseFloat(s, 64)
			return f, err
		}
	case reflect.Bool:
		return func(s string, isNull bool) (any, error) {
			if required && isNull {
				return nil, errFieldRequired
			}
			if len(s) == 0 {
				return false, nil
			}
			return strconv.ParseBool(s)
		}
	default:
		return func(s string, isNull bool) (any, error) {
			return "", fmt.Errorf("can not unserialize type %v", fieldType.Kind())
		}
	}
}

var _ rcache.InstructionSet = (*csvInstruction)(nil)

type tagParts []string

func (tp tagParts) Find(key string) (string, bool) {
	for _, v := range tp {
		// Loop through to find the sub tag
		if strings.HasPrefix(v, key) {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 1 {
				// Present but has no value
				return "", true
			}
			// Present with a value
			return kv[1], true
		}
	}
	return "", false
}

type csvInstruction struct {
	encoder           encoderFunction
	decoder           decoderFunction
	exportedFieldName string
}

func (c csvInstruction) GetExportedFieldName() string {
	return c.exportedFieldName
}
func (c csvInstruction) FieldName(tag string) string {
	return strings.SplitN(tag, ",", 2)[0]
}

func (c csvInstruction) TagNamespace() string {
	return "csv"
}

func (c csvInstruction) Skip(tag string) bool {
	if strings.SplitN(tag, ",", 2)[0] == "-" {
		return true
	}
	return false
}

func (c csvInstruction) GetMetadata(field reflect.StructField, tag string) rcache.InstructionSet {
	parts := tagParts(strings.Split(tag, ","))
	var omitEmpty bool
	_, omitEmpty = parts.Find("omitempty")
	var required bool
	_, required = parts.Find("required")
	var instruction csvInstruction
	fieldName := c.FieldName(tag)
	instruction.encoder = getEncoderProvider(field.Type, omitEmpty)
	instruction.decoder = getDecoderProvider(field.Type, fieldName, required)
	c.exportedFieldName = fieldName
	return instruction
}

func (c csvInstruction) GetDecoder() decoderFunction {
	return c.decoder
}
func (c csvInstruction) GetEncoder() encoderFunction {
	return c.encoder
}

var fieldCache = rcache.NewCache[csvInstruction]()
