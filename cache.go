package csv

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/weisbartb/rcache"
)

// various reflection marshalling methods
var tOfMarshalCSV = reflect.TypeFor[MarshalCSV]()
var tOfTextMarshaller = reflect.TypeFor[encoding.TextMarshaler]()
var tOfStringer = reflect.TypeFor[Stringer]()
var tOfUnmarshalCSV = reflect.TypeFor[UnmarshalCSV]()
var tOfTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
var tOfZeroer = reflect.TypeFor[Zeroer]()

// encoderFunction is what is used to take a value and encode it into a string response for the CSV
type encoderFunction func(val reflect.Value) (string, error)

// decoderFunction is a what is used to decode a value from a string into a response for the CSV.
// isNull is calculated by the instruction set;
// however, it will always generate a false positive for strings that do not use empty double quotes.
type decoderFunction func(val string, isNull bool) (any, error)

// zeroValueFunction is a helper stub to hold which isZero detection to use.
type zeroValueFunction func(value reflect.Value) bool

func isZero(value reflect.Value) bool {
	return value.IsZero()
}

func isZeroZeroer(value reflect.Value) bool {
	return value.Interface().(Zeroer).IsZero()
}

// getEncoderProvider returns a memoized function for encoding values based on their scalar types.
// structs, slices, and maps are not supported natively and should implement a MarshalCSV interface.
func getEncoderProvider(fieldType reflect.Type, omitEmpty bool) encoderFunction {
	var zeroerFunc zeroValueFunction = isZero
	if fieldType.Implements(tOfZeroer) {
		// Use the interface resolver rather than the reflection library
		zeroerFunc = isZeroZeroer
	}
	// Check to see if MarshalCSV is implemented
	if fieldType.Implements(tOfMarshalCSV) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			return val.Interface().(MarshalCSV).MarshalCSV()
		}
		// Check to see if encoding.TextMarshaler is implemented
	} else if fieldType.Implements(tOfTextMarshaller) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			out, err := val.Interface().(encoding.TextMarshaler).MarshalText()
			return string(out), err
		}
		// Check to see if Stringer is implemented
	} else if fieldType.Implements(tOfStringer) {
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
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
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			// This value should not be pre-quoted, the go CSV writer will automatically quote this.
			// Empty quotes are determined to be "less useful" than a null field.
			// See <stdlib>/src/encoding/csv/writer.go:148
			return val.String(), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			// All ints can be accessed via Int()
			return strconv.FormatInt(val.Int(), 10), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			// All uints can be accessed via Unit
			return strconv.FormatUint(val.Uint(), 10), nil
		}
	case reflect.Float32, reflect.Float64:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
				return "", nil
			}
			return strconv.FormatFloat(val.Float(), 'f', -1, 64), nil
		}
	case reflect.Bool:
		return func(val reflect.Value) (string, error) {
			if omitEmpty && zeroerFunc(val) {
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

// getDecoderProvider returns a memoized function for decoding values based on their scalar types.
// structs, slices, and maps are not supported natively and should implement an UnmarshalCSV interface.
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
			ref := reflect.ValueOf(data)
			if ref.Kind() == reflect.Ptr {
				// See comments for fieldType.Implements(tOfUnmarshalCSV)
				ref = ref.Elem()
			}
			return ref.Interface(), err
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

// tagParts is a quick helper type for parsing the extra tag arguments.
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

// Ensure that the csvInstruction can be used by rcache.
var _ rcache.InstructionSet = (*csvInstruction)(nil)

// csvInstruction provides instructions on how to extract data from structs for CSV parsing.
type csvInstruction struct {
	encoder           encoderFunction
	decoder           decoderFunction
	exportedFieldName string
}

// GetCSVHeaderIdentifier gets the mapping identifier for the CSV header.
func (c csvInstruction) GetCSVHeaderIdentifier() string {
	return c.exportedFieldName
}

// FieldName gets the name of the field from the given tag, this is needed by InstructionSet.
func (c csvInstruction) FieldName(tag string) string {
	return strings.SplitN(tag, ",", 2)[0]
}

// TagNamespace gets the namespace for the tag this instruction set wants to use
func (c csvInstruction) TagNamespace() string {
	return "csv"
}

// Skip determines if the potential field should be skipped based on its tag.
func (c csvInstruction) Skip(tag string) bool {
	if strings.SplitN(tag, ",", 2)[0] == "-" {
		return true
	}
	return false
}

// GetMetadata is a method for calculating metadata for a given field;
// this will return a new instruction set from the base instruction for how to parse the field.
func (c csvInstruction) GetMetadata(field reflect.StructField, tag string) rcache.InstructionSet {
	var omitEmpty bool
	var required bool
	parts := tagParts(strings.Split(tag, ","))
	if len(parts) > 1 {
		// Skip past the field name declaration.
		parts = parts[1:]
		_, omitEmpty = parts.Find("omitempty")
		_, required = parts.Find("required")
	}
	var instruction csvInstruction
	fieldName := c.FieldName(tag)
	instruction.encoder = getEncoderProvider(field.Type, omitEmpty)
	instruction.decoder = getDecoderProvider(field.Type, fieldName, required)
	c.exportedFieldName = fieldName
	return instruction
}

// GetDecoder gets the decoder for a given field.
func (c csvInstruction) GetDecoder() decoderFunction {
	return c.decoder
}

// GetEncoder gets the encoder for a given field.
func (c csvInstruction) GetEncoder() encoderFunction {
	return c.encoder
}

// Setup the cache
var fieldCache = rcache.NewCache[csvInstruction]()
