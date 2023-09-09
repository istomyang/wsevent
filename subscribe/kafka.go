package subscribe

import (
	"context"
	"github.com/IBM/sarama"
	"sync/atomic"
	"wsevent/log"
)

type KafkaConfig struct {
	Hosts []string

	// One consumer should bind to one partition.
	PartitionID int32
	Topic       string
}

type kafkaSubscriber struct {
	ctx        context.Context
	cancel     context.CancelFunc
	config     KafkaConfig
	consumer   sarama.Consumer
	messages   chan []byte
	chanClosed atomic.Bool
}

func NewKafkaSubscriber(ctx context.Context, config KafkaConfig) Subscribe {
	ctx, cancel := context.WithCancel(ctx)
	return &kafkaSubscriber{
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
		messages: make(chan []byte),
	}
}

func (k *kafkaSubscriber) Get() (<-chan []byte, error) {
	consumer, err := k.consumer.ConsumePartition(k.config.Topic, k.config.PartitionID, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}
	go func() {
		for message := range consumer.Messages() {
			if k.chanClosed.Load() {
				break
			}
			log.Debug("kafkaSubscriber-get: %s", string(message.Value))
			k.messages <- message.Value
		}
	}()
	return k.messages, err
}

func (k *kafkaSubscriber) Run() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumer(k.config.Hosts, config)
	if err != nil {
		return err
	}
	k.consumer = consumer

	go func() {
		select {
		case <-k.ctx.Done():
			if err := k.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	log.Debug("kafkaSubscriber: run")
	return nil
}

func (k *kafkaSubscriber) Close() error {
	defer k.cancel()
	var err error
	err = k.consumer.Close()
	k.chanClosed.Store(true)
	close(k.messages)

	log.Debug("kafkaSubscriber: close")
	return err
}

var _ Subscribe = &kafkaSubscriber{}
