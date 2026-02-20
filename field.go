package mach

import (
	"math"
	"time"
)

type FieldType uint8

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	DurationType
	ErrorType
	TimeType
	BytesType
)

type Field struct {
	Key  string
	Type FieldType
	Ival int64
	Str  string
	Bval []byte
}

func String(key, val string) Field {
	return Field{Key: key, Type: StringType, Str: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Type: IntType, Ival: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Type: Int64Type, Ival: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Type: Float64Type, Ival: int64(math.Float64bits(val))}
}

func Bool(key string, val bool) Field {
	var v int64
	if val {
		v = 1
	}
	return Field{Key: key, Type: BoolType, Ival: v}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: DurationType, Ival: int64(val)}
}

func Err(val error) Field {
	if val == nil {
		return Field{Key: "error", Type: StringType, Str: ""}
	}
	return Field{Key: "error", Type: ErrorType, Str: val.Error()}
}

func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: TimeType, Ival: val.UnixNano()}
}

func Bytes(key string, val []byte) Field {
	return Field{Key: key, Type: BytesType, Bval: val}
}
