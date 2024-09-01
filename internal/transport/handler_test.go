package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal"
	"github.com/ynuraddi/test-kami/internal/domain"
	mock_transport "github.com/ynuraddi/test-kami/internal/transport/mock"
)

func Test_CreateReservation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mock_transport.NewMockReservationService(ctrl)
	router := NewRouter(service)

	from := time.Now().Truncate(time.Second).UTC()
	to := from.Add(1 * time.Minute)

	unexpectedError := errors.New("unexpecte error")

	defaultInput := createReservationRequest{
		RoomID:    "1",
		StartTime: ReservationTime{from},
		EndTime:   ReservationTime{to},
	}

	testCases := []struct {
		name  string
		input *createReservationRequest

		buildStubs  func()
		checkResult func(t *testing.T, r *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			input: &defaultInput,
			buildStubs: func() {
				service.EXPECT().ReserveRoom(
					gomock.Any(),
					gomock.Eq(defaultInput.RoomID),
					gomock.Eq(defaultInput.StartTime.Time),
					gomock.Eq(defaultInput.EndTime.Time),
				).Times(1).Return(nil)
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, r.Code)
			},
		},
		{
			name:  "NOT OK nil body",
			input: nil, // note
			buildStubs: func() {
				service.EXPECT().ReserveRoom(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, r.Code)
			},
		},
		{
			name:  "NOT OK error from ReserveRoom validation failed",
			input: &defaultInput,
			buildStubs: func() {
				service.EXPECT().ReserveRoom(
					gomock.Any(),
					gomock.Eq(defaultInput.RoomID),
					gomock.Eq(defaultInput.StartTime.Time),
					gomock.Eq(defaultInput.EndTime.Time),
				).Times(1).Return(internal.ErrValidationFailed) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, r.Code)
			},
		},
		{
			name:  "NOT OK error from ReserveRoom reservation conflict",
			input: &defaultInput,
			buildStubs: func() {
				service.EXPECT().ReserveRoom(
					gomock.Any(),
					gomock.Eq(defaultInput.RoomID),
					gomock.Eq(defaultInput.StartTime.Time),
					gomock.Eq(defaultInput.EndTime.Time),
				).Times(1).Return(&domain.ReservationConflictError{}) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusConflict, r.Code)
			},
		},
		{
			name:  "NOT OK error from ReserveRoom unexpected",
			input: &defaultInput,
			buildStubs: func() {
				service.EXPECT().ReserveRoom(
					gomock.Any(),
					gomock.Eq(defaultInput.RoomID),
					gomock.Eq(defaultInput.StartTime.Time),
					gomock.Eq(defaultInput.EndTime.Time),
				).Times(1).Return(unexpectedError) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, r.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()

			body := &bytes.Buffer{}
			if tc.input != nil {
				b, err := json.Marshal(tc.input)
				assert.NoError(t, err)
				body = bytes.NewBuffer(b)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/v1/reservations", body)
			r.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, r)
			tc.checkResult(t, w)
		})
	}
}

