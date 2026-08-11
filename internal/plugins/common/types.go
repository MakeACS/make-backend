package common

import (
	"context"
	"io"
	"log"
	"log/slog"

	"github.com/hashicorp/go-hclog"
)

type UserDataProvider interface {
	FullNameForUser(userID int) (string, error)
	EmailForUser(userID int) (string, error)
}

type GRPCUserDataProvider struct {
	Impl UserDataProvider
}

type PluginLogAdapter struct {
	Underlying slog.Logger
	LoggerName string
}

// Debug implements [hclog.Logger].
func (p *PluginLogAdapter) Debug(msg string, args ...interface{}) {
	p.Underlying.Debug(msg, args...)
}

// Error implements [hclog.Logger].
func (p *PluginLogAdapter) Error(msg string, args ...interface{}) {
	p.Underlying.Error(msg, args...)
}

func (p *PluginLogAdapter) GetSLevel() slog.Level {
	if p.Underlying.Enabled(context.TODO(), slog.LevelDebug) {
		return slog.LevelDebug
	} else if p.Underlying.Enabled(context.TODO(), slog.LevelInfo) {
		return slog.LevelInfo
	} else if p.Underlying.Enabled(context.TODO(), slog.LevelWarn) {
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
	panic("unimplemented")
}

// IsDebug implements [hclog.Logger].
func (p *PluginLogAdapter) IsDebug() bool {
	panic("unimplemented")
}

// IsError implements [hclog.Logger].
func (p *PluginLogAdapter) IsError() bool {
	panic("unimplemented")
}

// IsInfo implements [hclog.Logger].
func (p *PluginLogAdapter) IsInfo() bool {
	panic("unimplemented")
}

// IsTrace implements [hclog.Logger].
func (p *PluginLogAdapter) IsTrace() bool {
	panic("unimplemented")
}

// IsWarn implements [hclog.Logger].
func (p *PluginLogAdapter) IsWarn() bool {
	panic("unimplemented")
}

// Log implements [hclog.Logger].
func (p *PluginLogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	panic("unimplemented")
}

// Name implements [hclog.Logger].
func (p *PluginLogAdapter) Name() string {
	return p.LoggerName
}

// Named implements [hclog.Logger].
func (p *PluginLogAdapter) Named(name string) hclog.Logger {
	return &PluginLogAdapter{
		Underlying: p.Underlying,
		LoggerName: name,
	}
}

// ResetNamed implements [hclog.Logger].
func (p *PluginLogAdapter) ResetNamed(name string) hclog.Logger {
	panic("unimplemented")
}

// SetLevel implements [hclog.Logger].
func (p *PluginLogAdapter) SetLevel(level hclog.Level) {
	// opt := slog.HandlerOptions{
	// AddSource: false,
	// Level:     nil,
	// ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
	// panic("TODO")
	// },
	// }
	panic("unimplemented")
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

	return slog.NewLogLogger(p.Underlying.Handler(), sLevel)
}

// StandardWriter implements [hclog.Logger].
func (p *PluginLogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("unimplemented")
}

// Trace implements [hclog.Logger].
func (p *PluginLogAdapter) Trace(msg string, args ...interface{}) {
	p.Underlying.Debug(msg, args...)
}

// Warn implements [hclog.Logger].
func (p *PluginLogAdapter) Warn(msg string, args ...interface{}) {
	p.Underlying.Warn(msg, args...)
}

// With implements [hclog.Logger].
func (p *PluginLogAdapter) With(args ...interface{}) hclog.Logger {
	return &PluginLogAdapter{*p.Underlying.With(args...), p.LoggerName}
}

var _ hclog.Logger = &PluginLogAdapter{}
