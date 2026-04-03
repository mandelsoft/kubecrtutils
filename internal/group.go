package internal

import (
	"fmt"
	"maps"

	"github.com/mandelsoft/goutils/errors"
)

type ErrorProvider interface {
	GetError() error
}

type ErrorContainer struct {
	lock    Mutex
	errlist *errors.ErrorList
}

func NewErrorContainer(name string) *ErrorContainer {
	return &ErrorContainer{
		errlist: errors.ErrListf(name),
	}
}

func (c *ErrorContainer) AddError(a any, ctx ...any) error {
	defer c.lock.Lock()()

	if a == nil {
		return nil
	}
	var err error
	switch v := a.(type) {
	case error:
		err = v
	case ErrorProvider:
		err = v.GetError()
	case string:
		err = fmt.Errorf("%s", v)
	}
	err = errors.Wrap(err, ctx...)
	c.errlist.Add(err)
	return err
}

func (c *ErrorContainer) GetError() error {
	defer c.lock.Lock()()
	return c.errlist.Result()
}

////////////////////////////////////////////////////////////////////////////////

type Group[T Named] interface {
	Get(name string) T
	Add(elem ...T) error
	Elements(yield func(string, T) bool)
	Len() int

	AddError(elem any, ctx ...any) error
	ErrorProvider
}

type _group[T Named] struct {
	Mutex
	ErrorContainer
	typename string
	elements map[string]T
}

func NewGroup[T Named](name string) Group[T] {
	return &_group[T]{typename: name, elements: make(map[string]T), ErrorContainer: *NewErrorContainer(fmt.Sprintf("%s set", name))}
}

func newGroup[T Named](name string) _group[T] {
	return _group[T]{typename: name, elements: make(map[string]T), ErrorContainer: *NewErrorContainer(fmt.Sprintf("%s set", name))}
}

func (c *_group[T]) GetName() string {
	return c.typename
}

func (c *_group[T]) Len() int {
	defer c.Lock()()
	return len(c.elements)
}

func (c *_group[T]) Add(elem ...T) error {
	defer c.Lock()()
	return c.add(elem...)
}

func (c *_group[T]) add(elem ...T) error {
	var err error

	for _, e := range elem {
		if _, ok := c.elements[e.GetName()]; ok {
			err = c.AddError(fmt.Errorf("%s %q already exists", c.typename, e.GetName()))
		} else {
			c.elements[e.GetName()] = e
			err = c.AddError(e, fmt.Sprintf("%s %s", c.typename, e.GetName()))
		}
	}
	return err
}

func (c *_group[T]) Get(name string) T {
	defer c.Lock()()
	return c.elements[name]
}

func (c *_group[T]) Elements(yield func(string, T) bool) {
	c.Mutex.Lock()
	m := maps.Clone(c.elements)
	c.Mutex.Unlock()
	for n, elem := range m {
		if !yield(n, elem) {
			return
		}
	}
}
