package config

import "os"

// Config agrupa los parámetros del receiver, configurables por
// variable de entorno para no tocar código entre desarrollo y
// producción.
type Config struct {
	// ListenAddr es donde el receiver escucha conexiones TCP de los GPS.
	ListenAddr string
	// RabbitMQURL, si está vacío, el receiver usa el publisher de
	// consola (útil para ver los datos llegar antes de meter RabbitMQ).
	RabbitMQURL string
}

func Load() Config {
	return Config{
		ListenAddr:  getEnv("RECEIVER_LISTEN_ADDR", ":6060"),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
