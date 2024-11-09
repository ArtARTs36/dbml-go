package parser

import (
	"fmt"

	"github.com/artarts36/dbml-go/core"
	"github.com/artarts36/dbml-go/token"
)

func (p *Parser) parseEnum() (*core.Enum, error) {
	enum := &core.Enum{}
	p.next()

	if !token.IsIdent(p.token) && p.token != token.DSTRING {
		return nil, fmt.Errorf("enum name is invalid: %s", p.lit)
	}
	enum.Name = p.lit
	p.next()
	if p.token != token.LBRACE {
		return nil, p.expect("{")
	}
	p.next()

	for token.IsIdent(p.token) || p.token == token.DSTRING {
		enumValue := core.EnumValue{
			Name: p.lit,
		}
		p.next()
		if p.token == token.LBRACK {
			// handle [Note: ...]
			p.next()
			if p.token == token.NOTE {
				note, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("note: 'string'")
				}
				enumValue.Note = note
				p.next()
			}
			if p.token != token.RBRACK {
				return nil, p.expect("]")
			}
			p.next()
		}
		enum.Values = append(enum.Values, enumValue)
	}

	if p.token != token.RBRACE {
		return nil, p.expect("}")
	}
	return enum, nil
}
