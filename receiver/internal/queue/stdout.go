package queue

import (
	"encoding/json"
	"log"

	"github.com/JhonyDev09/trackingsys/shared/messages"
)

// StdoutPublisher imprime cada mensaje como JSON en consola. Úsalo
// mientras pruebas con el dispositivo físico, antes de meter RabbitMQ.
type StdoutPublisher struct{}

func NewStdoutPublisher() *StdoutPublisher {
	return &StdoutPublisher{}
}

func (p *StdoutPublisher) Publish(msg messages.RawMessage) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.Printf("[MSG] %s", string(b))
	return nil
}

func (p *StdoutPublisher) Close() error { return nil }
