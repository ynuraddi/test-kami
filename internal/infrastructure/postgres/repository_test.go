package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal/domain"
)

func Test_CreateReservation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)

	// closing after check all expectations were met
	defer mock.Close()
	defer assert.NoError(t, mock.ExpectationsWereMet())

	repo := NewReservations(mock)

	from := time.Now().Truncate(time.Second).UTC()
	to := from.Add(1 * time.Minute)

	targetQuery := "insert into reservations"

	unexpectedError := errors.New("unexpected error")

	type args struct {
		rid domain.RoomID
		tr  domain.TimeRange
	}

	defaultArgs := args{
		rid: "1",
		tr: domain.TimeRange{
			Start: from,
			End:   to,
		},
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
				mock.ExpectExec(targetQuery).
					WithArgs(&defaultArgs.rid, &defaultArgs.tr.Start, &defaultArgs.tr.End).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			checkResult: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "NOT OK error unexpected",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectExec(targetQuery).
					WithArgs(&defaultArgs.rid, &defaultArgs.tr.Start, &defaultArgs.tr.End).
					WillReturnError(unexpectedError) // note
			},
			checkResult: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()
			err := repo.Create(context.Background(), tc.args.rid, tc.args.tr)
			tc.checkResult(t, err)
		})
	}
}

func Test_ListByRoom(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)

	// closing after check all expectations were met
	defer mock.Close()
	defer assert.NoError(t, mock.ExpectationsWereMet())

	repo := NewReservations(mock)

	from := time.Now().Truncate(time.Second).UTC()
	to := from.Add(1 * time.Minute)

	targetQuery := "select id, room_id, start_time, end_time from reservations"

	defaultRoomID := domain.RoomID("1")

	reservationsColumns := []string{"id", "room_id", "start_time", "end_time"}
	defaultReservation := domain.Reservation{
		ID:     1,
		RoomID: defaultRoomID,
		TimeRange: domain.TimeRange{
			Start: from,
			End:   to,
		},
	}

	unexpectedError := errors.New("unexpected error")

	type args struct {
		rid domain.RoomID
	}

	defaultArgs := args{
		rid: defaultRoomID,
	}

	testCases := []struct {
		name        string
		args        args
		buildStubs  func()
		checkResult func(t *testing.T, rs []domain.Reservation, err error)
	}{
		{
			name: "OK",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectQuery(targetQuery).
					RowsWillBeClosed().
					WithArgs(&defaultArgs.rid).
					WillReturnRows(pgxmock.NewRows(reservationsColumns).
						AddRow(
							defaultReservation.ID,
							defaultReservation.RoomID,
							defaultReservation.TimeRange.Start,
							defaultReservation.TimeRange.End,
						))
			},
			checkResult: func(t *testing.T, rs []domain.Reservation, err error) {
				assert.NoError(t, err)
				assert.True(t, len(rs) == 1)
				assert.Equal(t, defaultReservation, rs[0])
			},
		},
		{
			name: "NOT OK error unexpected",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectQuery(targetQuery).
					RowsWillBeClosed().
					WithArgs(&defaultArgs.rid).
					WillReturnError(unexpectedError) // note
			},
			checkResult: func(t *testing.T, rs []domain.Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
				assert.Empty(t, rs)
			},
		},
		{
			name: "NOT OK error scan",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectQuery(targetQuery).
					RowsWillBeClosed().
					WithArgs(&defaultArgs.rid).
					WillReturnRows(pgxmock.NewRows(reservationsColumns).
						AddRow(
							"incorrect value",
							defaultReservation.RoomID,
							defaultReservation.TimeRange.Start,
							defaultReservation.TimeRange.End,
						))
			},
			checkResult: func(t *testing.T, rs []domain.Reservation, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "NOT OK error rows",
			args: defaultArgs,
			buildStubs: func() {
				mock.ExpectQuery(targetQuery).
					RowsWillBeClosed().
					WithArgs(&defaultArgs.rid).
					WillReturnRows(pgxmock.NewRows(reservationsColumns).
						RowError(0, unexpectedError))
			},
			checkResult: func(t *testing.T, rs []domain.Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, unexpectedError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()
			reservations, err := repo.ListByRoom(context.Background(), tc.args.rid)
			tc.checkResult(t, reservations, err)
		})
	}
}
