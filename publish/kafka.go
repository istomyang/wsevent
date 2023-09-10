package publish

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/istomyang/wsevent/log"
)

type kafkaPublisher struct {
	ctx      context.Context
	cancel   context.CancelFunc
	config   KafkaConfig
	producer sarama.AsyncProducer
}

type KafkaConfig struct {
	Hosts []string
	Topic string
}

func NewKafkaPublisher(ctx context.Context, config KafkaConfig) Publish {
	ctx, cancel := context.WithCancel(ctx)
	return &kafkaPublisher{
		ctx:    ctx,
		cancel: cancel,
		config: config,
	}
}

func (k *kafkaPublisher) Send(data []byte) error {
	var message = &sarama.ProducerMessage{
		Topic: k.config.Topic,
		Value: sarama.ByteEncoder(data),
	}
lo:
	for {
		select {
		case k.producer.Input() <- message:
			log.Debug("kafkaPublisher-send: %v", string(data))
			break lo
		case err := <-k.producer.Errors():
			log.Debug("kafkaPublisher-send: error %s", err.Error())
			return err
		}
	}
	return nil
}

func (k *kafkaPublisher) Run() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(k.config.Hosts, config)
	if err != nil {
		return err
	}
	k.producer = producer

	go func() {
		select {
		case <-k.ctx.Done():
			log.Debug("kafkaPublisher: closed by context.Done")
			if err := k.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	log.Debug("kafkaPublisher: run")

	return nil
}

func (k *kafkaPublisher) Close() error {
	var err error
	defer k.cancel()
	err = k.producer.Close()

	log.Debug("kafkaPublisher: close")
	return err
}

var _ Publish = &kafkaPublisher{}
