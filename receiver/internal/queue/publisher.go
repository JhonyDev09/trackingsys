// Package queue desacopla al receiver de a dónde van los mensajes que
// recibe. Hoy: consola. Después: RabbitMQ — sin tocar internal/ingest.
package queue

import "github.com/JhonyDev09/trackingsys/shared/messages"

// Publisher es implementado por cada destino posible (consola,
// RabbitMQ, etc.)
type Publisher interface {
	Publish(msg messages.RawMessage) error
	Close() error
}
