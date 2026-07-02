// Package coban implementa el parseo mínimo del protocolo Coban
// (TK103/GPS103): extracción de IMEI, clasificación de mensajes y el
// framing de texto delimitado por ';'.
//
// La decodificación completa de lat/lng/velocidad no vive aquí — eso
// lo hace el servicio parser a partir del mismo Raw que el receiver
// publica sin tocar. Este paquete es compartido para que receiver y
// parser nunca tengan dos implementaciones del mismo protocolo.
package coban

import (
	"bufio"
	"strings"
)

// Name identifica este protocolo en los mensajes publicados a la cola.
const Name = "coban"

// LoginAck es lo que muchos firmwares TK103/GPS103 esperan tras el
// login para considerar la sesión activa. Verifícalo con tu dispositivo
// específico — no todos los firmwares lo requieren.
const LoginAck = "LOAD"

// Tipos de mensaje que ParseIMEI puede devolver.
const (
	MsgLogin     = "login"
	MsgPosition  = "position"
	MsgHeartbeat = "heartbeat"
	MsgAlarmSOS  = "alarm_sos"
	MsgUnknown   = "unknown"
)

// ReadFrame lee bytes del stream hasta encontrar ';', el delimitador
// que usa Coban para cerrar cada mensaje.
func ReadFrame(reader *bufio.Reader) (string, error) {
	data, err := reader.ReadString(';')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(data), nil
}

// ParseIMEI extrae el IMEI y clasifica el frame recibido.
//
// Ejemplos de frames Coban:
//
//	Login:    ##,imei:359586015829802,A;
//	Posición: imei:359586015829802,tracker,1501011035,,F,103552.000,A,2233.9058,N,11404.9536,E,0.00,183.65,;
//	SOS:      imei:359586015829802,help me,...;
func ParseIMEI(raw string) (imei string, msgType string, ok bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ";")

	switch {
	case strings.HasPrefix(raw, "##,imei:"):
		rest := strings.TrimPrefix(raw, "##,imei:")
		imei = strings.SplitN(rest, ",", 2)[0]
		return imei, MsgLogin, imei != ""

	case strings.HasPrefix(raw, "imei:"):
		rest := strings.TrimPrefix(raw, "imei:")
		parts := strings.SplitN(rest, ",", 2)
		imei = parts[0]
		if imei == "" {
			return "", "", false
		}
		if len(parts) < 2 {
			return imei, MsgUnknown, true
		}
		switch {
		case strings.Contains(parts[1], "help me"):
			return imei, MsgAlarmSOS, true
		case strings.HasPrefix(parts[1], "tracker"):
			return imei, MsgPosition, true
		case strings.HasPrefix(parts[1], "heartbeat"):
			return imei, MsgHeartbeat, true
		default:
			return imei, MsgUnknown, true
		}
	}
	return "", "", false
}
