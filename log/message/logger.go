package message

import (
	"github.com/mandelsoft/kubecrtutils/log/message/base"
)

type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})

	Enabled(lvl int) bool
}

type logger struct {
	base base.BaseLogger
}

func New(base base.BaseLogger) Logger {
	return logger{base: base}
}

func (l logger) Info(args ...interface{}) {
	Info(l.base, args...)
}

func (l logger) Debug(args ...interface{}) {
	Debug(l.base, args...)
}

func (l logger) Trace(args ...interface{}) {
	Trace(l.base, args...)
}

func (l logger) Enabled(lvl int) bool {
	return l.base.Enabled(lvl)
}
