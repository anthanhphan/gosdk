# jcodec - Intelligent JSON Codec

An architecture-aware JSON marshaling package that automatically selects the optimal JSON library based on CPU architecture.

## Features

- **Automatic Selection**: Uses Sonic on AMD64/x86_64, goccy/go-json on ARM64/others
- **High Performance**: 1.5-5x faster than standard library (see benchmarks below)
- **Drop-in Replacement**: Compatible with `encoding/json` API
- **Zero Configuration**: Works out of the box
- **Thread-Safe**: Concurrent operations validated with race detector
- **Error Context**: Enhanced error messages with engine context for easier debugging

## Installation

```bash
go get github.com/anthanhphan/gosdk/jcodec
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/anthanhphan/gosdk/jcodec"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
}

func main() {
    user := User{ID: 1, Name: "John"}

    // Marshal
    data, err := jcodec.Marshal(user)
    if err != nil {
        panic(err)
    }

    // Unmarshal
    var unmarshaled User
    err = jcodec.Unmarshal(data, &unmarshaled)
    if err != nil {
        panic(err)
    }
    
    // Pretty-print JSON
    prettyData, err := jcodec.MarshalIndent(user, "", "  ")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(prettyData))
    
    // Validate JSON
    if jcodec.Valid(data) {
        fmt.Println("Valid JSON!")
    }
}
```

## Migration

Simply replace your import:

```go
// Before
import "encoding/json"
data, err := json.Marshal(user)
err = json.Unmarshal(data, &user)

// After
import "github.com/anthanhphan/gosdk/jcodec"
data, err := jcodec.Marshal(user)
err = jcodec.Unmarshal(data, &user)
```

## API

### Marshal

```go
func Marshal(v interface{}) ([]byte, error)
```

Converts a Go value to JSON bytes.

### Unmarshal

```go
func Unmarshal(data []byte, v interface{}) error
```

Converts JSON bytes to a Go value.

### MarshalIndent

```go
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
```

Converts a Go value to pretty-printed JSON bytes using indentation.

### Valid

```go
func Valid(data []byte) bool
```

Reports whether data is valid JSON without unmarshaling.

### Utility Functions

`jcodec` provides standard utility functions for JSON manipulation, fully compatible with `encoding/json`.

#### `Compact(dst *bytes.Buffer, src []byte) error`
Appends to `dst` the JSON-encoded `src` with insignificant space characters elided.

#### `Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error`
Appends to `dst` an indented form of the JSON-encoded `src`.

#### `HTMLEscape(dst *bytes.Buffer, src []byte)`
Appends to `dst` the JSON-encoded `src` with characters inside string literals changed to be safe for embedding in HTML `<script>` tags.

### Streaming API

#### `NewEncoder(w io.Writer) Encoder`

Returns a new encoder that writes to `w`.

```go
enc := jcodec.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
err := enc.Encode(data)
```

#### `NewDecoder(r io.Reader) Decoder`

Returns a new decoder that reads from `r`.

```go
dec := jcodec.NewDecoder(strings.NewReader(jsonStream))
for {
    var v MyStruct
    if err := dec.Decode(&v); err == io.EOF {
        break
    } else if err != nil {
        log.Fatal(err)
    }
    // process v
}
```

### RawMessage

`RawMessage` is a raw encoded JSON value. It implements `Marshaler` and `Unmarshaler` and can be used to delay JSON decoding or precompute a JSON encoding.

```go
type MyStruct struct {
    Type  string
    Value jcodec.RawMessage
}
```

## Performance

Benchmarks run on Apple M3 (ARM64) using goccy/go-json engine:

### Marshal Performance

| Operation | jcodec | stdlib | Speedup |
|-----------|--------|--------|---------|
| Simple Struct | 163.6 ns/op | 295.5 ns/op | **1.8x faster** |
| Complex Struct | 323.2 ns/op | - | - |
| Large Struct | 3.8 μs/op | - | - |

### Unmarshal Performance

| Operation | jcodec | stdlib | Speedup |
|-----------|--------|--------|---------|
| Simple Struct | 171.9 ns/op | 821.2 ns/op | **4.8x faster** |
| Complex Struct | 415.1 ns/op | - | - |
| Large Struct | 5.1 μs/op | - | - |

### MarshalIndent Performance

| Operation | jcodec | stdlib | Speedup |
|-----------|--------|--------|---------|
| Simple Struct | 250.5 ns/op | 1559 ns/op | **6.2x faster** |
| Complex Struct | 451.6 ns/op | - | - |

### Parallel Performance

Parallel operations show excellent scalability:
- Marshal: **80.55 ns/op** in parallel
- Unmarshal: **64.64 ns/op** in parallel

**Note**: On AMD64/x86_64 platforms using Sonic engine, you can expect even better performance (3-10x faster than standard library).

## Concurrency

jcodec is fully thread-safe and optimized for concurrent use:
- Engine initialization uses `sync.Once` for safe lazy loading
- Validated with Go race detector
- Stress-tested with 100 goroutines × 1000 operations

## Error Handling

Errors include engine context for easier debugging:

```go
data, err := jcodec.Marshal(invalidValue)
// Error: "goccy engine: json: unsupported type: chan int"
```

## License

Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>
