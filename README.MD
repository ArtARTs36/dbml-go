# DBML parser for Go

DBML-go is a Go parser for [DBML](https://www.dbml.org) syntax.

This package is fork from https://github.com/artarts36/dbml-go

Key points:
* Changed default `column.null` value to false
* Removed model generator
* Parse default value with determine type (string, number, expression)

## Installation

Go get

```bash
go get github.com/artarts36/dbml-go/...
```

