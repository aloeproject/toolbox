package crontab

import (
	"context"
	"github.com/aloeproject/toolbox/logger"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var _ cron.Logger = (*CronLog)(nil)

type CronLog struct {
	log logger.ILogger
}

func NewCronLog(log logger.ILogger) *CronLog {
	return &CronLog{log: log}
}

func (c *CronLog) Info(msg string, keysAndValues ...interface{}) {
	c.log.Info(msg)
}

func (c *CronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	c.log.Errorf("error:[%v] msg:[%s]", err, msg)
}

type WorkCrontab interface {
	cron.Job
	WithContext(ctx context.Context)
}

func NewCrontab(job WorkCrontab) *Crontab {
	c := &Crontab{
		job: job,
	}
	return c
}

var _ cron.Job = (*Crontab)(nil)

type Crontab struct {
	job WorkCrontab
}

func (c *Crontab) Run() {
	tracer := otel.Tracer("crontab")
	ctx, span := tracer.Start(context.Background(), "crontab",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	c.job.WithContext(ctx)
	c.job.Run()
}
