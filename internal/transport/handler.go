package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ynuraddi/test-kami/internal"
	"github.com/ynuraddi/test-kami/internal/domain"
)

type ReservationService interface {
	ListByRoom(ctx context.Context, roomID string) ([]domain.Reservation, error)
	ReserveRoom(ctx context.Context, roomID string, from time.Time, to time.Time) (err error)
}

type reservationController struct {
	service ReservationService
}

func NewReservationController(service ReservationService) *reservationController {
	return &reservationController{
		service: service,
	}
}

type createReservationRequest struct {
	RoomID    string          `json:"room_id"`
	StartTime ReservationTime `json:"start_time"`
	EndTime   ReservationTime `json:"end_time"`
}

func (h reservationController) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var req createReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.service.ReserveRoom(ctx, req.RoomID, req.StartTime.Time, req.EndTime.Time)
	if errors.Is(err, internal.ErrValidationFailed) {
		writeError(w, http.StatusBadRequest, err)
		return
	} else if errors.Is(err, &domain.ReservationConflictError{}) {
		writeError(w, http.StatusConflict, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	write(w, http.StatusCreated, nil)
}

func (h reservationController) ListByRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "room_id")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	reservations, err := h.service.ListByRoom(ctx, roomID)
	if errors.Is(err, internal.ErrValidationFailed) {
		writeError(w, http.StatusBadRequest, err)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	out := make([]reservation, 0, len(reservations))
	for _, r := range reservations {
		out = append(out, newResevation(r))
	}

	write(w, http.StatusOK, out)
}

func write(w http.ResponseWriter, status int, msg any) {
	if msg == nil {
		w.WriteHeader(status)
		return
	}

	body, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type jsonError struct {
	Err string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	write(w, status, jsonError{Err: err.Error()})
}
