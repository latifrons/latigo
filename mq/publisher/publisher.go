package publisher

import (
	"context"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"time"
)

type PublisherOption func(*ReliableRabbitPublisher)

type ReliableRabbitPublisher struct {
	URL       string
	initFunc  func(channel *amqp091.Channel) error
	cleanFunc func(channel *amqp091.Channel) error
	logger    logger.Logger

	dailer    *amqpextra.Dialer
	publisher *publisher.Publisher
}

func NewReliableRabbitPublisher(url string, opts ...PublisherOption) *ReliableRabbitPublisher {
	c := &ReliableRabbitPublisher{
		URL: url,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithInitFunc(f func(*amqp091.Channel) error) PublisherOption {
	return func(c *ReliableRabbitPublisher) {
		c.initFunc = f
	}
}

func WithCleanFunc(f func(*amqp091.Channel) error) PublisherOption {
	return func(c *ReliableRabbitPublisher) {
		c.cleanFunc = f
	}
}

func (c *ReliableRabbitPublisher) Start() (err error) {
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

	c.publisher, err = c.dailer.Publisher(publisher.WithLogger(c.logger))
	return
}

func (c *ReliableRabbitPublisher) Publish(ctx context.Context, exchange, key string, msg amqp091.Publishing) (err error) {
	return c.publisher.Publish(publisher.Message{
		Context:      ctx,
		Exchange:     exchange,
		Key:          key,
		Mandatory:    false,
		Immediate:    false,
		ErrOnUnready: false,
		Publishing:   msg,
		ResultCh:     nil,
	})
}

func (c *ReliableRabbitPublisher) Reset() (err error) {
	// must delete all
	log.Info().Msg("reliable rabiit publisher reset")
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

	return
}

func (c *ReliableRabbitPublisher) Stop() {
	c.publisher.Close()
	c.dailer.Close()
}
