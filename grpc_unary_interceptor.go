package lit

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/viebiz/lit/monitoring"
	"github.com/viebiz/lit/monitoring/instrumentgrpc"
)

func unaryServerInterceptor(rootCtx context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rs any, err error) {
		// Start tracing for incoming unary call request
		ctx, reqMeta, endInstrumentation := instrumentgrpc.StartUnaryIncomingCall(ctx, monitoring.FromContext(rootCtx), info.FullMethod, req)
		defer func() {
			if p := recover(); p != nil {
				rcvErr, ok := p.(error)
				if !ok {
					rcvErr = fmt.Errorf("%v", p)
				}

				monitoring.FromContext(ctx).Errorf(rcvErr, "Caught a panic: %s", debug.Stack())
				endInstrumentation(rcvErr)

				err = ErrDefaultInternal
			}
		}()

		// Log incoming grpc call
		logIncomingGRPCCall(ctx, reqMeta)

		rs, err = handler(ctx, req)

		endInstrumentation(err)

		// Log response body
		monitoring.FromContext(ctx).
			WithTag("grpc.response_body", string(parseProtoMessage(rs))).
			Infof("Wrote gRPC response")

		return rs, err
	}
}

func parseProtoMessage(resp any) []byte {
	msg, ok := resp.(proto.Message)
	if !ok {
		return nil // Ignore req body if it not proto message
	}

	b, err := protojson.Marshal(msg)
	if err != nil {
		return nil // Ignore if it is invalid proto message
	}

	return b
}

func logIncomingGRPCCall(ctx context.Context, reqMeta instrumentgrpc.RequestMetadata) {
	logTags := map[string]string{
		"grpc.service_method": reqMeta.ServiceMethod,
	}

	// Also skip logging if the request body is empty, by default protojson marshaler will return "{}"
	if len(reqMeta.BodyToLog) > 2 {
		logTags["grpc.request_body"] = string(reqMeta.BodyToLog)
	}

	monitoring.FromContext(ctx).
		With(logTags).
		Infof("grpc.unary_incoming_call")
}
