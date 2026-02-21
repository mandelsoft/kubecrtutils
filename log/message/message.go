package message

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/log/message/base"
	"github.com/mandelsoft/logging"
)

type MessageProvider interface {
	logging.KeyValueProvider
	Message() string
}

var KeyValue = logging.KeyValue

func Values(args ...interface{}) []interface{} {
	return args
}

func Info(logger base.InfoLogger, args ...interface{}) {
	if logger.Enabled(logging.InfoLevel) {
		log(logger.Info, args...)
	}
}

func Debug(logger base.DebugLogger, args ...interface{}) {
	if logger.Enabled(logging.DebugLevel) {
		log(logger.Debug, args...)
	}
}

func Trace(logger base.TraceLogger, args ...interface{}) {
	if logger.Enabled(logging.TraceLevel) {
		log(logger.Trace, args...)
	}
}

func log(logger func(string, ...interface{}), args ...any) {
	msg := ""
	var values []interface{}

	sep := ""
	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch v := arg.(type) {
		case string:
			msg += v
			sep = ""
		case MessageProvider:
			msg += sep + v.Message()
			v.NormalizeTo(&values)
			sep = "  "
		case logging.KeyValueProvider:
			l := len(values)
			v.NormalizeTo(&values)
			for i := l; i < len(values); i += 2 {
				msg += sep + fmt.Sprintf("{{%s}}", values[i])
				sep = " "
			}
		case []interface{}:
			values = append(values, v...)
		default:
			msg += fmt.Sprintf("%s%v", sep, v)
		}
	}
	logger(msg, values...)
}
