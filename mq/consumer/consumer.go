package consumer

import (
	"context"
	"github.com/latifrons/amqpextra"
	"github.com/latifrons/amqpextra/consumer"
	"github.com/latifrons/amqpextra/logger"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type ExchangeArgs struct {
	ExchangeName string
	RoutingKey   string
}

type DeclaredQueueArgs struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp091.Table
}

type ConsumerArgs struct {
	Consumer  string
	QueueName string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp091.Table
}

type ConsumerOption func(*ReliableRabbitConsumer)

type ReliableRabbitConsumer struct {
	URL               string
	ExchangeArgs      ExchangeArgs
	DeclaredQueueArgs DeclaredQueueArgs
	ConsumerArgs      ConsumerArgs
	HandleFunc        func(ctx context.Context, msg amqp091.Delivery) interface{}
	logger            logger.Logger
	dailer            *amqpextra.Dialer
	consumer          *consumer.Consumer
	prefetchCount     int
	global            bool
}

func NewReliableRabbitConsumer(url string, handleFunc func(ctx context.Context, msg amqp091.Delivery) interface{}, opts ...ConsumerOption) *ReliableRabbitConsumer {
	c := &ReliableRabbitConsumer{
		URL:           url,
		HandleFunc:    handleFunc,
		prefetchCount: 1,
		global:        false,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithLogger(l logger.Logger) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.logger = l
	}
}

func WithDeclaredQueueArgs(args DeclaredQueueArgs) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.DeclaredQueueArgs = args
	}
}
func WithConsumerArgs(args ConsumerArgs) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.ConsumerArgs = args
	}
}
func WithExchangeArgs(args ExchangeArgs) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.ExchangeArgs = args
	}
}

func WithQos(prefetchCount int, global bool) ConsumerOption {
	return func(c *ReliableRabbitConsumer) {
		c.prefetchCount = prefetchCount
		c.global = global
	}
}

func (c *ReliableRabbitConsumer) Start() (err error) {
	//ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancelFunc()

	dialerChannel := make(chan amqpextra.State, 10)
	consumerChannel := make(chan consumer.State, 10)

	c.dailer, err = amqpextra.NewDialer(
		amqpextra.WithURL(c.URL),
		//amqpextra.WithContext(ctx),
		amqpextra.WithLogger(c.logger),
		amqpextra.WithNotify(dialerChannel),
	)
	if err != nil {
		return
	}

	h := consumer.HandlerFunc(c.HandleFunc)

	c.consumer, err = c.dailer.Consumer(
		consumer.WithNotify(consumerChannel),
		consumer.WithExchange(c.ExchangeArgs.ExchangeName, c.ExchangeArgs.RoutingKey),
		consumer.WithDeclareQueue(c.DeclaredQueueArgs.Name, c.DeclaredQueueArgs.Durable, c.DeclaredQueueArgs.AutoDelete, c.DeclaredQueueArgs.Exclusive, c.DeclaredQueueArgs.NoWait, c.DeclaredQueueArgs.Args),
		//consumer.WithQueue(c.ConsumerArgs.QueueName),
		consumer.WithLogger(c.logger),
		consumer.WithHandler(h),
		consumer.WithQos(c.prefetchCount, c.global),
		consumer.WithConsumeArgs(c.ConsumerArgs.Consumer,
			c.ConsumerArgs.AutoAck,
			c.ConsumerArgs.Exclusive,
			c.ConsumerArgs.NoLocal,
			c.ConsumerArgs.NoWait,
			c.ConsumerArgs.Args))

	go func() {
		for {
			select {
			case v := <-dialerChannel:
				log.Info().Interface("v", v).Msg("dialer updates")
			case v := <-consumerChannel:
				log.Info().Interface("v", v).Msg("consumer updates")
			}
		}

	}()
	return
}

func (c *ReliableRabbitConsumer) Stop() {
	c.consumer.Close()
	c.dailer.Close()
}
