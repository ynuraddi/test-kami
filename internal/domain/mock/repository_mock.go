// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/domain/repository.go

// Package mock_domain is a generated GoMock package.
package mock_domain

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/ynuraddi/test-kami/internal/domain"
)

// MockReservationRepository is a mock of ReservationRepository interface.
type MockReservationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockReservationRepositoryMockRecorder
}

// MockReservationRepositoryMockRecorder is the mock recorder for MockReservationRepository.
type MockReservationRepositoryMockRecorder struct {
	mock *MockReservationRepository
}

// NewMockReservationRepository creates a new mock instance.
func NewMockReservationRepository(ctrl *gomock.Controller) *MockReservationRepository {
	mock := &MockReservationRepository{ctrl: ctrl}
	mock.recorder = &MockReservationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReservationRepository) EXPECT() *MockReservationRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockReservationRepository) Create(ctx context.Context, roomID domain.RoomID, timeRange domain.TimeRange) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, roomID, timeRange)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockReservationRepositoryMockRecorder) Create(ctx, roomID, timeRange interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockReservationRepository)(nil).Create), ctx, roomID, timeRange)
}

// ListByRoom mocks base method.
func (m *MockReservationRepository) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Reservation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByRoom", ctx, roomID)
	ret0, _ := ret[0].([]domain.Reservation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByRoom indicates an expected call of ListByRoom.
func (mr *MockReservationRepositoryMockRecorder) ListByRoom(ctx, roomID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByRoom", reflect.TypeOf((*MockReservationRepository)(nil).ListByRoom), ctx, roomID)
}
