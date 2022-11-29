package logger

import "context"

type Valuer func(ctx context.Context) interface{}

func bindValues(ctx context.Context, keyvals []interface{}) {
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(Valuer); ok {
			keyvals[i] = v(ctx)
		}
	}
}

type ILogger interface {
	WithKeyValue(kvs ...interface{}) ILogger

	WithContext(ctx context.Context) ILogger

	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
	Debugw(ctx context.Context, format string, a ...interface{})

	Info(a ...interface{})
	Infof(format string, a ...interface{})
	Infow(ctx context.Context, format string, a ...interface{})

	Warn(a ...interface{})
	Warnf(format string, a ...interface{})
	Warnw(ctx context.Context, format string, a ...interface{})

	Error(a ...interface{})
	Errorf(format string, a ...interface{})
	Errorw(ctx context.Context, format string, a ...interface{})
}
