package pnpzapsentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
)

var sentryLevelValues = map[sentry.Level]int{
	sentry.LevelDebug:   1,
	sentry.LevelInfo:    2,
	sentry.LevelWarning: 3,
	sentry.LevelError:   4,
	sentry.LevelFatal:   5,
}

type ZapCore struct {
	underlying       zapcore.Core
	client           *sentry.Client
	hub              *sentry.Hub
	err              error
	flushTimeout     time.Duration
	reportLevelValue int
}

func NewZapCore(underlying zapcore.Core, client *sentry.Client, flushTimeout time.Duration, reportLevelValue int) ZapCore {
	return ZapCore{
		underlying:       underlying,
		client:           client,
		flushTimeout:     flushTimeout,
		reportLevelValue: reportLevelValue,
	}
}

func (z ZapCore) Enabled(level zapcore.Level) bool {
	return z.underlying.Enabled(level)
}

func (z ZapCore) With(fields []zapcore.Field) zapcore.Core {
	newFields := make([]zapcore.Field, len(fields))
	copy(newFields, fields)

	result := ZapCore{
		underlying:       z.underlying.With(newFields),
		client:           z.client,
		hub:              z.hub,
		err:              z.err,
		flushTimeout:     z.flushTimeout,
		reportLevelValue: z.reportLevelValue,
	}

	for _, field := range fields {
		switch field.Key {
		case "error":
			if err, ok := field.Interface.(error); ok {
				result.err = err
			}
		case "context":
			if ctx, ok := field.Interface.(context.Context); ok {
				result.hub = sentry.GetHubFromContext(ctx)
			}
		}

	}

	return result
}

func (z ZapCore) Check(entry zapcore.Entry, entry2 *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	result := z.underlying.Check(entry, entry2)
	result.AddCore(entry, z)

	return result
}

func (z ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	level := map[zapcore.Level]sentry.Level{
		zapcore.DebugLevel:  sentry.LevelDebug,
		zapcore.InfoLevel:   sentry.LevelInfo,
		zapcore.WarnLevel:   sentry.LevelWarning,
		zapcore.ErrorLevel:  sentry.LevelError,
		zapcore.DPanicLevel: sentry.LevelFatal,
		zapcore.PanicLevel:  sentry.LevelFatal,
		zapcore.FatalLevel:  sentry.LevelFatal,
	}[entry.Level]
	levelValue := sentryLevelValues[level]
	//pc := make([]uintptr, 50)
	//callers := runtime.Callers(0, pc)
	//frames := runtime.CallersFrames(pc[:callers])
	trace := sentry.NewStacktrace()

	for i, frame := range trace.Frames {
		// if fram abspath and linno matches then we skip all that left
		if frame.AbsPath == entry.Caller.File && frame.Lineno == entry.Caller.Line {
			trace.Frames = trace.Frames[:i+1]
			break
		}
	}

	if levelValue >= z.reportLevelValue {
		go func() {
			extra := make(map[string]interface{})
			for _, field := range fields {
				extra[field.Key] = field.Interface
			}
			hub := z.hub
			if hub == nil {
				hub = sentry.NewHub(z.client, sentry.NewScope())
			}
			//frames := runtime.CallersFrames([]uintptr{entry.Caller.PC})

			event := &sentry.Event{
				Level:     level,
				Message:   entry.Message,
				Extra:     extra,
				Logger:    entry.LoggerName,
				Timestamp: entry.Time,
			}
			if z.err != nil {
				event.Exception = []sentry.Exception{
					{
						Value:      z.err.Error(),
						Stacktrace: trace,
					},
				}
			} else {
				event.Exception = []sentry.Exception{
					{
						Value:      entry.Message,
						Stacktrace: trace,
					},
				}
			}

			hub.CaptureEvent(event)
		}()
	}

	return nil
}

func (z ZapCore) Sync() error {
	return z.underlying.Sync()
}
