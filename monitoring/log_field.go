package monitoring

import (
	"encoding/json"

	"go.uber.org/zap"
)

type Field = zap.Field

func StringField(key string, value string) Field {
	return zap.String(key, value)
}

func IntField(key string, value int) Field {
	return zap.Int(key, value)
}

func JSONField(key string, value json.RawMessage) Field {
	if value == nil {
		return zap.Skip()
	}

	return zap.Reflect(key, value)
}
