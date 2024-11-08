package parser

import (
	"context"
	"io"

	"github.com/artarts36/dbml-go/core"
	"github.com/artarts36/dbml-go/scanner"
)

func Parse(ctx context.Context, spec io.Reader) (*core.DBML, error) {
	return NewParser(scanner.NewScanner(spec)).Parse(ctx)
}

func ParseWithDebug(ctx context.Context, spec io.Reader, logger Logger) (*core.DBML, error) {
	return NewParserWithLogger(scanner.NewScanner(spec), logger).Parse(ctx)
}
