package constraints

type _activated []string

func Activated(names ...string) Constraint {
	return _activated(names)
}

func (a _activated) Match(ctx *Context) (Activation, error) {
	if len(a) == 0 {
		return NoOpinion, nil
	}
	for _, name := range a {
		if !ctx.Has(name) {
			return No, nil
		}
	}
	return Yes, nil
}
