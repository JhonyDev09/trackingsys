// Package consume conecta con RabbitMQ y entrega cada mensaje crudo a
// un Handler. No sabe nada de Postgres ni de protocolos GPS — solo
// mueve bytes de la cola hacia quien los procese.

package consume

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JhonyDev09/trackingsys/shared/messages"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, msg messages.RawMessage) error

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// New se conecta a RabbitMQ y declara exchange/cola/binding/DLQ.
// Es idempotente: si ya existen con esta misma configuración, no pasa
// nada; si existen con otra configuración incompatible, falla (lo cual
// es preferible a operar sobre una topología distinta a la esperada).

func New(url, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare("gps.exchange", "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare("gps.dlx", "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare("gps.raw.dlq", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("gps.raw.dlq", "gps.raw.failed", "gps.dlx", false, nil); err != nil {
		return nil, err
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "gps.dlx",
		"x-dead-letter-routing-key": "gps.raw.failed",
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
		return nil, err
	}
	// gps.raw.# cubre gps.raw.login, gps.raw.position, gps.raw.alarm_sos, etc.
	if err := ch.QueueBind(queueName, "gps.raw.#", "gps.exchange", false, nil); err != nil {
		return nil, err
	}

	if err := ch.Qos(50, 0, false); err != nil {
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch, queue: queueName}, nil
}

// Run consume mensajes indefinidamente, en bloqueo, hasta que ctx se
// cancele o el canal falle.

func (c *Consumer) Run(ctx context.Context, handle Handler) error {
	deliveries, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			c.process(ctx, d, handle)
		}
	}
}

func (c *Consumer) process(ctx context.Context, d amqp.Delivery, hanle Handler) {
	var msg messages.RawMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("mensaje no es JSON valdio, va a DLQ: %v", err)
		d.Nack(false, false)
		return
	}

	if err := hanle(ctx, msg); err != nil {
		log.Printf("[%s] error procesando mensaje, va a DLQ: %v", msg.IMEI, err)
		d.Nack(false, false)
		return
	}

	d.Ack(false)
}

func (c *Consumer) Close() error {
	c.channel.Close()
	return c.conn.Close()
}
