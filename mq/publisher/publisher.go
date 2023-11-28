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

	dailer     *amqpextra.Dialer
	connection *amqp091.Connection
	channel    *amqp091.Channel
	publisher  *publisher.Publisher
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
	if c.cleanFunc != nil {
		err := c.cleanFunc(c.channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to clean")
		}
		return err
	}
	if c.initFunc != nil {
		err := c.initFunc(c.channel)
		if err != nil {
			log.Error().Err(err).Msg("failed to init")
		}
		return err
	}
	return
}

func (c *ReliableRabbitPublisher) Stop() {
	c.publisher.Close()
	c.dailer.Close()
}
