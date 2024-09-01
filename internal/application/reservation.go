package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ynuraddi/test-kami/internal/domain"
)

type Transaction interface {
	Execute(ctx context.Context, f func(txCtx context.Context) error, options pgx.TxOptions) error
}

type ReservationRepository interface {
	Create(ctx context.Context, roomID domain.RoomID, timeRange domain.TimeRange) error
	ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Reservation, error)
}

type reservationService struct {
	repo ReservationRepository
	tx   Transaction
}

func NewReservationService(repo ReservationRepository, tx Transaction) *reservationService {
	return &reservationService{
		repo: repo,
		tx:   tx,
	}
}

func (s reservationService) ReserveRoom(ctx context.Context, roomID string, from, to time.Time) (err error) {
	rid, err := domain.NewRoomID(roomID)
	if err != nil {
		return err
	}
	tr, err := domain.NewTimeRange(from, to)
	if err != nil {
		return err
	}

	return s.tx.Execute(ctx, func(txCtx context.Context) error {
		reservations, err := s.ListByRoom(txCtx, roomID)
		if err != nil {
			return err
		}

		// быстрее можно сделать если создать запрос на roomAndRange
		for _, r := range reservations {
			if r.TimeRange.CrossWith(tr) {
				return domain.ReservationConflictError{
					Reservation:         tr,
					ConflictReservation: r.TimeRange,
				}
			}
		}

		return s.repo.Create(txCtx, rid, tr)
	}, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
}

func (s reservationService) ListByRoom(ctx context.Context, roomID string) ([]domain.Reservation, error) {
	rid, err := domain.NewRoomID(roomID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByRoom(ctx, rid)
}
