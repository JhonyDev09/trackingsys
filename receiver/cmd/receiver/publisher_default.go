//go:build !rabbitmq

package main

import (
	"log"

	"github.com/JhonyDev09/trackingsys/receiver/internal/queue"
)

func newPublisher(rabbitMQURL string) (queue.Publisher, error) {
	if rabbitMQURL != "" {
		log.Printf("advertencia: RABBITMQ_URL está configurado pero este binario se compiló sin soporte RabbitMQ (usa: go build -tags rabbitmq). Usando consola.")
	}
	return queue.NewStdoutPublisher(), nil
}
