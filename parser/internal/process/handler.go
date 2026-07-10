// Package process conecta lo que llega de la cola con lo que se
// escribe en la base de datos. Por ahora solo maneja posiciones del
// protocolo Coban — eventos (SOS, corte de energía, etc.) y otros
// protocolos (Teltonika) se agregan aquí después, sin tocar consume
// ni store.
package process

import (
	"context"
	"errors"
	"log"

	"github.com/JhonyDev09/trackingsys/parser/internal/store"
	"github.com/JhonyDev09/trackingsys/shared/messages"
	"github.com/JhonyDev09/trackingsys/shared/protocol/coban"
	sharedcoban "github.com/JhonyDev09/trackingsys/shared/protocol/coban"
)

const knotsToKmh = 1.852

// NewHandler arma la función que consume.Consumer.Run() invoca por
// cada mensaje ya deserializado.
func NewHandler(st *store.Store) func(ctx context.Context, msg messages.RawMessage) error {
	return func(ctx context.Context, msg messages.RawMessage) error {
		log.Printf("[DEBUG] handler recibió: protocol=%q msg_type=%q imei=%s raw=%s", msg.Protocol, msg.MsgType, msg.IMEI, msg.Raw)
		switch msg.Protocol {
		case sharedcoban.Name:
			return handleCoban(ctx, st, msg)
		default:
			log.Printf("[%s] protocolo desconocido %q, se descarta", msg.IMEI, msg.Protocol)
			return nil // no reintentar: nunca vamos a saber decodificarlo
		}
	}
}

func handleCoban(ctx context.Context, st *store.Store, msg messages.RawMessage) error {
	switch msg.MsgType {
	case sharedcoban.MsgPosition:
		return handlePosition(ctx, st, msg)

	case sharedcoban.MsgLogin, sharedcoban.MsgHeartbeat:
		// nada que persistir todavía; sirven para mantener viva la
		// conexión en el receiver, no traen datos de negocio.
		return nil

	case sharedcoban.MsgAlarmSOS, sharedcoban.MsgAlarmPowerCut, sharedcoban.MsgAlarmPowerBack:
		// TODO: escribir en gps_evento una vez definamos ese flujo.
		log.Printf("[%s] evento %s recibido (pendiente de guardar en gps_evento)", msg.IMEI, msg.MsgType)
		return nil

	default:
		log.Printf("[%s] msg_type %q sin manejar todavía, se descarta", msg.IMEI, msg.MsgType)
		return nil
	}
}

func handlePosition(ctx context.Context, st *store.Store, msg messages.RawMessage) error {
	pos, err := coban.DecodePosition(msg.Raw)
	if err != nil {
		if errors.Is(err, coban.ErrNoFix) {
			// No es un error del sistema: el dispositivo reportó sin
			// fix de GPS (LBS). No hay nada que guardar en esta fila.
			log.Printf("[%s] posición sin fix de GPS, se omite", msg.IMEI)
			return nil
		}
		return err // formato inesperado: sí queremos verlo en la DLQ
	}

	idGPS, err := st.IDByIMEI(ctx, msg.IMEI)
	if err != nil {
		// El dispositivo no está registrado en la tabla `gps` — no
		// reintentar infinitamente, pero sí queremos verlo en la DLQ
		// para darnos cuenta y registrarlo.
		return err
	}

	write := store.PositionWrite{
		IDGPS:     idGPS,
		Latitude:  pos.Latitude,
		Longitude: pos.Longitude,
		SpeedKmh:  pos.SpeedKnots * knotsToKmh,
		Course:    pos.Course,
		Timestamp: pos.Timestamp,
	}

	if err := st.InsertPosicion(ctx, write); err != nil {
		return err
	}
	if err := st.UpsertUltimaPosicion(ctx, write); err != nil {
		return err
	}

	log.Printf("[%s] posición guardada: lat=%.6f lon=%.6f", msg.IMEI, pos.Latitude, pos.Longitude)
	return nil
}
