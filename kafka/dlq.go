package kafka

import (
	"context"
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	// DLQHeaderOriginalTopic contains the source record topic.
	DLQHeaderOriginalTopic = "kafka-dlq-original-topic"
	// DLQHeaderOriginalPartition contains the source partition in base 10.
	DLQHeaderOriginalPartition = "kafka-dlq-original-partition"
	// DLQHeaderOriginalOffset contains the source offset in base 10.
	DLQHeaderOriginalOffset = "kafka-dlq-original-offset"
	// DLQHeaderFailureKind is either "decode" or "handler" and never contains
	// an error string.
	DLQHeaderFailureKind = "kafka-dlq-failure-kind"
)

type recordRejection struct {
	record *kgo.Record
	kind   string
	cause  error
}

func (consumer *Consumer[T]) deliverDLQ(ctx context.Context, rejections []recordRejection) (bool, error) {
	if len(rejections) == 0 {
		return true, nil
	}
	records := make([]*kgo.Record, len(rejections))
	for index, rejection := range rejections {
		records[index] = dlqRecord(consumer.config.DLQ.Topic, rejection)
	}
	sources := make([]*kgo.Record, len(rejections))
	for index, rejection := range rejections {
		sources[index] = rejection.record
	}
	err := doWithRetry(ctx, effectiveRetry(consumer.config.Retry, consumer.config.DLQ.Retry), effectiveObserver(consumer.config.Observer), AttemptDLQ, sources, func(produceCtx context.Context) error {
		return deliveryError(consumer.client.ProduceSync(produceCtx, records...))
	})
	if err != nil && consumer.config.DLQ.OnFailure == DLQFailureStop {
		return false, err
	}
	return err == nil, nil
}

func dlqRecord(topic string, rejection recordRejection) *kgo.Record {
	source := rejection.record
	headers := make([]kgo.RecordHeader, 0, len(source.Headers)+4)
	for _, header := range source.Headers {
		if !isDLQHeader(header.Key) {
			headers = append(headers, header)
		}
	}
	headers = append(headers,
		kgo.RecordHeader{Key: DLQHeaderOriginalTopic, Value: []byte(source.Topic)},
		kgo.RecordHeader{Key: DLQHeaderOriginalPartition, Value: []byte(strconv.FormatInt(int64(source.Partition), 10))},
		kgo.RecordHeader{Key: DLQHeaderOriginalOffset, Value: []byte(strconv.FormatInt(source.Offset, 10))},
		kgo.RecordHeader{Key: DLQHeaderFailureKind, Value: []byte(rejection.kind)},
	)
	return &kgo.Record{
		Topic:     topic,
		Key:       source.Key,
		Value:     source.Value,
		Headers:   headers,
		Timestamp: source.Timestamp,
	}
}

func isDLQHeader(key string) bool {
	switch key {
	case DLQHeaderOriginalTopic, DLQHeaderOriginalPartition, DLQHeaderOriginalOffset, DLQHeaderFailureKind:
		return true
	default:
		return false
	}
}
