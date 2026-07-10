package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool

	imeiCachemu sync.RWMutex
	imeiCache   map[string]int
}

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("Conectado a la DB: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Ping a la DB: %w", err)
	}
	return &Store{pool: pool, imeiCache: make(map[string]int)}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

// ErrGPSNoRegistrado indica que el IMEI no existe en la tabla `gps`.
// El parser no auto-registra dispositivos nuevos — eso es una decisión
// administrativa que debe pasar por el API, no ocurrir en silencio
// aquí. Mientras tanto, estos mensajes se descartan con log visible.

type ErrGPSNoRegistrado struct{ IMEI string }

func (e ErrGPSNoRegistrado) Error() string {
	return fmt.Sprintf("El IMEI %s no esta registrado en la base de datos", e.IMEI)
}

func (s *Store) IDByIMEI(ctx context.Context, imei string) (int, error) {
	s.imeiCachemu.RLock()
	id, ok := s.imeiCache[imei]
	s.imeiCachemu.RUnlock()
	if ok {
		return id, nil
	}

	err := s.pool.QueryRow(ctx, `SELECT id_gps FROM gps WHERE imei = $1`, imei).Scan(&id)
	if err != nil {
		return 0, ErrGPSNoRegistrado{IMEI: imei}
	}

	s.imeiCachemu.Lock()
	s.imeiCache[imei] = id
	s.imeiCachemu.Unlock()
	return id, nil
}

// PositionWrite es lo que el parser ya decodificó y quiere persistir.
type PositionWrite struct {
	IDGPS     int
	Latitude  float64
	Longitude float64
	SpeedKmh  float64
	Course    float64
	Timestamp time.Time
}

// InsertPosicion agrega una fila al histórico particionado.
func (s *Store) InsertPosicion(ctx context.Context, p PositionWrite) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO gps_posicion (id_gps, ubicacion, velocidad, rumbo, altitud, fecha_posicion)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5, NULL, $6)
	`, p.IDGPS, p.Longitude, p.Latitude, p.SpeedKmh, p.Course, p.Timestamp)
	if err != nil {
		return fmt.Errorf("insertando gps_posicion: %w", err)
	}
	return nil
}

// UpsertUltimaPosicion actualiza la posición actual del GPS, solo si
// el mensaje entrante es más reciente que lo que ya había guardado
// (protege contra mensajes que llegan fuera de orden).
func (s *Store) UpsertUltimaPosicion(ctx context.Context, p PositionWrite) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO gps_ultima_posicion (id_gps, ubicacion, velocidad, rumbo, altitud, fecha_posicion)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5, NULL, $6)
		ON CONFLICT (id_gps) DO UPDATE SET
			ubicacion = EXCLUDED.ubicacion,
			velocidad = EXCLUDED.velocidad,
			rumbo = EXCLUDED.rumbo,
			fecha_posicion = EXCLUDED.fecha_posicion,
			updated_at = NOW()
		WHERE EXCLUDED.fecha_posicion > gps_ultima_posicion.fecha_posicion
	`, p.IDGPS, p.Longitude, p.Latitude, p.SpeedKmh, p.Course, p.Timestamp)
	if err != nil {
		return fmt.Errorf("actualizando gps_ultima_posicion: %w", err)
	}
	return nil
}
