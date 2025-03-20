package kafka

import (
	"errors"
)

var (
	ErrEmptyTopic = errors.New("topic is empty")

	ErrNoPayloadProvided = errors.New("no payload provided")
)
