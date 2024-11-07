package core

type ColumnDefaultType int

const (
	ColumnDefaultTypeUnknown ColumnDefaultType = iota
	ColumnDefaultTypeString
	ColumnDefaultTypeExpression
	ColumnDefaultTypeNumber
)

type ColumnDefault struct {
	Raw   string
	Value interface{}
	Type  ColumnDefaultType
}
