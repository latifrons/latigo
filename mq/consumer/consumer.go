package consumer

import (
	"context"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/logger"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"time"
)

type ConsumerOption func(*ReliableRabbitConsumer)

type ReliableRabbitConsumerArgs struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp091.Table
}

type ReliableRabbitConsumer struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string

	HandleFunc func(ctx context.Context, msg amqp091.Delivery) interface{}
	initFunc   func(channel *amqp091.Channel) error
	cleanFunc  func(channel *amqp091.Channel) error
	logger     logger.Logger

	dailer        *amqpextra.Dialer
	consumer      *consumer.Consumer
	consumerArgs  ReliableRabbitConsumerArgs
	prefetchCount int
	global        bool
}

func NewReliableRabbitConsumer(url string, exchageName string, queueName string, routingKey string, handleFunc func(ctx context.Context, msg amqp091.Delivery) interface{}, opts ...ConsumerOption) *ReliableRabbitConsumer {
	c := &ReliableRabbitConsumer{
		URL:           url,
		ExchangeName:  exchageName,
		QueueName:     queueName,
		RoutingKey:    routingKey,
		HandleFunc:    handleFunc,
		prefetchCount: 1,
		global:        false,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithInitFunc(f func(*amqp091.Channel) error) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.initFunc = f
	}
}

func WithCleanFunc(f func(*amqp091.Channel) error) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.cleanFunc = f
	}
}

func WithLogger(l logger.Logger) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.logger = l
	}
}

func WithConsumerArgs(consumerArgs ReliableRabbitConsumerArgs) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.consumerArgs = consumerArgs
	}
}

func WithQos(prefetchCount int, global bool) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.prefetchCount = prefetchCount
		c.global = global
	}
}

func (c *ReliableRabbitConsumer) Start() (err error) {
	c.dailer, err = amqpextra.NewDialer(amqpextra.WithURL(c.URL))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	connection, err := c.dailer.Connection(ctx)
	if err != nil {
		return
	}

	channel, err := connection.Channel()
	if err != nil {
		return
	}
	defer func(channel *amqp091.Channel) {
		err := channel.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close channel")
		}
	}(channel)

	if c.initFunc != nil {
		err = c.initFunc(channel)
		if err != nil {
			return
		}
	}

	h := consumer.HandlerFunc(c.HandleFunc)

	c.consumer, err = c.dailer.Consumer(
		consumer.WithExchange(c.ExchangeName, c.RoutingKey),
		consumer.WithQueue(c.QueueName),
		consumer.WithLogger(c.logger),
		consumer.WithHandler(h),
		consumer.WithQos(c.prefetchCount, c.global),
		consumer.WithConsumeArgs(c.consumerArgs.Consumer,
			c.consumerArgs.AutoAck,
			c.consumerArgs.Exclusive,
			c.consumerArgs.NoLocal,
			c.consumerArgs.NoWait,
			c.consumerArgs.Args))

	return
}

func (c *ReliableRabbitConsumer) Reset() {
	// must delete all
	log.Info().Msg("reliable rabbit consumer reset")
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()

	conn, err := c.dailer.Connection(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to reset")
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("failed to reset")
	}
	defer func(channel *amqp091.Channel) {
		err := channel.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close channel")
		}
	}(channel)

	if c.cleanFunc != nil {
		err = c.cleanFunc(channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to clean")
		}
	}
	if c.initFunc != nil {
		err := c.initFunc(channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to init")
		}
	}
}

func (c *ReliableRabbitConsumer) Stop() {
	c.consumer.Close()
	c.dailer.Close()
}
