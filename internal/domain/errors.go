package domain

import (
	"fmt"
)

type ReservationConflictError struct {
	Reservation         TimeRange
	ConflictReservation TimeRange
}

var _ error = (*ReservationConflictError)(nil)

func (e ReservationConflictError) Error() string {
	return fmt.Sprintf("reservation with time [%s] conflicts with [%s]",
		e.Reservation.String(),
		e.ConflictReservation.String())
}

func (e ReservationConflictError) Is(target error) bool {
	if _, ok := target.(*ReservationConflictError); !ok {
		return false
	}
	return true
}
