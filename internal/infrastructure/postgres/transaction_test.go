package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal/domain"
)

func Test_Transaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)

	defer mock.Close()
	defer assert.NoError(t, mock.ExpectationsWereMet())

	transaction := NewTxManager(mock)

	// если он не будет использовать внедреннуб транзакцию то он запникует
	// из-за nil
	repo := NewReservations(nil)

	unexpectedError := errors.New("unexpected error")
	someErr := errors.New("some error")

	reservationsColumns := []string{"id", "room_id", "start_time", "end_time"}

	type args struct {
		do     func(txCtx context.Context) error
		option pgx.TxOptions
	}

	defaultOptions := pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	}

	defaultArgs := args{
		do:     func(txCtx context.Context) error { return nil },
		option: defaultOptions,
	}

	defaultRoomID := domain.RoomID("1")

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
				mock.ExpectBeginTx(defaultOptions)
				mock.ExpectCommit()
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "OK with repo call",
			args: args{
				option: defaultOptions,
				do: func(txCtx context.Context) error {
					_, err := repo.ListByRoom(txCtx, defaultRoomID)
					return err
				},
			},
			buildStubs: func() {
				mock.ExpectBeginTx(defaultOptions)
				mock.ExpectQuery("select id, room_id, start_time, end_time from reservations").
					WithArgs(&defaultRoomID).
					WillReturnRows(pgxmock.NewRows(reservationsColumns)) // note len zero
				mock.ExpectCommit()
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "NOT OK with repo call",
			args: args{
				option: defaultOptions,
				do: func(txCtx context.Context) error {
					_, err := repo.ListByRoom(txCtx, "1")
					return err
				},
			},
			buildStubs: func() {
				mock.ExpectBeginTx(defaultOptions)
				mock.ExpectQuery("select id, room_id, start_time, end_time from reservations").
					WithArgs(&defaultRoomID).
					WillReturnError(unexpectedError) // note len zero
				mock.ExpectRollback()
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
		{
			name: "NOT OK error begin tx",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectBeginTx(defaultOptions).WillReturnError(unexpectedError)
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
		{
			name: "NOT OK error rollback",
			args: args{
				option: defaultOptions,
				do:     func(txCtx context.Context) error { return someErr }, // note
			},
			buildStubs: func() {
				mock.ExpectBeginTx(defaultOptions)
				mock.ExpectRollback().WillReturnError(unexpectedError) // note
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, someErr)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()
			err := transaction.Execute(context.Background(), tc.args.do, tc.args.option)
			tc.checkResult(t, err)
		})
	}
}
