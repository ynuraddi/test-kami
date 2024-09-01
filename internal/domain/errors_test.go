package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ReservationConflictError(t *testing.T) {
	now := time.Now()

	tr1, err := NewTimeRange(now, now.Add(1*time.Minute))
	assert.NoError(t, err)

	tr2, err := NewTimeRange(now, now.Add(30*time.Second))
	assert.NoError(t, err)

	conflict := ReservationConflictError{
		Reservation:         tr1,
		ConflictReservation: tr2,
	}

	assert.Equal(t, fmt.Sprintf(
		"reservation with time [%s - %s] conflicts with [%s - %s]",
		tr1.Start.Format(time.DateTime),
		tr1.End.Format(time.DateTime),
		tr2.Start.Format(time.DateTime),
		tr2.End.Format(time.DateTime),
	), conflict.Error())

	targetErr := &ReservationConflictError{}
	assert.True(t, conflict.Is(targetErr))
	assert.False(t, conflict.Is(fmt.Errorf("some error")))

	wrappedErr := fmt.Errorf("some error: %w", conflict)
	assert.ErrorIs(t, wrappedErr, targetErr)
}
