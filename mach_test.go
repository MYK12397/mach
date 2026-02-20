package mach

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newMachLogger() *Logger {
	return New(Config{
		Output: io.Discard,
		Level:  DebugLevel,
	})
}

func newZapLogger() *zap.Logger {
	cfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(io.Discard),
		zapcore.DebugLevel,
	)
	return zap.New(core)
}

func newSlogLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func BenchmarkSimple_Mach(b *testing.B) {
	l := newMachLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("simple log message with no fields")
	}
}

func BenchmarkSimple_Zap(b *testing.B) {
	l := newZapLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("simple log message with no fields")
	}
}

func BenchmarkSimple_Slog(b *testing.B) {
	l := newSlogLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("simple log message with no fields")
	}
}

func BenchmarkFiveFields_Mach(b *testing.B) {
	l := newMachLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request completed",
			String("method", "GET"),
			String("path", "/api/v1/users"),
			Int("status", 200),
			Duration("latency", 1532*time.Microsecond),
			String("ip", "192.168.1.42"),
		)
	}
}

func BenchmarkFiveFields_Zap(b *testing.B) {
	l := newZapLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request completed",
			zap.String("method", "GET"),
			zap.String("path", "/api/v1/users"),
			zap.Int("status", 200),
			zap.Duration("latency", 1532*time.Microsecond),
			zap.String("ip", "192.168.1.42"),
		)
	}
}

func BenchmarkFiveFields_Slog(b *testing.B) {
	l := newSlogLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request completed",
			slog.String("method", "GET"),
			slog.String("path", "/api/v1/users"),
			slog.Int("status", 200),
			slog.Duration("latency", 1532*time.Microsecond),
			slog.String("ip", "192.168.1.42"),
		)
	}
}

func BenchmarkTenFields_Mach(b *testing.B) {
	l := newMachLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("audit event",
			String("action", "user.login"),
			String("user_id", "usr_9f8a7b6c"),
			String("email", "alice@example.com"),
			String("ip", "10.0.0.1"),
			String("user_agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"),
			Int("status", 200),
			Int64("request_size", 1456),
			Float64("confidence", 0.9987),
			Bool("mfa_used", true),
			Duration("auth_latency", 45*time.Millisecond),
		)
	}
}

func BenchmarkTenFields_Zap(b *testing.B) {
	l := newZapLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("audit event",
			zap.String("action", "user.login"),
			zap.String("user_id", "usr_9f8a7b6c"),
			zap.String("email", "alice@example.com"),
			zap.String("ip", "10.0.0.1"),
			zap.String("user_agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"),
			zap.Int("status", 200),
			zap.Int64("request_size", 1456),
			zap.Float64("confidence", 0.9987),
			zap.Bool("mfa_used", true),
			zap.Duration("auth_latency", 45*time.Millisecond),
		)
	}
}

func BenchmarkTenFields_Slog(b *testing.B) {
	l := newSlogLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("audit event",
			slog.String("action", "user.login"),
			slog.String("user_id", "usr_9f8a7b6c"),
			slog.String("email", "alice@example.com"),
			slog.String("ip", "10.0.0.1"),
			slog.String("user_agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"),
			slog.Int("status", 200),
			slog.Int64("request_size", 1456),
			slog.Float64("confidence", 0.9987),
			slog.Bool("mfa_used", true),
			slog.Duration("auth_latency", 45*time.Millisecond),
		)
	}
}

func BenchmarkWithContext_Mach(b *testing.B) {
	base := newMachLogger()
	l := base.With(
		String("service", "api-gateway"),
		String("version", "2.4.1"),
		String("env", "production"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request handled",
			String("method", "POST"),
			Int("status", 201),
		)
	}
}

func BenchmarkWithContext_Zap(b *testing.B) {
	base := newZapLogger()
	l := base.With(
		zap.String("service", "api-gateway"),
		zap.String("version", "2.4.1"),
		zap.String("env", "production"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request handled",
			zap.String("method", "POST"),
			zap.Int("status", 201),
		)
	}
}

