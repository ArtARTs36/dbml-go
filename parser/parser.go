package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/artarts36/dbml-go/core"
	"github.com/artarts36/dbml-go/scanner"
	"github.com/artarts36/dbml-go/token"
)

// Parser declaration.
type Parser struct {
	s *scanner.Scanner

	// current token & literal
	token token.Token
	lit   string

	logger Logger
}

// NewParser ...
func NewParser(s *scanner.Scanner) *Parser {
	return NewParserWithLogger(s, NoopLogger)
}

func NewParserWithLogger(s *scanner.Scanner, logger Logger) *Parser {
	return &Parser{
		s:      s,
		token:  token.ILLEGAL,
		lit:    "",
		logger: logger,
	}
}

// Parse ...
func (p *Parser) Parse(ctx context.Context) (*core.DBML, error) {
	dbml := &core.DBML{}
	for {
		p.next()
		switch p.token {
		case token.PROJECT:
			project, err := p.parseProject()
			if err != nil {
				return nil, err
			}
			p.debug(ctx, "found project", map[string]any{"project": project})
			dbml.Project = *project
		case token.TABLE:
			table, err := p.parseTable(ctx)
			if err != nil {
				return nil, err
			}
			p.debug(ctx, "found table", map[string]any{"table": table})

			// TODO:
			// * register table to tables map, for check ref
			dbml.Tables = append(dbml.Tables, *table)

		case token.REF:
			ref, err := p.parseRefs()
			if err != nil {
				return nil, err
			}
			p.debug(ctx, "found refs", map[string]any{
				"ref": ref,
			})

			// TODO:
			// * Check refs is valid or not (by tables map)
			dbml.Refs = append(dbml.Refs, *ref)

		case token.ENUM:
			enum, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			p.debug(ctx, "found enum", map[string]any{
				"enum": enum,
			})
			dbml.Enums = append(dbml.Enums, *enum)

		case token.TABLEGROUP:
			tableGroup, err := p.parseTableGroup()
			if err != nil {
				return nil, err
			}
			p.debug(ctx, "found table group", map[string]any{
				"table_group": tableGroup,
			})
			dbml.TableGroups = append(dbml.TableGroups, *tableGroup)
		case token.EOF:
			return dbml, nil
		default:
			p.debug(ctx, "got unexpected token", map[string]any{
				"token": p.token.String(),
				"lit":   p.lit,
			})
			return nil, p.expect("Project, Ref, Table, Enum, TableGroup")
		}
	}
}

func (p *Parser) parseTableGroup() (*core.TableGroup, error) {
	tableGroup := &core.TableGroup{}
	p.next()
	if p.token != token.IDENT && p.token != token.DSTRING {
		return nil, fmt.Errorf("TableGroup name is invalid: %s", p.lit)
	}
	tableGroup.Name = p.lit
	p.next()
	if p.token != token.LBRACE {
		return nil, p.expect("{")
	}
	p.next()

	for p.token == token.IDENT || p.token == token.DSTRING {
		tableGroup.Members = append(tableGroup.Members, p.lit)
		p.next()
	}
	if p.token != token.RBRACE {
		return nil, p.expect("}")
	}
	return tableGroup, nil
}

func (p *Parser) parseRefs() (*core.Ref, error) {
	ref := &core.Ref{}
	p.next()

	// Handle for Ref <optional_name>...
	if p.token == token.IDENT {
		ref.Name = p.lit
		p.next()
	}

	// Ref: from > to
	if p.token == token.COLON {
		p.next()
		rel, err := p.parseRelationship()
		if err != nil {
			return nil, err
		}
		ref.Relationships = append(ref.Relationships, *rel)
		return ref, nil
	}

	if p.token == token.LBRACE {
		p.next()

		for {
			if p.token == token.RBRACE {
				return ref, nil
			} else if p.token == token.IDENT || p.token == token.DSTRING {
				rel, err := p.parseRelationship()
				if err != nil {
					return nil, err
				}
				ref.Relationships = append(ref.Relationships, *rel)
			} else {
				return nil, p.expect("Ref: { from > to }")
			}
			p.next()
		}
	}

	return nil, p.expect("Ref: | Refs {}")
}

func (p *Parser) parseRelationship() (*core.Relationship, error) {
	rel := &core.Relationship{}
	if p.token != token.IDENT && p.token != token.DSTRING {
		return nil, p.expect("(rel from) table.column_name")
	}

	rel.From = p.lit

	p.next()
	if reltype, ok := core.RelationshipMap[p.token]; ok {
		rel.Type = reltype
	} else {
		return nil, p.expect("> | < | -")
	}

	p.next()
	if p.token != token.IDENT {
		return nil, p.expect("(rel to) table.column_name")
	}
	rel.To = p.lit
	return rel, nil
}

