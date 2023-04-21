package log

import (
	"context"
	"errors"
	"github.com/qiafan666/quickweb/commons"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	LogLevel                  int
	IgnoreRecordNotFoundError bool
	SlowSqlTime               time.Duration
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = int(level)
	return l
}
func (l *GormLogger) Info(ctx context.Context, template string, args ...interface{}) {
	if l.LogLevel <= commons.Debug {
		GormLog.Infof(getTraceId(ctx)+template, args...)
	}
}
func (l *GormLogger) Warn(ctx context.Context, template string, args ...interface{}) {
	if l.LogLevel <= commons.Warn {
		GormLog.Warnf(getTraceId(ctx)+template, args...)
	}
}
func (l *GormLogger) Error(ctx context.Context, template string, args ...interface{}) {
	if l.LogLevel <= commons.Error {
		GormLog.Errorf(getTraceId(ctx)+template, args...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= commons.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel <= commons.Error && (errors.Is(err, gorm.ErrRecordNotFound)) || !l.IgnoreRecordNotFoundError:
		sql, _ := fc()
		l.Error(ctx, "sql: %s time: %.3f error : %s", sql, float64(elapsed.Nanoseconds())/1e6, err.Error())
	case l.LogLevel <= commons.Info:
		sql, rows := fc()
		l.Info(ctx, "sql : %s , affected : %d , time: %.3f ms", sql, rows, float64(elapsed.Nanoseconds())/1e6)
	case elapsed > l.SlowSqlTime && l.SlowSqlTime != 0 && l.LogLevel <= commons.Warn:
		sql, rows := fc()
		l.Warn(ctx, "slow sql : %s , affected : %d, time: %.3f ms", sql, rows, float64(elapsed.Nanoseconds())/1e6)
	}
}
