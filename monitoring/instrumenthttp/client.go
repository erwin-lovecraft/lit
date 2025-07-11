package instrumenthttp

import (
	"context"
	"net/http"

	"github.com/viebiz/lit/monitoring"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// StartOutgoingGroupSegment starts an outgoing HTTP client unary request group segment to group all client calls
// including retries. The return is the `End` func to close the segment
func StartOutgoingGroupSegment(
	ctx context.Context,
	extSvcInfo monitoring.ExternalServiceInfo,
	serviceName,
	reqMethod,
	reqURL string,
) (context.Context, func(error)) {
	m := monitoring.FromContext(ctx)

	ctx, span := tracer.Start(ctx, httpOutgoingSpanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.HTTPRequestMethodKey.String(reqMethod),
			semconv.URLFull(reqURL),
			semconv.ServerAddress(extSvcInfo.Hostname+":"+extSvcInfo.Port),
		),
	)
	ctx = monitoring.SetInContext(ctx, monitoring.InjectOutgoingTracingInfo(m, span.SpanContext()))

	return ctx, func(err error) {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.RecordError(err)
		span.End()
	}
}

// StartOutgoingSegment starts a outgoing HTTP client segment. The return is the `End` func to close the segment
func StartOutgoingSegment(ctx context.Context, r *http.Request) (context.Context, func(int, error)) {
	// Start a child span as segment
	ctx, span := tracer.Start(ctx, httpRequestSpanName, trace.WithSpanKind(trace.SpanKindClient))

	// Inject Span Context to request header before send
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	return ctx, func(statusCode int, err error) {
		if statusCode != 0 {
			span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
		}
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
		span.End()
	}
}
