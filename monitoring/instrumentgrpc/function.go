package instrumentgrpc

import (
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// extractFullMethod extracts full method /weather.WeatherService/GetWeatherInfo
func extractFullMethod(fullMethod string) (string, string) {
	fullMethod = strings.Trim(fullMethod, "/")
	parts := strings.Split(fullMethod, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], ""
}

// serializeProtoMessage converts protobuf request to JSON bytes
// output may be unstable due to known issues: https://github.com/golang/protobuf/issues/1121
func serializeProtoMessage(req any) []byte {
	msg, ok := req.(proto.Message)
	if !ok {
		return nil // Ignore req body if it not proto message
	}

	b, err := protojson.Marshal(msg)
	if err != nil {
		return nil // Ignore if it is invalid proto message
	}

	return b
}
