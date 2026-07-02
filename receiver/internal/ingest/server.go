// Package ingest contiene el servidor TCP que acepta conexiones de los
// dispositivos GPS y traduce cada frame en un mensaje publicado en la
// cola.
package ingest

import (
	"bufio"
	"log"
	"net"
	"time"

	"github.com/JhonyDev09/trackingsys/receiver/internal/queue"
	"github.com/JhonyDev09/trackingsys/shared/messages"
	"github.com/JhonyDev09/trackingsys/shared/protocol/coban"
)

// Server escucha conexiones TCP y despacha cada frame al publisher.
type Server struct {
	ListenAddr string
	Publisher  queue.Publisher
}

func New(listenAddr string, publisher queue.Publisher) *Server {
	return &Server{ListenAddr: listenAddr, Publisher: publisher}
}

// Run arranca el listener y bloquea aceptando conexiones. Cada
// conexión se maneja en su propio goroutine.
func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	log.Printf("receiver escuchando en %s (protocolo: Coban)", s.ListenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error aceptando conexión: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("nueva conexión desde %s", remoteAddr)

	defer func() {
		conn.Close()
		log.Printf("conexión cerrada: %s", remoteAddr)
	}()

	reader := bufio.NewReader(conn)
	imei := "" // se conoce después del primer login

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		frame, err := coban.ReadFrame(reader)
		if err != nil {
			log.Printf("[%s] fin de conexión o error de lectura: %v", remoteAddr, err)
			return
		}
		if frame == "" {
			continue
		}

		log.Printf("[%s] RAW: %s", remoteAddr, frame)

		frameIMEI, msgType, ok := coban.ParseIMEI(frame)
		if !ok {
			log.Printf("[%s] no se pudo interpretar el frame, se descarta", remoteAddr)
			continue
		}
		if frameIMEI != "" {
			imei = frameIMEI
		}

		if msgType == coban.MsgLogin {
			if _, err := conn.Write([]byte(coban.LoginAck)); err != nil {
				log.Printf("[%s] error enviando ACK de login: %v", remoteAddr, err)
			}
		}

		msg := messages.RawMessage{
			IMEI:       imei,
			Protocol:   coban.Name,
			MsgType:    msgType,
			Raw:        frame,
			RemoteAddr: remoteAddr,
			ReceivedAt: time.Now().UTC(),
		}
		if err := s.Publisher.Publish(msg); err != nil {
			log.Printf("[%s] error publicando mensaje: %v", remoteAddr, err)
		}
	}
}
