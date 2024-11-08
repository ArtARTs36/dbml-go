package parser

import (
	"context"
	"github.com/artarts36/dbml-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/artarts36/dbml-go/scanner"
)

func p(str string) *Parser {
	r := strings.NewReader(str)
	s := scanner.NewScanner(r)
	parser := NewParser(s)
	return parser
}

func TestIllegalSyntax(t *testing.T) {
	parser := p(`Project test { abc , xyz`)
	_, err := parser.Parse(context.Background())
	require.Error(t, err)
}

func TestParseSimple(t *testing.T) {
	parser := p(`
	Project test {
		note: 'just test note'
	}
	table users {
		id int [pk, note: 'just test column note']
	}
	table float_number {
		
	}
	`)
	dbml, err := parser.Parse(context.Background())
	if err != nil {
		t.Fail()
	}
	if dbml.Project.Name != "test" {
		t.Fail()
	}

	if dbml.Project.Note != "just test note" {
		t.Fail()
	}

	usersTable := dbml.Tables[0]
	if usersTable.Name != "users" {
		t.Fail()
	}
	idColumn := usersTable.Columns[0]
	if idColumn.Name != "id" {
		t.Fail()
	}
	if !idColumn.Settings.PK {
		t.Fail()
	}
	if idColumn.Settings.Note != "just test column note" {
		t.Fail()
	}
}

func TestParseTableName(t *testing.T) {
	parser := p(`
	Table int {
		id int
	}
	`)
	dbml, err := parser.Parse(context.Background())
	if err != nil {
		t.Fail()
	}
	table := dbml.Tables[0]
	if table.Name != "int" {
		t.Fatalf("table name should be 'int'")
	}
}

func TestParseTableWithType(t *testing.T) {
	parser := p(`
	Table int {
		type int
	}
	`)
	dbml, err := parser.Parse(context.Background())
	if err != nil {
		t.Fail()
	}
	table := dbml.Tables[0]
	if table.Columns[0].Name != "type" {
		t.Fatalf("column name should be 'type'")
	}
}

func TestParseTableWithNoteColumn(t *testing.T) {
	parser := p(`
	Table int {
		note int
	}
	`)
	dbml, err := parser.Parse(context.Background())

	// t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	table := dbml.Tables[0]
	if table.Columns[0].Name != "note" {
		t.Fatalf("column name should be 'note'")
	}
}

func TestAllowKeywordsAsTable(t *testing.T) {
	parser := p(`
	Table project {
		note int
	}
	`)
	dbml, err := parser.Parse(context.Background())

	// t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	table := dbml.Tables[0]
	if table.Name != "project" {
		t.Fatalf("table name should be 'project'")
	}
}

func TestAllowKeywordsAsEnum(t *testing.T) {
	parser := p(`
	Enum project {
		key
	}
	`)
	dbml, err := parser.Parse(context.Background())

	// t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	enum := dbml.Enums[0]
	if enum.Name != "project" {
		t.Fatalf("enum name should be 'project'")
	}

	if enum.Values[0].Name != "key" {
		t.Fatalf("enum value should be 'key'")
	}
}

func TestParser_Parse_Column_Settings_Default(t *testing.T) {

	cases := []struct {
		Title    string
		Spec     string
		Expected core.ColumnDefault
	}{
		{
			Title: "parse string value with single quotes",
			Spec: `
	Table user {
		name varchar [default: 'test']
	}
`,
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeString,
				Raw:   "test",
				Value: "test",
			},
		},
		{
			Title: "parse string value with double quotes",
			Spec: `
	Table user {
		name varchar [default: "test"]
	}
`,
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeString,
				Raw:   "test",
				Value: "test",
			},
		},
		{
			Title: "parse int value",
			Spec: `
	Table user {
		name varchar [default: 123]
	}
`,
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeNumber,
				Raw:   "123",
				Value: 123,
			},
		},
		{
			Title: "parse int value",
			Spec: `
	Table user {
		name varchar [default: 123]
	}
`,
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeNumber,
				Raw:   "123",
				Value: 123,
			},
		},
		{
			Title: "parse float value",
			Spec: `
	Table user {
		name varchar [default: 123.456]
	}
`,
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeNumber,
				Raw:   "123.456",
				Value: 123.456,
			},
		},
		{
			Title: "parse expression value",
			Spec:  "Table user { name varchar [default: `now()`]}",
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeExpression,
				Raw:   "now()",
				Value: "now()",
			},
		},
		{
			Title: "parse false value",
			Spec:  "Table user { name varchar [default: false]}",
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeBoolean,
				Raw:   "false",
				Value: false,
			},
		},
		{
			Title: "parse true value",
			Spec:  "Table user { name varchar [default: true]}",
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeBoolean,
				Raw:   "true",
				Value: true,
			},
		},
		{
			Title: "parse null value",
			Spec:  "Table user { name varchar [default: null]}",
			Expected: core.ColumnDefault{
				Type:  core.ColumnDefaultTypeBoolean,
				Raw:   "null",
				Value: nil,
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.Title, func(t *testing.T) {
			dbml, err := p(tCase.Spec).Parse(context.Background())
			require.NoError(t, err)

			assert.Equal(t, tCase.Expected, dbml.Tables[0].Columns[0].Settings.Default)
		})
	}
}
