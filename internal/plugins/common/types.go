package common

import (
	"context"
	"io"
	"log"
	"log/slog"

	"github.com/hashicorp/go-hclog"
)

type PluginLogAdapter struct {
	underlying slog.Logger
	name       string
	leveler    slog.LevelVar
}

func NewPluginLogAdapter(parent slog.Logger, name string, level slog.Level) *PluginLogAdapter {
	a := PluginLogAdapter{
		underlying: parent,
		name:       name,
		leveler:    slog.LevelVar{},
	}
	a.leveler.Set(level)
	return &a
}

// Debug implements [hclog.Logger].
func (p *PluginLogAdapter) Debug(msg string, args ...interface{}) {
	p.underlying.Debug(msg, args...)
}

// Error implements [hclog.Logger].
func (p *PluginLogAdapter) Error(msg string, args ...interface{}) {
	p.underlying.Error(msg, args...)
}

func (p *PluginLogAdapter) GetSLevel() slog.Level {
	if p.underlying.Enabled(context.TODO(), slog.LevelDebug) {
		return slog.LevelDebug
	} else if p.underlying.Enabled(context.TODO(), slog.LevelInfo) {
		return slog.LevelInfo
	} else if p.underlying.Enabled(context.TODO(), slog.LevelWarn) {
		return slog.LevelWarn
	} else {
		return slog.LevelError
	}
}

// GetLevel implements [hclog.Logger].
func (p *PluginLogAdapter) GetLevel() hclog.Level {
	switch p.GetSLevel() {
	case slog.LevelDebug:
		return hclog.Debug
	case slog.LevelInfo:
		return hclog.Info
	case slog.LevelWarn:
		return hclog.Warn
	case slog.LevelError:
		return hclog.Error
	default:
		return hclog.Error
	}

}

// ImpliedArgs implements [hclog.Logger].
func (p *PluginLogAdapter) ImpliedArgs() []interface{} {
	panic("unimplemented")
}

// Info implements [hclog.Logger].
func (p *PluginLogAdapter) Info(msg string, args ...interface{}) {
	p.underlying.Info(msg, args...)
}

// IsDebug implements [hclog.Logger].
func (p *PluginLogAdapter) IsDebug() bool {
	return p.underlying.Handler().Enabled(context.Background(), slog.LevelDebug)

}

// IsError implements [hclog.Logger].
func (p *PluginLogAdapter) IsError() bool {
	return p.underlying.Handler().Enabled(context.Background(), slog.LevelError)

}

// IsInfo implements [hclog.Logger].
func (p *PluginLogAdapter) IsInfo() bool {
	return p.underlying.Handler().Enabled(context.Background(), slog.LevelInfo)
}

// IsTrace implements [hclog.Logger].
func (p *PluginLogAdapter) IsTrace() bool {
	return p.underlying.Handler().Enabled(context.Background(), slog.LevelDebug)
}

// IsWarn implements [hclog.Logger].
func (p *PluginLogAdapter) IsWarn() bool {
	return p.underlying.Handler().Enabled(context.Background(), slog.LevelWarn)
}

// Log implements [hclog.Logger].
func (p *PluginLogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Debug:
		p.underlying.Debug(msg, args...)
	case hclog.Error:
		p.underlying.Error(msg, args...)
	case hclog.Info:
		p.underlying.Info(msg, args...)
	case hclog.Warn:
		p.underlying.Warn(msg, args...)

	case hclog.Trace:
		return
	case hclog.Off:
		return
	case hclog.NoLevel:
		return
	default:
		slog.Error("unexpected hclog.Level", "level", level)
	}
}

// Name implements [hclog.Logger].
func (p *PluginLogAdapter) Name() string {
	return p.name
}

// Named implements [hclog.Logger].
func (p *PluginLogAdapter) Named(name string) hclog.Logger {
	return &PluginLogAdapter{
		underlying: p.underlying,
		name:       p.name + " " + name,
	}
}

// ResetNamed implements [hclog.Logger].
func (p *PluginLogAdapter) ResetNamed(name string) hclog.Logger {
	return &PluginLogAdapter{
		underlying: p.underlying,
		name:       name,
	}
}

// SetLevel implements [hclog.Logger].
func (p *PluginLogAdapter) SetLevel(level hclog.Level) {

	p.leveler.Set(slog.LevelInfo)

}

// StandardLogger implements [hclog.Logger].
func (p *PluginLogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	sLevel := p.GetSLevel()
	// opt := slog.HandlerOptions{
	// AddSource: false,
	// Level:     nil,
	// ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
	// panic("TODO")
	// },
	// }

	return slog.NewLogLogger(p.underlying.Handler(), sLevel)
}

// StandardWriter implements [hclog.Logger].
func (p *PluginLogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("unimplemented")
}

// Trace implements [hclog.Logger].
func (p *PluginLogAdapter) Trace(msg string, args ...interface{}) {
	p.underlying.Debug(msg, args...)
}

// Warn implements [hclog.Logger].
func (p *PluginLogAdapter) Warn(msg string, args ...interface{}) {
	p.underlying.Warn(msg, args...)
}

// With implements [hclog.Logger].
func (p *PluginLogAdapter) With(args ...interface{}) hclog.Logger {
	p2 := &PluginLogAdapter{*p.underlying.With(args...), p.name, slog.LevelVar{}}
	p2.leveler.Set(p.leveler.Level())
	return p2
}

var _ hclog.Logger = &PluginLogAdapter{}
