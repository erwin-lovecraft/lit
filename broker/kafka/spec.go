package kafka

import (
	"github.com/IBM/sarama"
)

// AckMode denotes the ack mode for producing messages
type AckMode sarama.RequiredAcks

const (
	// AckModeNone don't wait for response
	AckModeNone = AckMode(sarama.NoResponse)

	// AckModeLocal wait for local commit only
	AckModeLocal = AckMode(sarama.WaitForLocal)

	// AckModeInSync wait for commit to all in-sync replicas
	AckModeInSync = AckMode(sarama.WaitForAll)
)
