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

// CompressionCodec denotes the compression codec for producing messages
type CompressionCodec sarama.CompressionCodec

const (
	// CompressionNone no compression
	CompressionNone = CompressionCodec(sarama.CompressionNone)

	// CompressionGZIP gzip compression
	CompressionGZIP = CompressionCodec(sarama.CompressionGZIP)

	// CompressionSnappy snappy compression
	CompressionSnappy = CompressionCodec(sarama.CompressionSnappy)

	// CompressionLZ4 lz4 compression
	CompressionLZ4 = CompressionCodec(sarama.CompressionLZ4)
)
