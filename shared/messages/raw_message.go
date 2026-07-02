// Package messages define los contratos de mensajes que viajan entre
// el receiver y el parser. Ambos servicios importan este paquete para
// no arriesgarse a que sus structs se desincronicen con el tiempo.
package messages

import "time"

// RawMessage es lo que el receiver publica hacia la cola, y lo que el
// parser consume para decodificar y escribir en gps_posicion,
// gps_ultima_posicion y gps_evento.
type RawMessage struct {
	IMEI       string    `json:"imei"`
	Protocol   string    `json:"protocol"`
	MsgType    string    `json:"msg_type"`
	Raw        string    `json:"raw"`
	RemoteAddr string    `json:"remote_addr"`
	ReceivedAt time.Time `json:"received_at"`
}