func BenchmarkWithContext_Slog(b *testing.B) {
	base := newSlogLogger()
	l := base.With(
		slog.String("service", "api-gateway"),
		slog.String("version", "2.4.1"),
		slog.String("env", "production"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("request handled",
			slog.String("method", "POST"),
			slog.Int("status", 201),
		)
	}
}

func BenchmarkDisabled_Mach(b *testing.B) {
	l := New(Config{
		Output: io.Discard,
		Level:  ErrorLevel,
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("this should be skipped",
			String("key", "value"),
			Int("count", 42),
		)
	}
}

func BenchmarkDisabled_Zap(b *testing.B) {
	l := newZapLogger().WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("this should be skipped",
			zap.String("key", "value"),
			zap.Int("count", 42),
		)
	}
}

func BenchmarkDisabled_Slog(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("this should be skipped",
			slog.String("key", "value"),
			slog.Int("count", 42),
		)
	}
}

func BenchmarkParallel_Mach(b *testing.B) {
	l := newMachLogger()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel log entry",
				String("method", "GET"),
				String("path", "/health"),
				Int("status", 200),
				Duration("latency", 250*time.Microsecond),
			)
		}
	})
}

func BenchmarkParallel_Zap(b *testing.B) {
	l := newZapLogger()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel log entry",
				zap.String("method", "GET"),
				zap.String("path", "/health"),
				zap.Int("status", 200),
				zap.Duration("latency", 250*time.Microsecond),
			)
		}
	})
}

func BenchmarkParallel_Slog(b *testing.B) {
	l := newSlogLogger()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel log entry",
				slog.String("method", "GET"),
				slog.String("path", "/health"),
				slog.Int("status", 200),
				slog.Duration("latency", 250*time.Microsecond),
			)
		}
	})
}

var longString = func() string {
	b := make([]byte, 1024)
	for i := range b {
		b[i] = 'a' + byte(i%26)
	}
	return string(b)
}()

func BenchmarkLargePayload_Mach(b *testing.B) {
	l := newMachLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("large payload test",
			String("body", longString),
			String("trace_id", "abc123def456ghi789jkl012mno345pq"),
			Int64("content_length", 102400),
			Float64("score", 99.9876),
			Bool("compressed", true),
		)
	}
}

func BenchmarkLargePayload_Zap(b *testing.B) {
	l := newZapLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("large payload test",
			zap.String("body", longString),
			zap.String("trace_id", "abc123def456ghi789jkl012mno345pq"),
			zap.Int64("content_length", 102400),
			zap.Float64("score", 99.9876),
			zap.Bool("compressed", true),
		)
	}
}

func BenchmarkLargePayload_Slog(b *testing.B) {
	l := newSlogLogger()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("large payload test",
			slog.String("body", longString),
			slog.String("trace_id", "abc123def456ghi789jkl012mno345pq"),
			slog.Int64("content_length", 102400),
			slog.Float64("score", 99.9876),
			slog.Bool("compressed", true),
		)
	}
}

func BenchmarkError_Mach(b *testing.B) {
	l := newMachLogger()
	err := io.ErrUnexpectedEOF
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("operation failed",
			Err(err),
			String("component", "database"),
			Int("retry", 3),
		)
	}
}

func BenchmarkError_Zap(b *testing.B) {
	l := newZapLogger()
	err := io.ErrUnexpectedEOF
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("operation failed",
			zap.Error(err),
			zap.String("component", "database"),
			zap.Int("retry", 3),
		)
	}
}

func BenchmarkError_Slog(b *testing.B) {
	l := newSlogLogger()
	err := io.ErrUnexpectedEOF
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Error("operation failed",
			slog.String("error", err.Error()),
			slog.String("component", "database"),
			slog.Int("retry", 3),
		)
	}
}

func BenchmarkParallelWithContext_Mach(b *testing.B) {
	base := newMachLogger()
	l := base.With(
		String("service", "user-api"),
		String("version", "3.1.0"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("request",
				String("method", "GET"),
				String("path", "/users/123"),
				Int("status", 200),
				Duration("latency", 800*time.Microsecond),
			)
		}
	})
}

func BenchmarkParallelWithContext_Zap(b *testing.B) {
	base := newZapLogger()
	l := base.With(
		zap.String("service", "user-api"),
		zap.String("version", "3.1.0"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("request",
				zap.String("method", "GET"),
				zap.String("path", "/users/123"),
				zap.Int("status", 200),
				zap.Duration("latency", 800*time.Microsecond),
			)
		}
	})
}

func BenchmarkParallelWithContext_Slog(b *testing.B) {
	base := newSlogLogger()
	l := base.With(
		slog.String("service", "user-api"),
		slog.String("version", "3.1.0"),
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("request",
				slog.String("method", "GET"),
				slog.String("path", "/users/123"),
				slog.Int("status", 200),
				slog.Duration("latency", 800*time.Microsecond),
			)
		}
	})
}
