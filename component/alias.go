package component

type _alias struct {
	Component
	name string
}

func NewAlias(name string, c Component) Component {
	if name == c.GetName() {
		return c
	}
	return &_alias{
		Component: c,
		name:      name,
	}
}

func (a *_alias) GetName() string {
	return a.name
}

func (a *_alias) Unwrap() Component {
	return a.Component
}
