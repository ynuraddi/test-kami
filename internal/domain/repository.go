package domain

import "context"

type ReservationRepository interface {
	Create(ctx context.Context, roomID RoomID, timeRange TimeRange) error
	ListByRoom(ctx context.Context, roomID RoomID) ([]Reservation, error)
}
