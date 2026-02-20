# mach
 
A zero-allocation structured, leveled JSON logger for Go, built on [gohotpool](https://github.com/MYK12397/gohotpool).
 
**2x faster than zap. 3-5x faster than slog. Zero allocations.**
 
## Why
 
Every structured logger in Go pays the same costs: buffer allocation, JSON encoding overhead, interface boxing for field values, and lock contention under parallelism. mach eliminates all of them:
 
- **Buffer pooling via gohotpool** — per-P (processor) cached byte buffers with clock-sweep eviction. No `sync.Pool` GC jitter, no allocation on the hot path.
- **Typed fields** — `Field` structs carry data in fixed-size `int64`/`string` slots selected by a type tag, avoiding `interface{}` heap escapes entirely.
- **Hand-rolled JSON encoding** — all encoding functions operate directly on `[]byte` with zero intermediate allocations. String escaping uses a bulk-scan fast path over safe ASCII ranges.
- **No logger-level mutex** — each goroutine encodes into its own pooled buffer and writes directly to the output. Locking is opt-in at the writer level via `SyncWriter()`.
 
## Install
 
```
go get github.com/MYK12397/gohotpool
```
 
mach lives alongside gohotpool. Copy the `mach/` package into your project or reference it as a local module.
 
## Usage
 
```go
package main
 
import (
    "os"
    "time"
 
    "mach"
)
 
func main() {
    log := mach.New(mach.Config{
        Output: mach.SyncWriter(os.Stdout),
        Level:  mach.InfoLevel,
    })
 
    log.Info("server started",
        mach.String("addr", ":8080"),
        mach.Int("workers", 4),
    )
 
    reqLog := log.With(
        mach.String("service", "api"),
        mach.String("version", "1.2.0"),
    )
 
    reqLog.Info("request handled",
        mach.String("method", "GET"),
        mach.String("path", "/users"),
        mach.Int("status", 200),
        mach.Duration("latency", 1200*time.Microsecond),
    )
 
    reqLog.Error("query failed",
        mach.Err(os.ErrNotExist),
        mach.String("table", "users"),
    )
}
```
 
Output:
 
```json
{"level":"INFO","ts":"2026-02-17T22:30:00.123456789Z","msg":"server started","addr":":8080","workers":4}
{"level":"INFO","ts":"2026-02-17T22:30:00.123556789Z","msg":"request handled","service":"api","version":"1.2.0","method":"GET","path":"/users","status":200,"latency":0.0012}
{"level":"ERROR","ts":"2026-02-17T22:30:00.123656789Z","msg":"query failed","service":"api","version":"1.2.0","error":"file does not exist","table":"users"}
```
 
## API
 
### Logger
 
```go
mach.New(cfg Config) *Logger
 
logger.Debug(msg string, fields ...Field)
logger.Info(msg string, fields ...Field)
logger.Warn(msg string, fields ...Field)
logger.Error(msg string, fields ...Field)
logger.Fatal(msg string, fields ...Field)   // calls os.Exit(1)
 
logger.With(fields ...Field) *Logger        // child logger with pre-encoded context
logger.SetLevel(level Level)                // change level at runtime (atomic)
```
 
### Fields
 
```go
mach.String(key, val string)
mach.Int(key string, val int)
mach.Int64(key string, val int64)
mach.Float64(key string, val float64)
mach.Bool(key string, val bool)
mach.Duration(key string, val time.Duration)
mach.Time(key string, val time.Time)
mach.Err(val error)                    // key is "error"
mach.Bytes(key string, val []byte)
```
 
### Writer Safety
 
mach does not hold a logger-level mutex. The output `io.Writer` is called directly from each goroutine. For writers that aren't inherently thread-safe, wrap them:
 
```go
mach.New(mach.Config{
    Output: mach.SyncWriter(file), // adds a mutex
})
```
 
`io.Discard`, and `os.File` writes under `PIPE_BUF` (4KB on Linux) are already atomic at the OS level and don't need wrapping.
 
## Benchmark
 
All benchmarks write to `io.Discard` to isolate pure encoding speed. Run with:
 
```
go test -bench=. -benchmem -count=3
```
 
### Results (Apple M4 Pro, Go 1.23.1)
 
```
goos: darwin
goarch: arm64
cpu: Apple M4 Pro
 
BenchmarkSimple_Mach-12                       75 ns/op     0 B/op    0 allocs/op
BenchmarkSimple_Zap-12                       176 ns/op     0 B/op    0 allocs/op
BenchmarkSimple_Slog-12                      272 ns/op     0 B/op    0 allocs/op
 
BenchmarkFiveFields_Mach-12                  163 ns/op     0 B/op    0 allocs/op
BenchmarkFiveFields_Zap-12                   349 ns/op   320 B/op    1 allocs/op
BenchmarkFiveFields_Slog-12                  563 ns/op   240 B/op    5 allocs/op
 
BenchmarkTenFields_Mach-12                   262 ns/op     0 B/op    0 allocs/op
BenchmarkTenFields_Zap-12                    544 ns/op   705 B/op    1 allocs/op
BenchmarkTenFields_Slog-12                  1031 ns/op   808 B/op   14 allocs/op
 
BenchmarkWithContext_Mach-12                 100 ns/op     0 B/op    0 allocs/op
BenchmarkWithContext_Zap-12                  242 ns/op   128 B/op    1 allocs/op
BenchmarkWithContext_Slog-12                 392 ns/op    96 B/op    2 allocs/op
 
BenchmarkDisabled_Mach-12                    6.3 ns/op    0 B/op    0 allocs/op
BenchmarkDisabled_Zap-12                      22 ns/op  128 B/op    1 allocs/op
BenchmarkDisabled_Slog-12                     32 ns/op   96 B/op    2 allocs/op
 
BenchmarkParallel_Mach-12                     66 ns/op    0 B/op    0 allocs/op
BenchmarkParallel_Zap-12                      96 ns/op  256 B/op    1 allocs/op
BenchmarkParallel_Slog-12                    239 ns/op  192 B/op    4 allocs/op
 
BenchmarkLargePayload_Mach-12               473 ns/op    0 B/op    0 allocs/op
BenchmarkLargePayload_Zap-12               1024 ns/op  320 B/op    1 allocs/op
BenchmarkLargePayload_Slog-12              1351 ns/op  360 B/op    8 allocs/op
 
BenchmarkError_Mach-12                      123 ns/op    0 B/op    0 allocs/op
BenchmarkError_Zap-12                       288 ns/op  192 B/op    1 allocs/op
BenchmarkError_Slog-12                      457 ns/op  144 B/op    3 allocs/op
 
BenchmarkParallelWithContext_Mach-12          63 ns/op    0 B/op    0 allocs/op
BenchmarkParallelWithContext_Zap-12           98 ns/op  256 B/op    1 allocs/op
BenchmarkParallelWithContext_Slog-12         241 ns/op  192 B/op    4 allocs/op
```
 
### Summary
 
| Scenario | mach | zap | slog | vs zap | vs slog |
|---|---|---|---|---|---|
| Simple (no fields) | **75 ns** | 176 ns | 272 ns | 2.3x faster | 3.6x faster |
| 5 fields (mixed types) | **163 ns** | 349 ns | 563 ns | 2.1x faster | 3.5x faster |
| 10 fields | **262 ns** | 544 ns | 1031 ns | 2.1x faster | 3.9x faster |
| With() context | **100 ns** | 242 ns | 392 ns | 2.4x faster | 3.9x faster |
| Disabled level | **6.3 ns** | 22 ns | 32 ns | 3.5x faster | 5.1x faster |
| Parallel (12 cores) | **66 ns** | 96 ns | 239 ns | 1.5x faster | 3.6x faster |
| Large payload (1KB) | **473 ns** | 1024 ns | 1351 ns | 2.2x faster | 2.9x faster |
| Error with fields | **123 ns** | 288 ns | 457 ns | 2.3x faster | 3.7x faster |
| Parallel + context | **63 ns** | 98 ns | 241 ns | 1.6x faster | 3.8x faster |
 
**Zero allocations across every benchmark.** Zap allocates 1 per log (128-705 bytes). Slog allocates 2-14 per log.
 
### Why mach is faster
 
| Technique | What it avoids |
|---|---|
| gohotpool per-P buffer cache | `sync.Pool` overhead + GC jitter |
| Typed `Field` struct with `int64`/`string` slots | `interface{}` heap escapes |
| Direct `[]byte` append encoding | Encoder pointer indirection |
| Bulk-scan string escaping | Per-byte branch for safe ASCII |
| No logger-level mutex | Goroutine serialization under parallelism |
| Pre-encoded `With()` context | Re-encoding context fields on every log |
| Atomic level check short-circuits before variadic alloc | Wasted work on disabled levels |
 
## Design
 
```
log.Info("msg", fields...)
  |
  +-- atomic level check (6ns, short-circuits if disabled)
  |
  +-- gohotpool.Get() (per-P cache hit: ~5ns, 0 allocs)
  |
  +-- encode JSON into local []byte
  |     +-- level string (constant)
  |     +-- timestamp via time.Now().AppendFormat (no temp string)
  |     +-- message via bulk-scan JSON escaping
  |     +-- pre-encoded With() context (memcpy)
  |     +-- typed fields (switch on FieldType, no reflection)
  |
  +-- output.Write(buf) (single syscall, no logger mutex)
  |
  +-- gohotpool.Put() (back to per-P cache)
```
 
## License
 
MIT