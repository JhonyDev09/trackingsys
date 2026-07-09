package coban

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var ErrNoFix = errors.New("coban: frame sin fix de GPS")

type Position struct {
	Latitude   float64 // grados decimales, positivo = norte
	Longitude  float64 // grados decimales, positivo = este
	SpeedKnots float64
	Course     float64   // grados respecto al norte, 0-359
	Timestamp  time.Time // UTC, tomado del propio GPS (no la hora local del dispositivo)
}

func DecodePosition(raw string) (Position, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ";")

	rest := strings.TrimPrefix(raw, "imei:")
	parts := strings.SplitN(rest, ",", 2)
	if len(parts) < 2 || !strings.HasPrefix(parts[1], "tracker") {
		return Position{}, errors.New("coban: no es un frame de tipo tracker")
	}

	fields := strings.Split(parts[1], ",")
	// fields[0]  = "tracker"
	// fields[1]  = fecha local (no usada, usamos la hora UTC del campo 4)
	// fields[2]  = vacío
	// fields[3]  = indicador de fix: "F" (GPS) o "L" (LBS, sin posición confiable)
	// fields[4]  = hora UTC HHMMSS.ss
	// fields[5]  = validez: "A" válido, "V" inválido
	// fields[6]  = latitud DDMM.MMMMM
	// fields[7]  = N/S
	// fields[8]  = longitud DDDMM.MMMMM
	// fields[9]  = E/W
	// fields[10] = velocidad en nudos
	// fields[11] = rumbo
	if len(fields) < 12 {
		return Position{}, errors.New("Coban: frame tracker con muy pocos daatos")
	}

	fixIndicator := fields[3]
	validity := fields[5]
	if fixIndicator != "F" || validity != "A" {
		return Position{}, ErrNoFix
	}

	lat, err := parseDMM(fields[6], fields[7])
	if err != nil {
		return Position{}, err
	}

	lon, err := parseDMM(fields[8], fields[9])
	if err != nil {
		return Position{}, err
	}

	speed, err := strconv.ParseFloat(fields[10], 64)
	if err != nil {
		speed = 0
	}

	course, err := strconv.ParseFloat(fields[11], 64)
	if err != nil {
		course = 0
	}

	ts, err := parseUTCTime(fields[4])
	if err != nil {
		return Position{}, err
	}

	return Position{
		Latitude:   lat,
		Longitude:  lon,
		SpeedKnots: speed,
		Course:     course,
		Timestamp:  ts,
	}, nil
}

// parseDMM convierte el formato grados+minutos decimales de Coban
// (DDMM.MMMMM o DDDMM.MMMMM) a grados decimales, aplicando signo
// según el hemisferio (S y W son negativos).
func parseDMM(value, hemisphere string) (float64, error) {
	raw, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	degrees := float64(int(raw / 100))
	minutes := raw - degrees*100
	decimal := degrees + minutes/100

	switch hemisphere {
	case "S", "W":
		decimal = -decimal
	}
	return decimal, nil
}

func parseUTCTime(hhmmss string) (time.Time, error) {
	t, err := time.Parse("150405.00", hhmmss)
	if err != nil {
		t, err = time.Parse("150405", hhmmss)
		if err != nil {
			return time.Time{}, err
		}
	}
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil

}
