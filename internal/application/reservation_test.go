package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal"
	mock_application "github.com/ynuraddi/test-kami/internal/application/mock"
	"github.com/ynuraddi/test-kami/internal/domain"
	mock_domain "github.com/ynuraddi/test-kami/internal/domain/mock"
)

func Test_CreateReservation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txManager := mock_application.NewMockTransaction(ctrl)
	repo := mock_domain.NewMockReservationRepository(ctrl)

	service := NewReservationService(repo, txManager)

	now := time.Now().Truncate(time.Second).UTC()

	unexpectedError := errors.New("unexpected error")

	type args struct {
		ctx      context.Context
		roomID   string
		from, to time.Time
	}

	defaultArgs := args{
		ctx:    context.Background(),
		roomID: "room",
		from:   now,
		to:     now.Add(1 * time.Hour),
	}

	testCases := []struct {
		name        string
		args        args
		buildStubs  func()
		checkResult func(t *testing.T, err error)
	}{
		{
			name: "OK",
			args: defaultArgs,
			buildStubs: func() {
				c1 := txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, f func(txCtx context.Context) error, txOptions pgx.TxOptions) error {
						return f(ctx)
					},
				).Times(1)

				c2 := repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Return(nil, nil).Times(1)

				c3 := repo.EXPECT().Create(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
					gomock.Eq(domain.TimeRange{Start: defaultArgs.from, End: defaultArgs.to}),
				).Return(nil).Times(1)

				c2.After(c1)
				c3.After(c2)
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "validation error room id",
			args: args{
				ctx:    defaultArgs.ctx,
				roomID: "", // note
				from:   defaultArgs.from,
				to:     defaultArgs.to,
			},
			buildStubs: func() {
				txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().ListByRoom(gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
			},
		},
		{
			name: "validation error time range",
			args: args{
				ctx:    defaultArgs.ctx,
				roomID: defaultArgs.roomID,
				from:   defaultArgs.to,   // note
				to:     defaultArgs.from, // note
			},
			buildStubs: func() {
				txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().ListByRoom(gomock.Any(), gomock.Any()).Times(0)
				repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
			},
		},
		{
			name: "unexpected error from ListByRoom",
			args: defaultArgs,
			buildStubs: func() {
				c1 := txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, f func(txCtx context.Context) error, txOptions pgx.TxOptions) error {
						return f(ctx)
					},
				).Times(1)

				c2 := repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Return(nil, unexpectedError).Times(1) // note

				c3 := repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				c2.After(c1)
				c3.After(c2)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
		{
			name: "unexpected error from Create",
			args: defaultArgs,
			buildStubs: func() {
				c1 := txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, f func(txCtx context.Context) error, txOptions pgx.TxOptions) error {
						return f(ctx)
					},
				).Times(1)

				c2 := repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Return(nil, nil).Times(1)

				c3 := repo.EXPECT().Create(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
					gomock.Eq(domain.TimeRange{Start: defaultArgs.from, End: defaultArgs.to}),
				).Return(unexpectedError).Times(1) // note

				c2.After(c1)
				c3.After(c2)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
		{
			name: "reservation time range conflict error",
			args: defaultArgs,
			buildStubs: func() {
				c1 := txManager.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, f func(txCtx context.Context) error, txOptions pgx.TxOptions) error {
						return f(ctx)
					},
				).Times(1)

				c2 := repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Return([]domain.Reservation{
					{
						TimeRange: domain.TimeRange{
							Start: defaultArgs.from, // note
							End:   defaultArgs.to,   // note
						},
					},
				}, nil).Times(1)

				c3 := repo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				c2.After(c1)
				c3.After(c2)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, &domain.ReservationConflictError{})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()
			err := service.ReserveRoom(tc.args.ctx, tc.args.roomID, tc.args.from, tc.args.to)
			tc.checkResult(t, err)
		})
	}

}

func Test_ListByRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txManager := mock_application.NewMockTransaction(ctrl)
	repo := mock_domain.NewMockReservationRepository(ctrl)

	service := NewReservationService(repo, txManager)

	now := time.Now()

	unexpectedError := errors.New("unexpected error")

	type args struct {
		ctx    context.Context
		roomID string
	}

	defaultArgs := args{
		ctx:    context.Background(),
		roomID: "room",
	}

	defaultReservations := []domain.Reservation{
		{
			ID:     1,
			RoomID: "1",
			TimeRange: domain.TimeRange{
				Start: now,
				End:   now.Add(1 * time.Minute),
			},
		},
	}

	testCases := []struct {
		name        string
		args        args
		buildStubs  func()
		checkResult func(t *testing.T, reservations []domain.Reservation, err error)
	}{
		{
			name: "OK",
			args: defaultArgs,
			buildStubs: func() {
				repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Times(1).Return(defaultReservations, nil)
			},
			checkResult: func(t *testing.T, reservations []domain.Reservation, err error) {
				assert.NoError(t, err)
				assert.Equal(t, defaultReservations, reservations)
			},
		},
		{
			name: "OK with empty reservations",
			args: defaultArgs,
			buildStubs: func() {
				repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Times(1).Return(nil, nil) // note
			},
			checkResult: func(t *testing.T, reservations []domain.Reservation, err error) {
				assert.NoError(t, err)
				assert.Nil(t, reservations)
			},
		},
		{
			name: "validation error room id",
			args: args{
				ctx:    defaultArgs.ctx,
				roomID: "", // note
			},
			buildStubs: func() {
				repo.EXPECT().ListByRoom(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, reservations []domain.Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Nil(t, reservations)
			},
		},
		{
			name: "unexpected error from ListByRoom",
			args: defaultArgs,
			buildStubs: func() {
				repo.EXPECT().ListByRoom(
					gomock.Any(),
					gomock.Eq(domain.RoomID(defaultArgs.roomID)),
				).Times(1).Return(nil, unexpectedError)
			},
			checkResult: func(t *testing.T, reservations []domain.Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Nil(t, reservations)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()
			reservation, err := service.ListByRoom(tc.args.ctx, tc.args.roomID)
			tc.checkResult(t, reservation, err)
		})
	}
}
