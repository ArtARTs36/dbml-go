package parser

import (
	"fmt"

	"github.com/artarts36/dbml-go/core"
	"github.com/artarts36/dbml-go/token"
)

func (p *Parser) parseTableSettings() (*core.TableSettings, error) {
	tableSetting := &core.TableSettings{}
	commaAllowed := false

	for {
		p.next()
		switch p.token {
		case token.HEADERCOLOR:
			p.next()
			if p.token != token.COLON {
				return nil, p.expect(":")
			}
			p.next()
			if p.lit != "#" {
				return nil, p.expect("#")
			}
			p.next()

			if p.token != token.IDENT {
				return nil, p.expect("color string")
			}

			tableSetting.HeaderColor = fmt.Sprintf("#%s", p.lit)
		case token.COMMA:
			if !commaAllowed {
				return nil, p.expect("pk | primary key | unique")
			}
		case token.RBRACK:
			return tableSetting, nil
		default:
			return nil, p.expect("pk, primary key, unique")
		}
		commaAllowed = !commaAllowed
	}
}
