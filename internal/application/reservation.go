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

type reservationService struct {
	repo domain.ReservationRepository
	tx   Transaction

	// это такой оркестратор
	// я собираюсь разделить транзакции по комнатам
	// я подумал что в рамках текущей логики комнаты никак не конфликтуют между собой
	// поэтому сериализовать можно операции по комнатам
	//
	// такое решение сложно не масштабируемое, потому что нужна согласованность между репликами
	// в таком случае лучше использовать брокер сообщений
	roomMutex *MutexManager
}

func NewReservationService(repo domain.ReservationRepository, tx Transaction) *reservationService {
	return &reservationService{
		repo: repo,
		tx:   tx,

		roomMutex: NewMutexManager(time.Minute, time.Minute),
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

	mu := s.roomMutex.GetMutex(roomID)
	mu.Lock()
	defer mu.Unlock()

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
		IsoLevel:       pgx.RepeatableRead,
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
