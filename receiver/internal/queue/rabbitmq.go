//go:build rabbitmq

package queue

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/JhonyDev09/trackingsys/shared/messages"
)

// RabbitMQPublisher publica cada mensaje crudo en gps.exchange, con
// routing key gps.raw.<msg_type> (gps.raw.position, gps.raw.login, etc.)
// Compilar con: go build -tags rabbitmq
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
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
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &RabbitMQPublisher{conn: conn, channel: ch}, nil
}

func (p *RabbitMQPublisher) Publish(msg messages.RawMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.channel.Publish(
		"gps.exchange",
		"gps.raw."+msg.MsgType,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	log.Printf("[DEBUG] publish imei=%s routingKey=%s err=%v", msg.IMEI, routingKey, err)
	return err
}

func (p *RabbitMQPublisher) Close() error {
	p.channel.Close()
	return p.conn.Close()
}
