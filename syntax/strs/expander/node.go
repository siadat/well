package expander

type ExecWrd struct {
	Lit string
}

// Whs is a whitespace node
type ExecWhs struct {
	Lit string
}

// Var is a var node
type ExecVar struct {
	Lit string
}

type ExecNode interface {
	node()
	Value() string
}

func (ExecWrd) node() {}
func (ExecVar) node() {}
func (ExecWhs) node() {}

func (e ExecWrd) Value() string { return e.Lit }
func (e ExecVar) Value() string { return e.Lit }
func (e ExecWhs) Value() string { return e.Lit }
