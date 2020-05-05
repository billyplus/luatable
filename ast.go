package luatable

type node interface {
	begin() Pos // position of first character belonging to node
	end() Pos   // position of first character immediately after the node
}

// all expression node implement expr interface
type expr interface {
	node
	exprNode()
}

//------------------------------------
// comment
type comment struct {
	pos  Pos
	text string
}

// implement node
func (c *comment) begin() Pos { return c.pos }
func (c *comment) end() Pos   { return Pos(int(c.pos) + len(c.text)) }

//------------------------------------
// expression and types

type(
	arrayType struct {
		lBracket Pos
	}
)