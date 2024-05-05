package csv

import (
	"net"
	"reflect"
	"testing"
	"time"

	testifyassert "github.com/stretchr/testify/assert"
)

type MarshallableTime struct {
	time.Time
}

func (m *MarshallableTime) UnmarshalCSV(data string) error {
	t, err := time.Parse(time.RFC3339, data)
	if err != nil {
		return err
	}
	m.Time = t
	return nil
}

func (m MarshallableTime) MarshalCSV() (string, error) {
	return m.Time.Format(time.RFC3339), nil
}

type TestStruct struct {
	Time    MarshallableTime `csv:"marshallableTime,required"`
	IP      net.IP           `csv:"ip,required"`
	String  string           `csv:"string,required"`
	Int     int              `csv:"int,required"`
	Int8    int8             `csv:"int8,required"`
	Int16   int16            `csv:"int16,required"`
	Int32   int32            `csv:"int32,required"`
	Int64   int64            `csv:"int64,required"`
	Uint    uint             `csv:"uint,required"`
	Uint8   uint8            `csv:"uint8,required"`
	Uint16  uint16           `csv:"uint16,required"`
	Uint32  uint32           `csv:"uint32,required"`
	Uint64  uint64           `csv:"uint64,required"`
	Float32 float32          `csv:"float32,required"`
	Float64 float64          `csv:"float64,required"`
	Bool    bool             `csv:"bool,required"`
}

type TestStructPtr struct {
	Time    *MarshallableTime `csv:"marshallableTime,required"`
	IP      *net.IP           `csv:"ip,required"`
	String  *string           `csv:"string,required"`
	Int     *int              `csv:"int,required"`
	Int8    *int8             `csv:"int8,required"`
	Int16   *int16            `csv:"int16,required"`
	Int32   *int32            `csv:"int32,required"`
	Int64   *int64            `csv:"int64,required"`
	Uint    *uint             `csv:"uint,required"`
	Uint8   *uint8            `csv:"uint8,required"`
	Uint16  *uint16           `csv:"uint16,required"`
	Uint32  *uint32           `csv:"uint32,required"`
	Uint64  *uint64           `csv:"uint64,required"`
	Float32 *float32          `csv:"float32,required"`
	Float64 *float64          `csv:"float64,required"`
	Bool    *bool             `csv:"bool,required"`
}

func Test_Cache(t *testing.T) {
	assert := testifyassert.New(t)
	instructions := fieldCache.GetTypeDataFor(reflect.TypeOf(TestStruct{}))
	tmp1 := TestStruct{
		String:  "a string",
		Int:     32,
		Int8:    12,
		Int16:   5235,
		Int32:   1561235,
		Int64:   16561354,
		Uint:    1252,
		Uint8:   12,
		Uint16:  1256,
		Uint32:  123672,
		Uint64:  565124,
		Float32: 5125.23,
		Float64: 5151267.53235,
		Bool:    true,
		Time: MarshallableTime{
			Time: time.Now().Truncate(time.Second),
		},
		IP: net.IPv4(10, 10, 10, 10),
	}
	var tmp2 TestStruct
	var tmp3 TestStructPtr
	valOfTmp1 := reflect.ValueOf(tmp1)
	valOfTmp2 := reflect.ValueOf(&tmp2)
	valOfTmp3 := reflect.ValueOf(&tmp3)
	for _, field := range instructions.Fields() {
		val, err := field.InstructionData().GetEncoder()(valOfTmp1.Field(field.Idx))
		assert.NoError(err)
		assert.NotEmpty(val)
		newTypedVal, err := field.InstructionData().GetDecoder()(val, false)
		assert.NoError(err)
		assert.NotEmpty(newTypedVal)
		for _, currentValue := range []reflect.Value{valOfTmp2, valOfTmp3} {
			targetField := currentValue.Elem().Field(field.Idx)
			if targetField.Kind() != reflect.Ptr && reflect.ValueOf(newTypedVal).Kind() == reflect.Ptr {
				targetField.Set(reflect.ValueOf(newTypedVal).Elem())
			} else if targetField.Kind() == reflect.Ptr && reflect.ValueOf(newTypedVal).Kind() != reflect.Ptr {
				if reflect.ValueOf(newTypedVal).CanAddr() {
					targetField.Set(reflect.ValueOf(newTypedVal).Addr())
				} else {
					vOf := reflect.ValueOf(newTypedVal)
					newCopy := reflect.New(vOf.Type()).Elem()
					newCopy.Set(vOf)
					targetField.Set(newCopy.Addr())
				}
			} else {
				targetField.Set(reflect.ValueOf(newTypedVal))
			}
		}
	}
	assert.Equal(tmp1, tmp2)
	assert.Equal(tmp1.Time, *tmp3.Time)
	assert.Equal(tmp1.String, *tmp3.String)
	assert.Equal(tmp1.Int, *tmp3.Int)
	assert.Equal(tmp1.Int8, *tmp3.Int8)
	assert.Equal(tmp1.Int16, *tmp3.Int16)
	assert.Equal(tmp1.Int32, *tmp3.Int32)
	assert.Equal(tmp1.Int64, *tmp3.Int64)
	assert.Equal(tmp1.Uint, *tmp3.Uint)
	assert.Equal(tmp1.Uint8, *tmp3.Uint8)
	assert.Equal(tmp1.Uint16, *tmp3.Uint16)
	assert.Equal(tmp1.Uint32, *tmp3.Uint32)
	assert.Equal(tmp1.Uint64, *tmp3.Uint64)
	assert.Equal(tmp1.Bool, *tmp3.Bool)
	assert.Equal(tmp1.IP, *tmp3.IP)
	assert.Equal(tmp1.Float32, *tmp3.Float32)
	assert.Equal(tmp1.Float64, *tmp3.Float64)
	var tmp4 TestStruct
	valOfTmp4 := reflect.ValueOf(&tmp4)
	for _, field := range fieldCache.GetTypeDataFor(reflect.TypeOf(tmp3)).Fields() {
		val, err := field.InstructionData().GetEncoder()(valOfTmp1.Field(field.Idx))
		assert.NoError(err)
		assert.NotEmpty(val)
		newTypedVal, err := field.InstructionData().GetDecoder()(val, false)
		assert.NoError(err)
		assert.NotEmpty(newTypedVal)
		targetField := valOfTmp4.Elem().Field(field.Idx)
		if targetField.Kind() != reflect.Ptr && reflect.ValueOf(newTypedVal).Kind() == reflect.Ptr {
			targetField.Set(reflect.ValueOf(newTypedVal).Elem())
		} else if targetField.Kind() == reflect.Ptr && reflect.ValueOf(newTypedVal).Kind() != reflect.Ptr {
			if reflect.ValueOf(newTypedVal).CanAddr() {
				targetField.Set(reflect.ValueOf(newTypedVal).Addr())
			} else {
				vOf := reflect.ValueOf(newTypedVal)
				newCopy := reflect.New(vOf.Type()).Elem()
				newCopy.Set(vOf)
				targetField.Set(newCopy.Addr())
			}
		} else {
			targetField.Set(reflect.ValueOf(newTypedVal))
		}
	}
	assert.Equal(tmp1, tmp4)
}
