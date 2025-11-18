# jcodec - Intelligent JSON Codec

An architecture-aware JSON marshaling package that automatically selects the optimal JSON library based on CPU architecture.

## Features

- **Automatic Selection**: Uses Sonic on AMD64/x86_64, goccy/go-json on ARM64/others
- **High Performance**: 3-10x faster on AMD64, 1.4-8x faster on ARM64
- **Drop-in Replacement**: Compatible with `encoding/json` API
- **Zero Configuration**: Works out of the box

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
}
```

## Migration

Simply replace your import:

```go
// Before
import "encoding/json"
data, err := json.Marshal(user)

// After
import "github.com/anthanhphan/gosdk/jcodec"
data, err := jcodec.Marshal(user)
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

## Performance

| Architecture | Library      | Performance Gain              |
| ------------ | ------------ | ----------------------------- |
| AMD64/x86_64 | Sonic        | 3-10x faster than standard    |
| ARM64/Others | goccy/go-json| 1.4-8x faster than standard   |

## License

Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>
