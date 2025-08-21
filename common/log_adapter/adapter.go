package log_adapter

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/yc-alpha/logger"
	"github.com/yc-alpha/logger/backend"
	"github.com/yc-alpha/variant"
)

type Logger struct {
	logger logger.Logger
}

func NewAdapter() log.Logger {
	return &Logger{
		logger: logger.NewLogger(
			logger.WithLevel(logger.DebugLevel),
			logger.WithBackends(logger.AnyLevel, backend.OSBackend().Build()),
			logger.WithSeparator(logger.AnyLevel, "    "),
			logger.WithFields(logger.AnyLevel,
				logger.DatetimeField(time.DateTime).Key("datetime"),
			),
			logger.WithFields(logger.DebugLevel|logger.InfoLevel,
				logger.LevelField().Key("level").Upper().Prefix("[").Suffix("]").Color(logger.Green),
			),
			logger.WithFields(logger.WarnLevel,
				logger.LevelField().Key("level").Upper().Prefix("[").Suffix("]").Color(logger.Yellow),
			),
			logger.WithFields(logger.ErrorLevel|logger.FatalLevel|logger.PanicLevel,
				logger.LevelField().Key("level").Upper().Prefix("[").Suffix("]").Color(logger.Red),
			),
			logger.WithFields(logger.AnyLevel,
				logger.MessageField().Key("msg"),
				logger.CallerField(true, true, 3).Key("caller"),
			),
			logger.WithEncoders(logger.AnyLevel, logger.PlainEncoder),
		),
	}

}

func (l *Logger) Log(level log.Level, keyvals ...any) error {
	msg := ""
	fields := []logger.FieldBuilder{}

	// 将 keyvals 转为字符串
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := ""
		if i+1 < len(keyvals) {
			v = fmt.Sprintf("%v", keyvals[i+1])
		}
		if k == "msg" || k == "message" {
			msg = v
		} else {
			fields = append(fields, logger.F(variant.New(k).ToString(), v))
		}
	}

	switch level {
	case log.LevelDebug:
		l.logger.Debugs(msg, fields...)
	case log.LevelInfo:
		l.logger.Infos(msg, fields...)
	case log.LevelWarn:
		l.logger.Warns(msg, fields...)
	case log.LevelError:
		l.logger.Errors(msg, fields...)
	case log.LevelFatal:
		l.logger.Fatals(msg, fields...)
	default:
		l.logger.Infos(msg, fields...)
	}
	return nil
}
