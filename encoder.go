package mach

import (
	"math"
	"time"
	"unicode/utf8"

	"github.com/MYK12397/gohotpool"
)

func appendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	dst = appendEscapedString(dst, s)
	return append(dst, '"')
}

func appendKey(dst []byte, key string) []byte {
	dst = append(dst, '"')
	dst = appendEscapedString(dst, key)
	return append(dst, '"', ':')
}

func appendBool(dst []byte, v bool) []byte {
	if v {
		return append(dst, "true"...)
	}
	return append(dst, "false"...)
}

func appendTime(dst []byte, t time.Time) []byte {
	dst = append(dst, '"')
	dst = t.AppendFormat(dst, time.RFC3339Nano)
	return append(dst, '"')
}

func appendDuration(dst []byte, d time.Duration) []byte {
	return appendFloat64(dst, d.Seconds())
}

func appendField(dst []byte, f Field) []byte {
	dst = appendKey(dst, f.Key)
	switch f.Type {
	case StringType:
		dst = appendJSONString(dst, f.Str)
	case IntType, Int64Type:
		dst = appendInt64(dst, f.Ival)
	case Float64Type:
		dst = appendFloat64(dst, math.Float64frombits(uint64(f.Ival)))
	case BoolType:
		dst = appendBool(dst, f.Ival == 1)
	case DurationType:
		dst = appendDuration(dst, time.Duration(f.Ival))
	case ErrorType:
		dst = appendJSONString(dst, f.Str)
	case TimeType:
		dst = appendTime(dst, time.Unix(0, f.Ival))
	case BytesType:
		dst = appendJSONString(dst, string(f.Bval))
	}
	return dst
}

type Encoder struct {
	buf *gohotpool.Buffer
}

func NewEncoder(buf *gohotpool.Buffer) *Encoder {
	return &Encoder{buf: buf}
}

func (e *Encoder) AppendByte(c byte)              { e.buf.B = append(e.buf.B, c) }
func (e *Encoder) AppendString(s string)          { e.buf.B = append(e.buf.B, s...) }
func (e *Encoder) AppendJSONString(s string)      { e.buf.B = appendJSONString(e.buf.B, s) }
func (e *Encoder) AppendKey(key string)           { e.buf.B = appendKey(e.buf.B, key) }
func (e *Encoder) AppendInt64(v int64)            { e.buf.B = appendInt64(e.buf.B, v) }
func (e *Encoder) AppendFloat64(v float64)        { e.buf.B = appendFloat64(e.buf.B, v) }
func (e *Encoder) AppendBool(v bool)              { e.buf.B = appendBool(e.buf.B, v) }
func (e *Encoder) AppendTime(t time.Time)         { e.buf.B = appendTime(e.buf.B, t) }
func (e *Encoder) AppendDuration(d time.Duration) { e.buf.B = appendDuration(e.buf.B, d) }
func (e *Encoder) AppendField(f Field)            { e.buf.B = appendField(e.buf.B, f) }

var safeSet [256]bool

func init() {
	for i := 0x20; i <= 0x7E; i++ {
		safeSet[i] = true
	}
	safeSet['"'] = false
	safeSet['\\'] = false
}

func appendEscapedString(dst []byte, s string) []byte {
	i := 0
	for i < len(s) {
		start := i
		for i < len(s) && safeSet[s[i]] {
			i++
		}
		if i > start {
			dst = append(dst, s[start:i]...)
		}
		if i >= len(s) {
			break
		}

		b := s[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size == 1 {
				dst = append(dst, '\\', 'u', 'f', 'f', 'f', 'd')
			} else {
				dst = append(dst, s[i:i+size]...)
			}
			i += size
			continue
		}
		switch b {
		case '"':
			dst = append(dst, '\\', '"')
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hexDigit(b>>4), hexDigit(b&0x0f))
		}
		i++
	}
	return dst
}

func hexDigit(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

func appendInt64(dst []byte, v int64) []byte {
	if v == 0 {
		return append(dst, '0')
	}
	if v < 0 {
		dst = append(dst, '-')
		if v == -9223372036854775808 {
			return append(dst, "9223372036854775808"...)
		}
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte(v%10) + '0'
		v /= 10
	}
	return append(dst, buf[i:]...)
}

func appendFloat64(dst []byte, v float64) []byte {
	if math.IsNaN(v) {
		return append(dst, `"NaN"`...)
	}
	if math.IsInf(v, 1) {
		return append(dst, `"+Inf"`...)
	}
	if math.IsInf(v, -1) {
		return append(dst, `"-Inf"`...)
	}
	if v == 0 {
		return append(dst, '0')
	}

	neg := false
	if v < 0 {
		neg = true
		v = -v
	}

	if v == math.Trunc(v) && v < 1e15 {
		if neg {
			dst = append(dst, '-')
		}
		return appendInt64(dst, int64(v))
	}

	if neg {
		dst = append(dst, '-')
	}

	intPart := int64(v)
	fracPart := int64(math.Round((v - float64(intPart)) * 1e6))
	if fracPart < 0 {
		fracPart = -fracPart
	}

	dst = appendInt64(dst, intPart)
	dst = append(dst, '.')

	var fbuf [6]byte
	for i := 5; i >= 0; i-- {
		fbuf[i] = byte(fracPart%10) + '0'
		fracPart /= 10
	}
	end := 6
	for end > 1 && fbuf[end-1] == '0' {
		end--
	}
	dst = append(dst, fbuf[:end]...)
	return dst
}
