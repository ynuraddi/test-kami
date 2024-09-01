package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(service ReservationService) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Mount("/api/v1", v1(service))

	return r
}

func v1(service ReservationService) *chi.Mux {
	r := chi.NewRouter()

	reservation := NewReservationController(service)

	r.Post("/reservations", reservation.CreateReservation)
	r.Get("/reservations/{room_id}", reservation.ListByRoom)

	return r
}
