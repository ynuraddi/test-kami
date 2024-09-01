package domain

import (
	"fmt"
	"time"

	"github.com/ynuraddi/test-kami/internal"
)

type Reservation struct {
	ID        int64
	RoomUUID  RoomID
	TimeRange TimeRange
}

func NewReservation(id int64, roomUUID string, from, to time.Time) (Reservation, error) {
	if id <= 0 {
		return Reservation{},
			fmt.Errorf("NewReservation: ID should be positive number: %w", internal.ErrValidationFailed)
	}

	rID, err := NewRoomID(roomUUID)
	if err != nil {
		return Reservation{}, err
	}

	tr, err := NewTimeRange(from, to)
	if err != nil {
		return Reservation{},
			fmt.Errorf("NewReservation: %w", internal.ErrValidationFailed)
	}

	return Reservation{
		ID:        id,
		RoomUUID:  rID,
		TimeRange: tr,
	}, nil
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (t TimeRange) String() string {
	return fmt.Sprintf("%s - %s",
		t.Start.Format(time.DateTime),
		t.End.Format(time.DateTime))
}

func (t TimeRange) CrossWith(other TimeRange) bool {
	return t.Start.Before(other.End) && other.Start.Before(t.End)
}

func NewTimeRange(from, to time.Time) (TimeRange, error) {
	if from.After(to) || from.Equal(to) {
		return TimeRange{},
			fmt.Errorf("NewTimeRange: %w: from_time is after to_time", internal.ErrValidationFailed)
	}

	return TimeRange{
		Start: from.Truncate(time.Second).UTC(),
		End:   to.Truncate(time.Second).UTC(),
	}, nil
}

type RoomID string

func NewRoomID(roomID string) (RoomID, error) {
	if len(roomID) == 0 {
		return "", fmt.Errorf("NewRoomID: %w: empty string", internal.ErrValidationFailed)
	}
	if len(roomID) > 72 {
		return "", fmt.Errorf("NewRoomID: %w: len of RoomID should be less than 72", internal.ErrValidationFailed)
	}

	return RoomID(roomID), nil
}
