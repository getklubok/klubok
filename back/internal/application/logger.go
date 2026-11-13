package application

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"klubok/internal/domain"
)

const (
	samplingDenominator = 10
	samplingFirst       = 1
)

// samplingCore applies zap sampling only to DEBUG and INFO levels,
// passing WARN and above through unchanged.
type samplingCore struct {
	core    zapcore.Core
	sampler zapcore.Core
}

func (c *samplingCore) Enabled(lvl zapcore.Level) bool {
	return c.core.Enabled(lvl)
}

func (c *samplingCore) With(fields []zapcore.Field) zapcore.Core {
	return &samplingCore{
		core:    c.core.With(fields),
		sampler: c.sampler.With(fields),
	}
}

func (c *samplingCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.core.Enabled(entry.Level) {
		return ce
	}
	if entry.Level >= zapcore.WarnLevel {
		return c.core.Check(entry, ce)
	}
	return c.sampler.Check(entry, ce)
}

func (c *samplingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// WARN+ пишем напрямую, INFO- пишем через семплер
	if entry.Level >= zapcore.WarnLevel {
		//nolint:wrapcheck // passthrough to inner core
		return c.core.Write(entry, fields)
	}
	//nolint:wrapcheck // passthrough to inner core
	return c.sampler.Write(entry, fields)
}

func (c *samplingCore) Sync() error {
	//nolint:wrapcheck // passthrough to inner core
	return c.core.Sync()
}

// NewLogger creates a new zap logger.
// In production, samples ~10% of INFO-level logs while keeping 100% of WARN+.
func NewLogger(logLevel zapcore.Level, env domain.Environment) (*zap.Logger, error) {
	if logLevel == zapcore.DebugLevel {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("zap.NewDevelopment: %w", err)
		}

		logger.Debug("Debug mode running")

		return logger, nil
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("zap.NewProduction: %w", err)
	}

	if env == domain.EnvironmentProd {
		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return &samplingCore{
				core:    core,
				sampler: zapcore.NewSamplerWithOptions(core, time.Second, samplingFirst, samplingDenominator),
			}
		}))
	}

	return logger, nil
}
