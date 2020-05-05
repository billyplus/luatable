package luatable

type parser struct {
	scanner scanner
}

func (p *parser) init(src []byte) {
	p.scanner.Init(src)
}

// func (p *parser) parseElement() expr {

// }

// func (p *parser) parse() {

// }