func Test_ListByRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mock_transport.NewMockReservationService(ctrl)
	router := NewRouter(service)

	from := time.Now().Truncate(time.Second).UTC()
	to := from.Add(1 * time.Minute)

	defaultRoomID := "1"

	var defaultReservations []domain.Reservation
	defaultReservations = append(defaultReservations,
		domain.Reservation{
			ID:        1,
			RoomUUID:  domain.RoomID(defaultRoomID),
			TimeRange: domain.TimeRange{Start: from, End: to},
		},
		domain.Reservation{
			ID:        2,
			RoomUUID:  domain.RoomID(defaultRoomID),
			TimeRange: domain.TimeRange{Start: to, End: to.Add(time.Minute)},
		})

	unexpectedError := errors.New("unexpecte error")

	testCases := []struct {
		name        string
		roomIDParam string

		buildStubs  func()
		checkResult func(t *testing.T, r *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			roomIDParam: defaultRoomID,
			buildStubs: func() {
				service.EXPECT().ListByRoom(gomock.Any(), gomock.Eq(defaultRoomID)).Times(1).Return(defaultReservations, nil)
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, r.Code)

				var reservations []reservation
				err := json.NewDecoder(r.Body).Decode(&reservations)
				assert.NoError(t, err)

				for i := range reservations {
					assert.Equal(t, defaultReservations[i].ID, reservations[i].ID)
					assert.Equal(t, string(defaultReservations[i].RoomUUID), reservations[i].RoomID)
					assert.Equal(t, defaultReservations[i].TimeRange.Start, reservations[i].StartTime.Time)
					assert.Equal(t, defaultReservations[i].TimeRange.End, reservations[i].EndTime.Time)
				}
			},
		},
		{
			name:        "OK no data",
			roomIDParam: defaultRoomID,
			buildStubs: func() {
				service.EXPECT().ListByRoom(gomock.Any(), gomock.Eq(defaultRoomID)).Times(1).Return(nil, nil) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, r.Code)

				var reservations []reservation
				err := json.NewDecoder(r.Body).Decode(&reservations)
				assert.NoError(t, err)
				assert.Empty(t, reservations)
			},
		},
		{
			name:        "NOT OK error from ListByRoom validation failed",
			roomIDParam: defaultRoomID,
			buildStubs: func() {
				service.EXPECT().ListByRoom(gomock.Any(), gomock.Eq(defaultRoomID)).Times(1).Return(nil, internal.ErrValidationFailed) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, r.Code)
			},
		},
		{
			name:        "NOT OK error from ListByRoom unexpected",
			roomIDParam: defaultRoomID,
			buildStubs: func() {
				service.EXPECT().ListByRoom(gomock.Any(), gomock.Eq(defaultRoomID)).Times(1).Return(nil, unexpectedError) // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, r.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/reservations/%s", tc.roomIDParam), nil)

			router.ServeHTTP(w, r)
			tc.checkResult(t, w)
		})
	}
}

type responseWriterMock struct {
	header http.Header
	t      *testing.T
}

func (m responseWriterMock) Header() http.Header {
	return m.header
}

func (m responseWriterMock) Write([]byte) (int, error) {
	return 0, errors.New("write error") // note
}

func (m responseWriterMock) WriteHeader(statusCode int) {
	assert.Equal(m.t, http.StatusInternalServerError, statusCode)
}

type message struct {
	Msg string `json:"message"`
}

func Test_Write(t *testing.T) {
	type args struct {
		status int
		msg    interface{}
	}

	defaultMsg := message{"123"}

	testCases := []struct {
		name        string
		rw          http.ResponseWriter
		args        args
		checkResult func(t *testing.T, r *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			rw:   httptest.NewRecorder(),
			args: args{
				status: http.StatusOK,
				msg:    defaultMsg,
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, r.Code)

				var msg message
				err := json.NewDecoder(r.Body).Decode(&msg)
				assert.NoError(t, err)
				assert.Equal(t, defaultMsg, msg)
			},
		},
		{
			name: "OK nil msg",
			rw:   httptest.NewRecorder(),
			args: args{
				status: http.StatusOK,
				msg:    nil, // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Empty(t, r.Body)
			},
		},
		{
			name: "OK valid type",
			rw:   httptest.NewRecorder(),
			args: args{
				status: http.StatusOK,
				msg:    map[string]string{"message": defaultMsg.Msg}, // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, r.Code)

				var msg message
				err := json.NewDecoder(r.Body).Decode(&msg)
				assert.NoError(t, err)
				assert.Equal(t, defaultMsg, msg)
			},
		},
		{
			name: "NOT OK invalid type",
			rw:   httptest.NewRecorder(),
			args: args{
				status: http.StatusOK,
				msg:    make(chan int), // note
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, r.Code)
				assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", r.Body.String())
			},
		},
		{
			name: "NOT OK writer error",
			rw: &responseWriterMock{
				header: http.Header{},
				t:      t,
			},
			args: args{
				status: http.StatusOK,
				msg:    defaultMsg,
			},
			checkResult: func(t *testing.T, r *httptest.ResponseRecorder) {}, // check in struct
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			write(tc.rw, tc.args.status, tc.args.msg)

			// need only for case when we useing mock writer
			w, ok := tc.rw.(*httptest.ResponseRecorder)
			if ok {
				tc.checkResult(t, w)
			}
		})
	}
}
