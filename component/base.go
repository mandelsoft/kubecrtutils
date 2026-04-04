package component

type Base struct {
	def  Definition
	self Component
}

func NewBase(def Definition, self Component) *Base {
	return &Base{def, self}
}

func (c *Base) GetName() string {
	return c.def.GetName()
}

func (c *Base) GetDefinition() Definition {
	return c.def
}

func (c *Base) GetEffective() Component {
	return c.self
}
