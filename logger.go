package mach

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/MYK12397/gohotpool"
)

type Logger struct {
	output  io.Writer
	level   *AtomicLevel
	pool    *gohotpool.Pool
	context []byte
}

type Config struct {
	Output     io.Writer
	Level      Level
	PoolConfig *gohotpool.Config
}

func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = SyncWriter(os.Stderr)
	}

	var pool *gohotpool.Pool
	if cfg.PoolConfig != nil {
		pool = gohotpool.NewPool(*cfg.PoolConfig)
	} else {
		pool = gohotpool.NewPool(gohotpool.Config{
			PoolSize:          512,
			ShardCount:        16,
			DefaultBufferSize: 1024,
			EnableRingBuffer:  false,
			TrackStats:        false,
		})
	}

	return &Logger{
		output: cfg.Output,
		level:  NewAtomicLevel(cfg.Level),
		pool:   pool,
	}
}

func (l *Logger) With(fields ...Field) *Logger {
	if len(fields) == 0 {
		return l
	}

	buf := l.pool.Get()
	b := buf.B
	for _, f := range fields {
		b = append(b, ',')
		b = appendField(b, f)
	}
	encoded := make([]byte, len(b))
	copy(encoded, b)
	buf.B = b
	buf.Reset()
	l.pool.Put(buf)

	child := &Logger{
		output: l.output,
		level:  l.level,
		pool:   l.pool,
	}

	if len(l.context) > 0 {
		child.context = make([]byte, len(l.context)+len(encoded))
		copy(child.context, l.context)
		copy(child.context[len(l.context):], encoded)
	} else {
		child.context = encoded
	}

	return child
}

func (l *Logger) SetLevel(level Level) {
	l.level.SetLevel(level)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	if !l.level.Enabled(DebugLevel) {
		return
	}
	l.log(DebugLevel, msg, fields)
}

func (l *Logger) Info(msg string, fields ...Field) {
	if !l.level.Enabled(InfoLevel) {
		return
	}
	l.log(InfoLevel, msg, fields)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	if !l.level.Enabled(WarnLevel) {
		return
	}
	l.log(WarnLevel, msg, fields)
}

func (l *Logger) Error(msg string, fields ...Field) {
	if !l.level.Enabled(ErrorLevel) {
		return
	}
	l.log(ErrorLevel, msg, fields)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields)
	os.Exit(1)
}

func (l *Logger) log(level Level, msg string, fields []Field) {
	buf := l.pool.Get()
	b := buf.B

	b = append(b, `{"level":"`...)
	b = append(b, level.String()...)
	b = append(b, `","ts":`...)
	b = appendTime(b, time.Now())
	b = append(b, `,"msg":`...)
	b = appendJSONString(b, msg)

	if len(l.context) > 0 {
		b = append(b, l.context...)
	}

	for i := range fields {
		b = append(b, ',')
		b = appendField(b, fields[i])
	}

	b = append(b, '}', '\n')

	buf.B = b
	_, _ = l.output.Write(buf.B)

	buf.Reset()
	l.pool.Put(buf)
}

type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

// SyncWriter wraps w with a mutex so concurrent Write calls are serialized.
func SyncWriter(w io.Writer) io.Writer {
	return &syncWriter{w: w}
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	n, err := s.w.Write(p)
	s.mu.Unlock()
	return n, err
}

func Nop() *Logger {
	return New(Config{
		Output: io.Discard,
		Level:  DebugLevel,
	})
}
