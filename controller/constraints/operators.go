package constraints

import (
	"fmt"
)

type _and []Constraint

func (o _and) Match(ctx *Context) (Activation, error) {
	var rerr error

	r := NoOpinion
	for _, c := range o {
		ok, err := c.Match(ctx)
		r = r.And(ok)
		if r == No {
			rerr = err
			break
		}
		if err != nil {
			rerr = join(rerr, err, "AND")
		}
	}
	return r, rerr
}

func And(constraints ...Constraint) Constraint {
	return _and(constraints)
}

type _or []Constraint

func (o _or) Match(ctx *Context) (Activation, error) {
	var rerr error
	found := false
	r := NoOpinion
	for _, c := range o {
		ok, err := c.Match(ctx)
		r = r.Or(ok)
		if ok != No {
			if err == nil {
				found = true
			} else {
				rerr = join(rerr, err, "OR")
			}
		}
	}
	if found {
		return r, nil
	}
	return r, rerr
}

func Or(constraints ...Constraint) Constraint {
	return _or(constraints)
}

type _not struct {
	cond Constraint
}

func Not(constraint Constraint) Constraint {
	return _not{constraint}
}

func (o _not) Match(ctx *Context) (Activation, error) {
	a, err := o.cond.Match(ctx)
	return a.Not(), err
}

func join(a, b error, op string) error {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return fmt.Errorf("%s %s %s", a, op, b)
}
