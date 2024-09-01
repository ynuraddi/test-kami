package transport

import (
	"fmt"
	"strings"
	"time"

	"github.com/ynuraddi/test-kami/internal/domain"
)

// custom time use for correct marshal and unmarshal time datas
type ReservationTime struct {
	time.Time
}

const ReservationTimeLayout = time.DateTime

func (t ReservationTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Format(ReservationTimeLayout))), nil
}

func (t *ReservationTime) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	dataStr = strings.ReplaceAll(dataStr, "\"", "")

	parsedTime, err := time.Parse(ReservationTimeLayout, dataStr)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

type reservation struct {
	ID        int64           `json:"id"`
	RoomID    string          `json:"room_id"`
	StartTime ReservationTime `json:"start_time"`
	EndTime   ReservationTime `json:"end_time"`
}

func newResevation(r domain.Reservation) reservation {
	return reservation{
		ID:        r.ID,
		RoomID:    string(r.RoomID),
		StartTime: ReservationTime{r.TimeRange.Start},
		EndTime:   ReservationTime{r.TimeRange.End},
	}
}
