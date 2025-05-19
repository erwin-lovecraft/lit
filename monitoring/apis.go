package monitoring

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/viebiz/lit/monitoring/tracing"
)

// Config holds Monitor configuration
type Config struct {
	ServerName      string
	Environment     string
	Version         string
	Writer          io.Writer // Support write log to buffer for testing
	SentryDSN       string    // To capture error, skip init Sentry if it's not provided
	OtelExporterURL string    // To support OpenTelemetry
	Tags            map[string]string
}

// New creates a new Monitor instance
func New(cfg Config) (*Monitor, error) {
	// Setup logger
	var w io.Writer = os.Stdout
	if cfg.Writer != nil {
		w = cfg.Writer
	}

	if cfg.Tags == nil {
		cfg.Tags = make(map[string]string)
	}
	cfg.Tags["server.name"] = cfg.ServerName
	cfg.Tags["environment"] = cfg.Environment
	cfg.Tags["version"] = cfg.Version
	m := &Monitor{
		logger:  zap.New(newZapCore(w)),
		logTags: cfg.Tags,
	}

	// Setup sentry
	sentryClient, err := initSentry(sentryConfig{
		DSN:         cfg.SentryDSN,
		ServerName:  cfg.ServerName,
		Environment: cfg.Environment,
		Version:     cfg.Version,
	}, m)
	if err != nil {
		return nil, err
	}
	m.sentryClient = sentryClient

	// Setup tracing service
	if err := tracing.Init(context.Background(), tracing.Config{
		ServerName:  cfg.ServerName,
		Environment: cfg.Environment,
		Version:     cfg.Version,
		ExporterURL: cfg.OtelExporterURL,
	}); err != nil {
		// Can skip if Exporter URL not provided
		if !errors.Is(err, tracing.ErrMissingExporterURL) {
			return nil, err
		}

		m.Infof("OTelExporter URL not provided. Not using Distributed Tracing")
	}

	return m, nil
}
