package integration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynuraddi/test-kami/internal/application"
	repository "github.com/ynuraddi/test-kami/internal/infrastructure/postgres"
	"github.com/ynuraddi/test-kami/pkg/postgres"
	"github.com/ynuraddi/test-kami/test/container"
)

func Test_ReserveRoom_Concurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbContainer, err := container.SetupPostgresContainer(ctx)
	assert.NoError(t, err)

	t.Cleanup(func() {
		dbContainer.Terminate(context.Background())
	})

	dbEndpoint, err := dbContainer.Endpoint(ctx, "")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("postgresql://user:1234@%s/test?sslmode=disable", dbEndpoint)

	psg, err := postgres.NewPool(ctx, dsn)
	assert.NoError(t, err)

	t.Cleanup(func() {
		psg.Close()
	})

	err = postgres.Migrate("file://../../migrations", dsn)
	assert.NoError(t, err)

	repo := repository.NewReservations(psg)
	txM := repository.NewTxManager(psg)

	service := application.NewReservationService(repo, txM)

	now := time.Now().Truncate(time.Second).UTC()

	t.Run("conflict reservations", func(t *testing.T) {
		roomID := "0"
		from := now
		to := from.Add(1 * time.Hour)

		concurrentReservesCount := 30

		var wg sync.WaitGroup
		wg.Add(concurrentReservesCount)

		var (
			success int32
			fail    int32
		)

		done := make(chan struct{})

		for i := 0; i < concurrentReservesCount; i++ {
			go func() {
				if err := service.ReserveRoom(context.Background(), roomID, from, to); err != nil {
					atomic.AddInt32(&fail, 1)
				} else {
					atomic.AddInt32(&success, 1)
				}
				wg.Done()
			}()
		}

		n := time.Now()

		close(done)
		wg.Wait()

		log.Println("serialization:", time.Since(n))

		assert.Equal(t, int32(1), success)
		assert.Equal(t, int32(29), fail)

		reservations, err := service.ListByRoom(context.Background(), roomID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(reservations))
	})

	t.Run("different room with one time reservations", func(t *testing.T) {
		rooms := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

		now := time.Now().Truncate(time.Minute).UTC()

		type reserve struct {
			roomId   string
			from, to time.Time
		}

		done := make(chan struct{})

		var (
			success int32
			fail    int32
		)

		// будет 60 резерваций на команту по 1 минуте
		// на каждую комнату каждой минуту будет претендовать 10 человек
		var wg sync.WaitGroup
		wg.Add(10 * 60 * len(rooms))

		for i := 0; i < 10; i++ {
			for t := 0; t < 60; t++ {
				for _, rid := range rooms {
					go func(roomId string) {
						<-done

						from := now.Add(time.Duration(t) * time.Minute)
						to := from.Add(1 * time.Minute)

						if err := service.ReserveRoom(context.Background(), roomId, from, to); err != nil {
							atomic.AddInt32(&fail, 1)
						} else {
							atomic.AddInt32(&success, 1)
						}
						wg.Done()
					}(rid)
				}
			}
		}

		n := time.Now()

		close(done)
		wg.Wait()

		log.Println("parallel:", time.Since(n))

		assert.Equal(t, int32(len(rooms)*60), success)
		assert.Equal(t, int32(len(rooms)*60*9), fail)
	})
}
