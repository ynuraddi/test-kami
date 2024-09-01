package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ynuraddi/test-kami/internal/domain"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type reservations struct {
	conn DBTX
}

func NewReservations(conn DBTX) *reservations {
	return &reservations{
		conn: conn,
	}
}

func (r reservations) Create(ctx context.Context, roomID domain.RoomID, timeRange domain.TimeRange) (err error) {
	tx := solveTx(r.conn, ctx)

	query := `insert into reservations(room_id, start_time, end_time)
	values($1, $2, $3)`

	if _, err := tx.Exec(ctx, query, &roomID, &timeRange.Start, &timeRange.End); err != nil {
		return err
	}
	return nil
}

func (r reservations) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Reservation, error) {
	tx := solveTx(r.conn, ctx)

	query := `select id, room_id, start_time, end_time from reservations
	where room_id = $1`

	rows, err := tx.Query(ctx, query, &roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		if err := rows.Scan(
			&reservation.ID,
			&reservation.RoomUUID,
			&reservation.TimeRange.Start,
			&reservation.TimeRange.End,
		); err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}
