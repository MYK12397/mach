package mach

import "sync/atomic"

type Level int32

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = [...]string{
	DebugLevel + 1: "DEBUG",
	InfoLevel + 1:  "INFO",
	WarnLevel + 1:  "WARN",
	ErrorLevel + 1: "ERROR",
	FatalLevel + 1: "FATAL",
}

func (l Level) String() string {
	idx := l + 1
	if idx >= 0 && int(idx) < len(levelNames) {
		return levelNames[idx]
	}
	return "UNKNOWN"
}

type AtomicLevel struct {
	v int32
}

func NewAtomicLevel(l Level) *AtomicLevel {
	return &AtomicLevel{v: int32(l)}
}

func (al *AtomicLevel) Level() Level {
	return Level(atomic.LoadInt32(&al.v))
}

func (al *AtomicLevel) SetLevel(l Level) {
	atomic.StoreInt32(&al.v, int32(l))
}

func (al *AtomicLevel) Enabled(l Level) bool {
	return l >= Level(atomic.LoadInt32(&al.v))
}
