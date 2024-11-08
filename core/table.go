package core

// Table ...
type Table struct {
	Name    string
	As      string
	Note    string
	Columns []Column
	Indexes []Index

	Settings TableSettings
}

type TableSettings struct {
	HeaderColor string
}
