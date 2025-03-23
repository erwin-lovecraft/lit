package instrumentgrpc

import (
	"context"

	"github.com/viebiz/lit/monitoring"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func StartUnaryIncomingCall(ctx context.Context, m *monitoring.Monitor, fullMethod string, req any) (context.Context, RequestMetadata, func(error)) {
	// Init log fields
	logTags := map[string]string{
		rpcSystemKey: "grpc",
	}

	attrs := []attribute.KeyValue{
		semconv.RPCSystemGRPC,
	}

	if pr, ok := peer.FromContext(ctx); ok {
		logTags[networkPeerAddressKey] = pr.Addr.String()
		logTags[networkTransportKey] = pr.Addr.Network()

		attrs = append(attrs,
			semconv.NetworkPeerAddress(pr.Addr.String()),
			semconv.NetworkTransportKey.String(pr.Addr.Network()),
		)
	}

	if svc, method := extractFullMethod(fullMethod); method != "" {
		logTags[rpcServiceKey] = svc
		logTags[rpcMethodKey] = method

		attrs = append(attrs,
			semconv.RPCService(svc),
			semconv.RPCMethod(method),
		)
	}

	reqMeta := RequestMetadata{
		ServiceMethod: fullMethod,
	}

	// Log request body
	if shouldLogUnaryRequestBody {
		reqMeta.BodyToLog = serializeProtoMessage(req)
	}

	// Extract metadata from incoming context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy() // because it's not safe to modify
	}

	// Extract span context from metadata
	curSpanCtx := otel.GetTextMapPropagator().Extract(ctx, mdCarrier(md))
	spanCtx := trace.SpanContextFromContext(curSpanCtx)

	ctx, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), unaryIncomingSpanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
	)
	m = monitoring.InjectTracingInfo(m, span.SpanContext())

	m = m.With(logTags)
	ctx = monitoring.SetInContext(ctx, m)

	return ctx,
		reqMeta,
		func(err error) {
			if err == nil {
				span.SetStatus(codes.Ok, "")
				span.SetAttributes(semconv.RPCGRPCStatusCodeOk)
			} else {
				span.SetStatus(codes.Error, err.Error())
				errStatus := status.Convert(err)
				span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(errStatus.Code())))
			}

			span.End()
		}
}

type RequestMetadata struct {
	ServiceMethod string
	BodyToLog     []byte
}
