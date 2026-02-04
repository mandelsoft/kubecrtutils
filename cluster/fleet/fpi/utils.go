package fpi

import (
	"strings"
)

type Composer struct {
	name string
}

func NewComposer(name string) Composer {
	return Composer{name}
}

func (c *Composer) GetName() string {
	return c.name
}

func (c *Composer) Compose(name string) string {
	if name == "" {
		return c.name
	}
	return c.name + "#" + name
}

func (c *Composer) Strip(name string) string {
	_, n := c.Split(name)
	return n
}

func (c *Composer) Match(name string) bool {
	if c.name == name {
		return true
	}
	b, _ := c.Split(name)
	return b == c.name
}

func (c *Composer) Split(name string) (string, string) {
	i := strings.Index(name, "#")
	if i < 0 {
		return "", name
	}
	return name[:i], name[i+1:]
}
