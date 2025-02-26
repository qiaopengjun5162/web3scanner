// Package utils provides utility functions for the web3scanner.
package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/log"
)

var (
	_ logger.Interface = Logger{}

	SlowThresholdMilliseconds = 200
)

type Logger struct {
	log log.Logger
}

// NewLogger creates a new Logger instance with a specific module name.
//
// It takes an Ethereum log.Logger as an argument and returns a Logger
// that implements gorm's logger.Interface. The new Logger is initialized
// with the module name set to "db", which can be used to filter or
// categorize log messages related to database operations.
//
// Parameters:
//
//	log - the Ethereum logger to be wrapped
//
// Returns:
//
//	A Logger instance implementing the gorm logger.Interface
func NewLogger(log log.Logger) Logger {
	return Logger{log.New("module", "db")}
}

func (l Logger) LogMode(lvl logger.LogLevel) logger.Interface {
	return l
}

// Info logs a message at the info level.
//
// Parameters:
//
//	ctx - the context for the log operation
//	msg - the format string for the log message
//	data - additional arguments to be formatted into the log message
//
// Returns:
//
//	none
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Info(fmt.Sprintf(msg, data...))
}

func (l Logger) Warn(_ context.Context, msg string, data ...interface{}) {
	l.log.Warn(fmt.Sprintf(msg, data...))
}

// Error logs a message at the error level.
//
// Parameters:
//
//	_ - not used
//	msg - the format string for the log message
//	data - arguments to be formatted into the log message
//
// Returns:
//
//	none
func (l Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if ctx == nil {
		// 如果需要处理 nil context，可以根据实际情况添加逻辑
		return
	}
	// 确保 msg 和 data 的长度匹配
	if len(data) > 0 && strings.Contains(msg, "%") {
		l.log.Error(fmt.Sprintf(msg, data...))
	} else {
		l.log.Error(msg)
	}
}

func (l Logger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsedMs := time.Since(begin).Milliseconds()

	// omit any values for batch inserts as they can be very long
	sql, rows := fc()
	if i := strings.Index(strings.ToLower(sql), "values"); i > 0 {
		sql = fmt.Sprintf("%sVALUES (...)", sql[:i])
	}

	if elapsedMs < 200 {
		l.log.Debug("database operation", "duration_ms", elapsedMs, "rows_affected", rows, "sql", sql)
	} else {
		l.log.Warn("database operation", "duration_ms", elapsedMs, "rows_affected", rows, "sql", sql)
	}
}
