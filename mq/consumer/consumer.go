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

type ReliableRabbitConsumer struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string

	HandleFunc func(ctx context.Context, msg amqp091.Delivery) interface{}
	initFunc   func(channel *amqp091.Channel) error
	cleanFunc  func(channel *amqp091.Channel) error
	logger     logger.Logger

	dailer     *amqpextra.Dialer
	connection *amqp091.Connection
	channel    *amqp091.Channel
	consumer   *consumer.Consumer
}

func NewReliableRabbitConsumer(url string, exchageName string, queueName string, routingKey string, handleFunc func(ctx context.Context, msg amqp091.Delivery) interface{}, opts ...ConsumerOption) *ReliableRabbitConsumer {
	c := &ReliableRabbitConsumer{
		URL:          url,
		ExchangeName: exchageName,
		QueueName:    queueName,
		RoutingKey:   routingKey,
		HandleFunc:   handleFunc,
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

func (c *ReliableRabbitConsumer) Start() (err error) {
	c.dailer, err = amqpextra.NewDialer(amqpextra.WithURL(c.URL))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c.connection, err = c.dailer.Connection(ctx)
	if err != nil {
		return
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		return
	}

	if c.initFunc != nil {
		err = c.initFunc(c.channel)
		if err != nil {
			return
		}
	}

	h := consumer.HandlerFunc(c.HandleFunc)

	c.consumer, err = c.dailer.Consumer(
		consumer.WithExchange(c.ExchangeName, c.RoutingKey),
		consumer.WithQueue(c.QueueName),
		consumer.WithLogger(c.logger),
		consumer.WithHandler(h))

	return
}

func (c *ReliableRabbitConsumer) Reset() {
	// must delete all
	log.Info().Msg("reliable rabbit consumer reset")
	if c.cleanFunc != nil {
		err := c.cleanFunc(c.channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to clean")
		}
	}
	if c.initFunc != nil {
		err := c.initFunc(c.channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to init")
		}
	}
}

func (c *ReliableRabbitConsumer) Stop() {
	c.consumer.Close()
	c.dailer.Close()
}

func (c *ReliableRabbitConsumer) Ack(deliveryTag uint64, multiple bool) error {
	return c.channel.Ack(deliveryTag, multiple)
}

func (c *ReliableRabbitConsumer) Nack(deliveryTag uint64, multiple bool) error {
	return c.channel.Nack(deliveryTag, multiple, true)
}
