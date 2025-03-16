package tracing

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func validateConfig(cfg Config) error {
	if cfg.ExporterURL == "" {
		return ErrMissingExporterURL
	}

	if !cfg.TransportType.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidTransportType, cfg.TransportType)
	}

	return nil
}

// buildResource constructs a OpenTelemetry Resource object based on the provided Config
func buildResource(cfg Config) *resource.Resource {
	return resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.ServerName),
		semconv.ServiceVersion(cfg.Version),
		semconv.DeploymentEnvironment(cfg.Environment),
		semconv.TelemetrySDKName("opentelemetry"),
		semconv.TelemetrySDKLanguageGo,
		semconv.TelemetrySDKVersion(sdk.Version()),
	)
}
