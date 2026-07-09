//go:build rabbitmq

package main

import "github.com/JhonyDev09/trackingsys/receiver/internal/queue"

func newPublisher(rabbitMQURL string) (queue.Publisher, error) {
	if rabbitMQURL == "" {
		return queue.NewStdoutPublisher(), nil
	}
	return queue.NewRabbitMQPublisher(rabbitMQURL)
}
