package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/JhonyDev09/trackingsys/parser/internal/config"
	"github.com/JhonyDev09/trackingsys/parser/internal/consume"
	"github.com/JhonyDev09/trackingsys/parser/internal/process"
	"github.com/JhonyDev09/trackingsys/parser/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	st, err := store.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("no se pudo conectar a postgres: %v", err)
	}
	defer st.Close()

	consumer, err := consume.New(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		log.Fatalf("no se pudo conectar a rabbitmq: %v", err)
	}
	defer consumer.Close()

	log.Printf("parser escuchando en la cola %q", cfg.QueueName)

	if err := consumer.Run(ctx, process.NewHandler(st)); err != nil && ctx.Err() == nil {
		log.Fatalf("error del consumidor: %v", err)
	}
	log.Println("parser detenido")
}
