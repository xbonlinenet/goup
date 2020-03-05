package data

import (
	"errors"
	"os"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/viper"
)

// ErrKafkaNoBrokers brokers 未配置错误
var ErrKafkaNoBrokers = errors.New("Brokers not configed")

// NewConsumer 创建 kafka 消费者
func NewConsumer(topics []string, groupID string) (*cluster.Consumer, error) {
	config := cluster.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.ChannelBufferSize = 1024
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := viper.GetStringSlice("data.kafka.brokers")
	if len(brokers) == 0 {
		return nil, ErrKafkaNoBrokers
	}
	// init consumer
	return cluster.NewConsumer(brokers, groupID, topics, config)
}

// NewConsumerWithNewestOffset 创建 kafka 消费者
func NewConsumerWithNewestOffset(topics []string, groupID string) (*cluster.Consumer, error) {
	config := cluster.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.ChannelBufferSize = 1024
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	brokers := viper.GetStringSlice("data.kafka.brokers")
	if len(brokers) == 0 {
		return nil, ErrKafkaNoBrokers
	}
	// init consumer
	return cluster.NewConsumer(brokers, groupID, topics, config)
}

// NewProducer 创建 kafka 生产者
func NewProducer() (sarama.SyncProducer, error) {
	brokers := viper.GetStringSlice("data.kafka.brokers")
	if len(brokers) == 0 {
		return nil, ErrKafkaNoBrokers
	}

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Compression = sarama.CompressionLZ4
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	return sarama.NewSyncProducer(brokers, config)
}

func KafkaConsume(signals chan os.Signal, topic, groupID string, errCallback func(err error), callback func(msgBytes []byte)) {
	kafkaConsumer, err := NewConsumerWithNewestOffset([]string{topic}, groupID)
	if err != nil {
		panic(err)
	}

	defer kafkaConsumer.Close()

	// consume errors
	go func() {
		for err := range kafkaConsumer.Errors() {
			errCallback(err)
		}
	}()

	// consume notifications
	go func() {
		for range kafkaConsumer.Notifications() {
		}
	}()

	// consume messages, watch signals
	for {
		select {
		case msg, ok := <-kafkaConsumer.Messages():
			if !ok {
				break
			}

			kafkaConsumer.MarkOffset(msg, "") // mark message as processed
			callback(msg.Value)
		case <-signals:
			return
		}
	}
}
