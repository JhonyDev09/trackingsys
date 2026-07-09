package config

import (
	"os"
)

type Config struct {
	RabbitMQURL string
	PostgresDSN string
	QueueName   string
}

func Load() Config {
	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://dbtrack:track#2026@localhost:5432/trackdb"),
		QueueName:   getEnv("PARSER_QUEUE_NAME", "gps.raw.queue"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
