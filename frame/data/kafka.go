package data

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/log"
)

var kafkaCtx context.Context

func InitKafka(ctx context.Context) {
	kafkaCtx = ctx
}

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
	config.Consumer.Offsets.CommitInterval = time.Second

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

// NewProducer 创建 kafka 生产者 异步
func NewAsyncProducer() (sarama.AsyncProducer, error) {
	brokers := viper.GetStringSlice("data.kafka.brokers")
	if len(brokers) == 0 {
		return nil, ErrKafkaNoBrokers
	}

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Compression = sarama.CompressionLZ4
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	return sarama.NewAsyncProducer(brokers, config)
}

var aSyncProducer sarama.AsyncProducer
var aonce sync.Once

func MustGetAsyncProducer() sarama.AsyncProducer {
	if aSyncProducer != nil {
		return aSyncProducer
	}

	aonce.Do(func() {
		producer, err := NewAsyncProducer()
		if err != nil {
			panic(err)
		}
		aSyncProducer = producer

		go func() {
			for {
				select {
				case <-producer.Successes():
					log.Default().Sugar().Infof("producer success")
				case <-kafkaCtx.Done():
					break
				case err := <-producer.Errors():
					log.Default().Sugar().Errorf("producer error: %s ", err.Error())
				}
			}
		}()
	})

	if aSyncProducer == nil {
		panic(errors.New("global aSyncProducer not inited."))
	}

	return aSyncProducer
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

// KafkaConsumeV1 带处理停止信号的 kafka 消费
func KafkaConsumeV1(signals chan os.Signal, topics []string, groupID string, errCallback func(err error),
	callback func(msgBytes []byte), finishCallback func()) {
	kafkaConsumer, err := NewConsumerWithNewestOffset(topics, groupID)
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
		case s := <-signals:
			signals <- s
			fmt.Println("stop starting....", len(signals))
			finishCallback()
			return
		}
	}
}

// KafkaConsumeV2 带处理停止信号的 kafka 消费
func KafkaConsumeV2(signals chan os.Signal, topics []string, groupID string, errCallback func(err error),
	callback func(msgBytes []byte), finishCallback func()) {
	kafkaConsumer, err := NewConsumer(topics, groupID)
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
		case s := <-signals:
			signals <- s
			fmt.Println("stop starting....", len(signals))
			finishCallback()
			return
		}
	}
}

// KafkaConsumeV3 带处理context停止信号的 kafka 消费
func KafkaConsumeV3(ctx context.Context, topics []string, groupID string, errCallback func(err error),
	callback func(msgBytes []byte), finishCallback func()) {
	kafkaConsumer, err := NewConsumer(topics, groupID)
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
		case <-ctx.Done():
			fmt.Println("stop starting....", ctx.Err())
			finishCallback()
			return
		}
	}
}
