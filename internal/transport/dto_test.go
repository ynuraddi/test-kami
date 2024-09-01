package transport

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testReservationTime struct {
	Time ReservationTime `json:"reservation_time"`
}

type testOtherFormat struct {
	Time time.Time `json:"reservation_time"`
}

func Test_ReservationTime(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()

	rt1 := ReservationTime{
		Time: now,
	}

	test1 := testReservationTime{
		Time: rt1,
	}

	data, err := json.Marshal(test1)
	assert.NoError(t, err)

	var test2 testReservationTime
	err = json.Unmarshal(data, &test2)
	assert.NoError(t, err)
	assert.Equal(t, test1, test2)

	data, err = json.Marshal(&testOtherFormat{rt1.Time})
	assert.NoError(t, err)

	var test3 testReservationTime
	err = json.Unmarshal(data, &test3)
	assert.Error(t, err)

}
