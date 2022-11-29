package orm

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloeproject/toolbox/logger"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type GormLogger struct {
	conf glogger.Config
	log  logger.ILogger
}

func NewGormLogger(logger logger.ILogger, conf glogger.Config) *GormLogger {
	return &GormLogger{
		conf: conf,
		log:  logger,
	}
}

func (m *GormLogger) LogMode(level glogger.LogLevel) glogger.Interface {
	m.conf.LogLevel = level
	return m
}

func (m *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	m.log.WithContext(ctx).Infof(s, i)
}

func (m *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	m.log.WithContext(ctx).Warnf(s, i)
}

func (m *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	m.log.WithContext(ctx).Errorf(s, i)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.conf.LogLevel >= glogger.Error && (!errors.Is(err, glogger.ErrRecordNotFound) || !l.conf.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v] err:[%v] time:[%v ms] sql:[%v]", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v] err:[%v] time:[%v ms] rows:[%v] sql:[%v]", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.conf.SlowThreshold && l.conf.SlowThreshold != 0 && l.conf.LogLevel >= glogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.conf.SlowThreshold)
		if rows == -1 {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v]  slowLog:[%v] time:[%v ms] sql:[%v]", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v]  slowLog:[%v] time:[%v ms] rows:[%v] sql:[%v]", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.conf.LogLevel == glogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v] time:[%v ms] sql:[%v]", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			l.log.WithContext(ctx).Debugf("MysqlLogger_Trace FileWithLineNum:[%v] time:[%v ms] rows:[%v] sql:[%v]", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
