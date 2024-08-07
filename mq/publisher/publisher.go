package publisher

import (
	"context"
	"github.com/latifrons/amqpextra"
	"github.com/latifrons/amqpextra/logger"
	pp "github.com/latifrons/amqpextra/publisher"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type DeclareExchangeArgs struct {
	ExchangeName string
	Kind         string
	Durable      bool
	AutoDelete   bool
	Internal     bool
	NoWait       bool
	Args         map[string]interface{}
}

type PublisherOption func(*ReliableRabbitPublisher)

type ReliableRabbitPublisher struct {
	URL                 string
	DeclareExchangeArgs DeclareExchangeArgs

	logger    logger.Logger
	dailer    *amqpextra.Dialer
	publisher *pp.Publisher
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

func WithDeclareExchangeArgs(args DeclareExchangeArgs) PublisherOption {
	return func(c *ReliableRabbitPublisher) {
		c.DeclareExchangeArgs = args
	}
}

func WithLogger(logger2 logger.Logger) PublisherOption {
	return func(c *ReliableRabbitPublisher) {
		c.logger = logger2
	}
}

func WithInitFunc(args func(channel *amqp091.Channel) (err error)) PublisherOption {
	return func(c *ReliableRabbitPublisher) {
	}
}

func (c *ReliableRabbitPublisher) Reset() error {
	return nil
}

func (c *ReliableRabbitPublisher) Start() (err error) {
	dialerChannel := make(chan amqpextra.State, 10)

	c.dailer, err = amqpextra.NewDialer(amqpextra.WithURL(c.URL),
		amqpextra.WithLogger(c.logger),
		amqpextra.WithNotify(dialerChannel),
	)
	if err != nil {
		return
	}

	c.publisher, err = c.dailer.Publisher(
		pp.WithLogger(c.logger),
		pp.WithInitFunc(c.initer),
	)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case v := <-dialerChannel:
				log.Info().Interface("v", v).Msg("dialer updates")
			}
		}

	}()

	return
}

func (c *ReliableRabbitPublisher) Publish(ctx context.Context, exchange, key string, msg amqp091.Publishing) (err error) {
	return c.publisher.Publish(pp.Message{
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

func (c *ReliableRabbitPublisher) Stop() {
	c.publisher.Close()
	c.dailer.Close()
}

func (c *ReliableRabbitPublisher) initer(conn pp.AMQPConnection) (channelP pp.AMQPChannel, err error) {
	channel, err := conn.(*amqp091.Connection).Channel()
	if err != nil {
		return
	}
	// declare exchange
	err = channel.ExchangeDeclare(c.DeclareExchangeArgs.ExchangeName, c.DeclareExchangeArgs.Kind, c.DeclareExchangeArgs.Durable, c.DeclareExchangeArgs.AutoDelete, c.DeclareExchangeArgs.Internal, c.DeclareExchangeArgs.NoWait, c.DeclareExchangeArgs.Args)
	if err != nil {
		return
	}
	return channel, err
}
