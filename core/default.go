package core

type ColumnDefaultType int

const (
	ColumnDefaultTypeUnknown ColumnDefaultType = iota
	ColumnDefaultTypeNumber
	ColumnDefaultTypeString
	ColumnDefaultTypeExpression
	ColumnDefaultTypeBoolean
)

type ColumnDefault struct {
	Raw   string
	Value interface{}
	Type  ColumnDefaultType
}