func (p *Parser) parseTable(ctx context.Context) (*core.Table, error) {
	table := &core.Table{}
	p.next()
	switch p.token {
	case token.IDENT, token.DSTRING:
		// pass
	default:
		if m, _ := regexp.MatchString("^[a-zA-Z1-9]+$", p.lit); !m {
			return nil, fmt.Errorf("table name is invalid: %s", p.lit)
		}
	}
	table.Name = p.lit

	p.next()

	switch p.token {
	case token.AS:
		// handle as
		p.next()
		switch p.token {
		case token.STRING, token.IDENT:
			table.As = p.lit
		default:
			return nil, p.expect("as NAME")
		}
		p.next()
		fallthrough
	case token.LBRACK:
		// handle parseColumn
		tableSetting, err := p.parseTableSettings()
		if err != nil {
			return nil, fmt.Errorf("parse table settings: %w", err)
		}
		p.next() // remove ']'
		table.Settings = *tableSetting
		fallthrough
	case token.LBRACE:
		p.next()
		for {
			switch p.token {
			case token.INDEXES:
				indexes, err := p.parseIndexes(ctx)
				if err != nil {
					return nil, err
				}
				table.Indexes = indexes
			case token.RBRACE:
				return table, nil
			default:
				columnName := p.lit
				currentToken := p.token
				p.next()
				if currentToken == token.NOTE && p.token == token.COLON {
					note, err := p.parseString()
					if err != nil {
						return nil, err
					}
					table.Note = note
					p.next()
				} else {
					column, err := p.parseColumn(ctx, columnName)
					if err != nil {
						return nil, err
					}
					table.Columns = append(table.Columns, *column)
				}
			}
		}
	default:
		return nil, p.expect("{")
	}
}

func (p *Parser) parseIndexes(ctx context.Context) ([]core.Index, error) {
	indexes := []core.Index{}

	p.next()
	if p.token != token.LBRACE {
		return nil, p.expect("{")
	}

	p.next()
	for {
		if p.token == token.RBRACE {
			p.next() // pop }
			return indexes, nil
		}
		// parse an Index
		index, err := p.parseIndex()
		if err != nil {
			return nil, err
		}
		p.debug(ctx, "found index", map[string]any{
			"index": index,
		})
		indexes = append(indexes, *index)
	}
}

func (p *Parser) parseIndex() (*core.Index, error) {
	index := &core.Index{}

	if p.token == token.LPAREN {
		p.next()
		for token.IsIdent(p.token) {
			index.Fields = append(index.Fields, p.lit)
			p.next()
			if p.token == token.COMMA {
				p.next()
			}
		}
		if p.token != token.RPAREN {
			return nil, p.expect(")")
		}
	} else if token.IsIdent(p.token) {
		index.Fields = append(index.Fields, p.lit)
	} else {
		return nil, p.expect("field_name")
	}

	p.next()

	if p.token == token.LBRACK {
		// Handle index setting [settings...]
		commaAllowed := false

		for {
			p.next()
			switch {
			case p.token == token.IDENT && strings.ToLower(p.lit) == "name":
				name, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("name: 'index_name'")
				}
				index.Settings.Name = name
			case p.token == token.NOTE:
				note, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("note: 'index note'")
				}
				index.Settings.Note = note
			case p.token == token.PK:
				index.Settings.PK = true
			case p.token == token.UNIQUE:
				index.Settings.Unique = true
			case p.token == token.TYPE:
				p.next()
				if p.token != token.COLON {
					return nil, p.expect(":")
				}
				p.next()
				if p.token != token.IDENT || (p.lit != "hash" && p.lit != "btree") {
					return nil, p.expect("hash|btree")
				}
				index.Settings.Type = p.lit
			case p.token == token.COMMA:
				if !commaAllowed {
					return nil, p.expect("[index settings...]")
				}
			case p.token == token.RBRACK:
				p.next()
				return index, nil
			default:
				return nil, p.expect("note|name|type|pk|unique")
			}
			commaAllowed = !commaAllowed
		}
	}

	return index, nil
}

func (p *Parser) parseColumn(ctx context.Context, name string) (*core.Column, error) {
	column := &core.Column{
		Name: name,
	}
	if p.token != token.IDENT {
		return nil, p.expect("int, varchar,...")
	}
	column.Type = p.lit
	p.next()

	// parse for type
	switch p.token {
	case token.LPAREN:
		p.next()
		if p.token != token.INT {
			return nil, p.expect("int")
		}
		column.Type = fmt.Sprintf("%s(%s)", column.Type, p.lit)
		p.next()
		if p.token != token.RPAREN {
			return nil, p.expect(token.RPAREN.String())
		}
		p.next()
		if p.token != token.LBRACK {
			break
		}
		fallthrough
	case token.LBRACK:
		// handle parseColumn
		columnSetting, err := p.parseColumnSettings()
		if err != nil {
			return nil, fmt.Errorf("parse column settings: %w", err)
		}
		p.next() // remove ']'
		column.Settings = *columnSetting
	}

	p.debug(ctx, "found column", map[string]any{
		"column": column,
	})
	return column, nil
}

