# soypat/uuid

A UUID library for Go with explicit dependency injection.

## Usage

```go
package main

import (
    "fmt"

    puuid "github.com/soypat/uuid"
)

// Almost drop-in replacement for google/uuid.
// Can also configure all aspects of generated UUIDs.
var uuid = puuid.NewGeneratorV4()

func main() {
    id := uuid.MustRandom()
    fmt.Println(id.String()) // e.g. "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
}
```

## Parsing

```go
id, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
id, err := uuid.Parse("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}")
id, err := uuid.Parse("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
```

## Differences from google/uuid

| | this package | google/uuid |
|---|---|---|
| Global state | None. Generator must be explicitly initialized | Uses global `rand.Reader` and clock state |
| Dependency injection | Required via `GeneratorConfig` | Optional via `SetRand()` |
| Concurrency | Each Generator has its own mutex | Global mutex for all operations |
| Testing | Easy to mock with custom rand/time sources | Requires global `SetRand()` calls |
| Zero value | `Generator{}` is not usable until `Init()` | Works immediately with defaults |
| Nil/Max UUIDs | `Zero()` and `Max()` functions return constant values | Global `Nil` and `Max` variables can be modified by any package |
| Error handling | All errors checked and propagated up the call chain | Some errors silently ignored (e.g. rand.Read, hash.Write) |

This package prioritizes explicit configuration over convenience, making it easier to test and avoiding hidden global state.


## Benchmarks

```
go test ./...  -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/soypat/uuid
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkParse-12                 499062              2322 ns/op               0 B/op          0 allocs/op
BenchmarkParseBytes-12            502353              2160 ns/op               0 B/op          0 allocs/op
BenchmarkAppendText-12            733390              1611 ns/op               0 B/op          0 allocs/op
BenchmarkEncode36-12             8865866               133.8 ns/op             0 B/op          0 allocs/op
BenchmarkEncode32-12             9611780               124.5 ns/op             0 B/op          0 allocs/op
```