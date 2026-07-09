package main

import (
	"log"

	"github.com/JhonyDev09/trackingsys/receiver/internal/config"
	"github.com/JhonyDev09/trackingsys/receiver/internal/ingest"
)

func main() {
	cfg := config.Load()

	// Por ahora siempre consola. Cuando quieras RabbitMQ:
	//   go build -tags rabbitmq ./cmd/receiver
	// y cambia esta línea para usar queue.NewRabbitMQPublisher(cfg.RabbitMQURL).
	publisher, err := newPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("no se pudo iniciar el publisher: %v", err)
	}
	defer publisher.Close()

	server := ingest.New(cfg.ListenAddr, publisher)
	if err := server.Run(); err != nil {
		log.Fatalf("error del servidor: %v", err)
	}
}
