# DBML parser for Go

DBML-go is a Go parser for [DBML](https://www.dbml.org) syntax.

This package is fork from https://github.com/duythinht/dbml-go

Key different points:
* Up min GO version to 1.18
* Changed default `column.null` value to false
* Removed model generator
* Added parsing default value with determine type (string, number, expression)
* Removed call to `os` package
* Improved logging
* Added parsing table settings (headercolor)

## Installation

Go get

```bash
go get github.com/artarts36/dbml-go/...
```

## Usage

```go
package main

import (
	"context"
	
	"os"
	
	"github.com/artarts36/dbml-go/parser"
)

func main() {
	f, _ := os.OpenFile("spec.dbml", os.O_RDONLY, 0755)
	
	parser.Parse(context.Background(), f)
}
```
