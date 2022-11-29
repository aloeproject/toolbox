package logger

import "gopkg.in/natefinch/lumberjack.v2"

type lumberjeckOption struct {
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	filename string

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	maxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	maxAge int

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	maxBackups int

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	localTime bool

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	compress bool
	// contains filtered or unexported fields
}

type LumberjeckOption func(*lumberjeckOption)

func WithMaxSize(size int) LumberjeckOption {
	return func(option *lumberjeckOption) {
		option.maxSize = size
	}
}

func WithMaxAge(age int) LumberjeckOption {
	return func(option *lumberjeckOption) {
		option.maxAge = age
	}
}

func WithMaxBackups(back int) LumberjeckOption {
	return func(option *lumberjeckOption) {
		option.maxBackups = back
	}
}

func WithLocalTime(back int) LumberjeckOption {
	return func(option *lumberjeckOption) {
		option.maxBackups = back
	}
}

func WithCompress(compress bool) LumberjeckOption {
	return func(option *lumberjeckOption) {
		option.compress = compress
	}
}

func NewLumberjackLogger(filename string, opt ...LumberjeckOption) *lumberjack.Logger {
	defaultOpt := lumberjeckOption{
		filename: filename,
		maxSize:  100,
		maxAge:   7,
		compress: false,
	}

	for _, o := range opt {
		o(&defaultOpt)
	}

	return &lumberjack.Logger{
		Filename:   defaultOpt.filename,   //// 日志文件路径，默认 os.TempDir()
		MaxSize:    defaultOpt.maxSize,    // megabytes  每个日志文件保存10M，默认 100M
		MaxBackups: defaultOpt.maxBackups, //保留30个备份，默认不限
		MaxAge:     defaultOpt.maxAge,     //days 保留7天，默认不限
		Compress:   defaultOpt.compress,   // disabled by default 是否压缩，默认不压缩
	}
}
