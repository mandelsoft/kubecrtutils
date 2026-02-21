package base

type InfoLogger interface {
	Info(string, ...interface{})
	Enabled(lvl int) bool
}

type DebugLogger interface {
	Debug(string, ...interface{})
	Enabled(lvl int) bool
}

type TraceLogger interface {
	Trace(string, ...interface{})
	Enabled(lvl int) bool
}

type BaseLogger interface {
	InfoLogger
	DebugLogger
	TraceLogger
}
