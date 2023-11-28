package main

import (
	"context"
	"fmt"
	"github.com/latifrons/latigo/mq/consumer"
	"github.com/latifrons/latigo/mq/publisher"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func p() {
	ops := publisher.NewReliableRabbitPublisher("amqp://guest:guest@localhost:5672",
		publisher.WithInitFunc(initPublisher))
	err := ops.Start()
	if err != nil {
		panic(err)
	}
	go func() {
		for i := 0; i < 100; i++ {
			log.Info().Int("i", i).Msg("publishing")
			msg := fmt.Sprintf(`%d`, i)
			err = ops.Publish(context.Background(),
				"xx_ex", "xx_key",
				amqp091.Publishing{
					Body: []byte(msg),
				})
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second)
		}
	}()
}

func initPublisher(channel *amqp091.Channel) (err error) {
	err = channel.ExchangeDeclare("xx_ex", "topic", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	return
}

func initConsumer(c *amqp091.Channel) (err error) {
	declare, err := c.QueueDeclare("xx_queue", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	err = c.QueueBind(declare.Name, "*", "xx_ex", false, nil)
	if err != nil {
		panic(err)
	}
	return
}
func cleanConsumer(c *amqp091.Channel) (err error) {
	i, err := c.QueueDelete("xx_queue", false, false, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(i)
	return
}

func c() {
	ops := consumer.NewReliableRabbitConsumer("amqp://guest:guest@localhost:5672",
		"xx_ex", "xx_queue", "xx_key",
		handler,
		consumer.WithInitFunc(initConsumer),
		consumer.WithCleanFunc(cleanConsumer))

	err := ops.Start()
	if err != nil {
		return
	}

	go func() {
		time.Sleep(time.Second * 20)
		ops.Reset()
	}()
}

func handler(ctx context.Context, msg amqp091.Delivery) interface{} {
	log.Info().Str("i", string(msg.Body)).Msg("consuming")
	v, err := strconv.ParseInt(string(msg.Body), 10, 64)
	if err != nil {
		panic(err)
	}
	if v < 10 {
		msg.Ack(false)
	}

	return nil
}

func main() {
	arg := os.Args[1]
	if arg == "p" {
		p()
	} else if arg == "c" {
		c()
	}

	select {}
}
