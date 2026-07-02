package main

import (
	"log"

	"github.com/JhonyDev09/trackingsys/receiver/internal/config"
	"github.com/JhonyDev09/trackingsys/receiver/internal/ingest"
	"github.com/JhonyDev09/trackingsys/receiver/internal/queue"
)

func main() {
	cfg := config.Load()

	// Por ahora siempre consola. Cuando quieras RabbitMQ:
	//   go build -tags rabbitmq ./cmd/receiver
	// y cambia esta línea para usar queue.NewRabbitMQPublisher(cfg.RabbitMQURL).
	var publisher queue.Publisher = queue.NewStdoutPublisher()
	defer publisher.Close()

	server := ingest.New(cfg.ListenAddr, publisher)
	if err := server.Run(); err != nil {
		log.Fatalf("error del servidor: %v", err)
	}
}
