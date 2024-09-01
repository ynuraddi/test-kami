package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal"
)

func Test_Reservation(t *testing.T) {
	type args struct {
		id       int64
		roomId   string
		from, to time.Time
	}

	now := time.Now().Truncate(time.Second).UTC()

	defaultArgs := args{
		id:     1,
		roomId: "1",
		from:   now,
		to:     now.Add(1 * time.Minute),
	}

	testCases := []struct {
		name        string
		args        args
		checkResult func(t *testing.T, reservation Reservation, err error)
	}{
		{
			name: "OK",
			args: defaultArgs,
			checkResult: func(t *testing.T, reservation Reservation, err error) {
				assert.NoError(t, err)
				assert.Equal(t, Reservation{
					ID:       defaultArgs.id,
					RoomUUID: RoomID(defaultArgs.roomId),
					TimeRange: TimeRange{
						Start: defaultArgs.from,
						End:   defaultArgs.to,
					},
				}, reservation)
			},
		},
		{
			name: "zero id",
			args: args{
				id:     0,
				roomId: defaultArgs.roomId,
				from:   defaultArgs.from,
				to:     defaultArgs.to,
			},
			checkResult: func(t *testing.T, reservation Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, reservation)
			},
		},
		{
			name: "negative id",
			args: args{
				id:     -1,
				roomId: defaultArgs.roomId,
				from:   defaultArgs.from,
				to:     defaultArgs.to,
			},
			checkResult: func(t *testing.T, reservation Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, reservation)
			},
		},
		{
			name: "error from NewRoomID",
			args: args{
				id:     defaultArgs.id,
				roomId: "",
				from:   defaultArgs.from,
				to:     defaultArgs.to,
			},
			checkResult: func(t *testing.T, reservation Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, reservation)
			},
		},
		{
			name: "error from NewTimeRange",
			args: args{
				id:     defaultArgs.id,
				roomId: defaultArgs.roomId,
				from:   defaultArgs.to,
				to:     defaultArgs.from,
			},
			checkResult: func(t *testing.T, reservation Reservation, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, reservation)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reservation, err := NewReservation(tc.args.id, tc.args.roomId, tc.args.from, tc.args.to)
			tc.checkResult(t, reservation, err)
		})
	}
}

func Test_TimeRange(t *testing.T) {
	type args struct {
		from, to time.Time
	}

	now := time.Now()

	defaultArgs := args{
		from: now,
		to:   now.Add(1 * time.Minute),
	}

	testCases := []struct {
		name        string
		args        args
		checkResult func(t *testing.T, timeRange TimeRange, err error)
	}{
		{
			name: "OK",
			args: defaultArgs,
			checkResult: func(t *testing.T, timeRange TimeRange, err error) {
				assert.NoError(t, err)
				assert.Equal(t, TimeRange{
					Start: defaultArgs.from.Truncate(time.Second).UTC(),
					End:   defaultArgs.to.Truncate(time.Second).UTC(),
				}, timeRange)
			},
		},
		{
			name: "error start time is after end time",
			args: args{
				from: defaultArgs.to,
				to:   defaultArgs.from,
			},
			checkResult: func(t *testing.T, timeRange TimeRange, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, timeRange)
			},
		},
		{
			name: "error start time is equal end time",
			args: args{
				from: defaultArgs.from,
				to:   defaultArgs.from,
			},
			checkResult: func(t *testing.T, timeRange TimeRange, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, timeRange)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tr, err := NewTimeRange(tc.args.from, tc.args.to)
			tc.checkResult(t, tr, err)
		})
	}
}

func Test_TimeRange_IsCross(t *testing.T) {
	from := time.Now().Truncate(time.Second).UTC() // image its 12:00
	to := from.Add(1 * time.Hour)                  // image its 13:00

	baseTR, err := NewTimeRange(from, to)
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		crossingTime TimeRange
		checkResult  func(t *testing.T, isCross bool)
	}{
		{
			name: "OK",
			crossingTime: TimeRange{ // 14:00 - 15:00
				Start: to.Add(1 * time.Hour),
				End:   to.Add(2 * time.Hour),
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.False(t, isCross)
			},
		},
		{
			name: "OK equal base start time and crossing end time",
			crossingTime: TimeRange{ // 11:00 - 12:00
				Start: from.Add(-1 * time.Hour),
				End:   from,
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.False(t, isCross)
			},
		},
		{
			name: "OK equal base end time and crossing start time",
			crossingTime: TimeRange{ // 13:00 - 14:00
				Start: to,
				End:   to.Add(1 * time.Hour),
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.False(t, isCross)
			},
		},
		{
			name: "NOT OK crossing end time in base",
			crossingTime: TimeRange{ // 11:00 - 12:30
				Start: from.Add(-1 * time.Hour),
				End:   from.Add(30 * time.Minute),
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.True(t, isCross)
			},
		},
		{
			name: "NOT OK crossing start time in base",
			crossingTime: TimeRange{ // 12:30 - 14:00
				Start: to.Add(-30 * time.Minute),
				End:   to.Add(1 * time.Hour),
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.True(t, isCross)
			},
		},
		{
			name: "NOT OK crossing start time and end time in base",
			crossingTime: TimeRange{ // 12:15 - 12:45
				Start: from.Add(15 * time.Minute),
				End:   to.Add(-15 * time.Minute),
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.True(t, isCross)
			},
		},
		{
			name: "NOT OK crossing equal base",
			crossingTime: TimeRange{ // 12:00 - 13:00
				Start: from,
				End:   to,
			},
			checkResult: func(t *testing.T, isCross bool) {
				assert.True(t, isCross)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.checkResult(t, baseTR.CrossWith(tc.crossingTime))
		})
	}
}

func Test_TimeRange_String(t *testing.T) {
	from := time.Now().Truncate(time.Second).UTC()
	to := from.Add(time.Minute)

	tr, err := NewTimeRange(from, to)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("%s - %s", from.Format(time.DateTime), to.Format(time.DateTime)), fmt.Sprint(tr))
}

func Test_RoomID(t *testing.T) {
	defaultRoomID := "roomID"
	longRoomID := "123456789012345678901234567890123456789012345678901234567890123456789012" // len = 72
	assert.True(t, len(longRoomID) == 72)

	testCases := []struct {
		name        string
		input       string
		checkResult func(t *testing.T, roomID RoomID, err error)
	}{
		{
			name:  "OK",
			input: defaultRoomID,
			checkResult: func(t *testing.T, roomID RoomID, err error) {
				assert.NoError(t, err)
				assert.Equal(t, RoomID(defaultRoomID), roomID)
			},
		},
		{
			name:  "OK long id",
			input: longRoomID,
			checkResult: func(t *testing.T, roomID RoomID, err error) {
				assert.NoError(t, err)
				assert.Equal(t, RoomID(longRoomID), roomID)
			},
		},
		{
			name:  "NOT OK empty input",
			input: "",
			checkResult: func(t *testing.T, roomID RoomID, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, roomID)
			},
		},
		{
			name:  "NOT OK too long id",
			input: longRoomID + "1", // len = 73
			checkResult: func(t *testing.T, roomID RoomID, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, internal.ErrValidationFailed)
				assert.Empty(t, roomID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roomID, err := NewRoomID(tc.input)
			tc.checkResult(t, roomID, err)
		})
	}
}
