package constraints

type _always struct{}

func Always() Constraint {
	return _always{}
}

func (a _always) Match(ctx Context) (Activation, error) {
	return Yes, nil
}