func (p *Parser) parseColumnDefault() (*core.ColumnDefault, error) {
	colDef := &core.ColumnDefault{
		Raw:   p.lit,
		Value: p.lit,
		Type:  core.ColumnDefaultTypeUnknown,
	}

	switch p.token {
	case token.STRING, token.DSTRING:
		colDef.Type = core.ColumnDefaultTypeString
	case token.INT:
		intVal, err := strconv.Atoi(p.lit)
		if err != nil {
			return nil, p.expect(fmt.Sprintf("default int value: %s", err.Error()))
		}

		colDef.Value = intVal
		colDef.Type = core.ColumnDefaultTypeNumber
	case token.FLOAT:
		floatVal, err := strconv.ParseFloat(p.lit, 4)
		if err != nil {
			return nil, p.expect(fmt.Sprintf("default float value: %s", err.Error()))
		}

		colDef.Value = floatVal
		colDef.Type = core.ColumnDefaultTypeNumber
	case token.EXPR:
		colDef.Type = core.ColumnDefaultTypeExpression
	case token.IDENT:
		if p.lit == "true" {
			colDef.Type = core.ColumnDefaultTypeBoolean
			colDef.Value = true
		} else if p.lit == "false" {
			colDef.Type = core.ColumnDefaultTypeBoolean
			colDef.Value = false
		} else {
			return nil, p.expect("default value")
		}
	case token.NULL:
		colDef.Type = core.ColumnDefaultTypeBoolean
		colDef.Value = nil
	default:
		return nil, p.expect("default value")
	}

	return colDef, nil
}

func (p *Parser) parseColumnSettings() (*core.ColumnSetting, error) {
	columnSetting := &core.ColumnSetting{}
	commaAllowed := false

	for {
		p.next()
		switch p.token {
		case token.PK:
			columnSetting.PK = true
		case token.PRIMARY:
			p.next()
			if p.token != token.KEY {
				return nil, p.expect("KEY")
			}
			columnSetting.PK = true
		case token.REF:
			p.next()
			if p.token != token.COLON {
				return nil, p.expect(":")
			}
			p.next()
			if p.token != token.LSS && p.token != token.GTR && p.token != token.SUB {
				return nil, p.expect("< | > | -")
			}
			columnSetting.Ref.Type = core.RelationshipMap[p.token]
			p.next()
			if p.token != token.IDENT {
				return nil, p.expect("table.column_id")
			}
			columnSetting.Ref.To = p.lit
		case token.NOT:
			p.next()
			if p.token != token.NULL {
				return nil, p.expect("null")
			}
			columnSetting.Null = false
		case token.UNIQUE:
			columnSetting.Unique = true
		case token.INCREMENT:
			columnSetting.Increment = true
		case token.DEFAULT:
			p.next()
			if p.token != token.COLON {
				return nil, p.expect(":")
			}

			p.next()

			def, err := p.parseColumnDefault()
			if err != nil {
				return nil, err
			}

			columnSetting.Default = *def
		case token.NOTE:
			str, err := p.parseDescription()
			if err != nil {
				return nil, err
			}
			columnSetting.Note = str
		case token.COMMA:
			if !commaAllowed {
				return nil, p.expect("pk | primary key | unique")
			}
		case token.RBRACK:
			return columnSetting, nil
		default:
			return nil, p.expect("pk, primary key, unique")
		}
		commaAllowed = !commaAllowed
	}
}

func (p *Parser) parseProject() (*core.Project, error) {
	project := &core.Project{}
	p.next()
	if p.token != token.IDENT && p.token != token.DSTRING {
		return nil, p.expect("project_name")
	}

	project.Name = p.lit
	p.next()

	if p.token != token.LBRACE {
		return nil, p.expect("{")
	}
	for {
		p.next()
		switch p.token {
		case token.IDENT:
			switch p.lit {
			case "database_type":
				str, err := p.parseDescription()
				if err != nil {
					return nil, err
				}
				project.DatabaseType = str
			default:
				return nil, p.expect("database_type")
			}
		case token.NOTE:
			note, err := p.parseDescription()
			if err != nil {
				return nil, err
			}
			project.Note = note
		case token.RBRACE:
			return project, nil
		default:
			return nil, fmt.Errorf("invalid token %s", p.lit)
		}
	}
}

func (p *Parser) parseString() (string, error) {
	p.next()
	switch p.token {
	case token.STRING, token.DSTRING, token.TSTRING:
		return p.lit, nil
	default:
		return "", p.expect("string, double quote string, triple string")
	}
}

func (p *Parser) parseDescription() (string, error) {
	p.next()
	if p.token != token.COLON {
		return "", p.expect(":")
	}
	return p.parseString()
}

func (p *Parser) next() {
	for {
		p.token, p.lit = p.s.Read()
		// p.debug("token:", p.token.String(), "lit:", p.lit)
		if p.token != token.COMMENT {
			break
		}
	}
}

func (p *Parser) expect(expected string) error {
	l, c := p.s.LineInfo()
	return fmt.Errorf("[%d:%d] invalid token '%s' determined as %s, expected: '%s'", l, c, p.lit, p.token, expected)
}

func (p *Parser) debug(ctx context.Context, msg string, params map[string]any) {
	p.logger(ctx, msg, params)
}
