package integration

import (
	"context"
	"fmt"
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

	roomID := "1"
	from := time.Now().Truncate(time.Second)
	to := from.Add(1 * time.Hour)

	concurrentReservesCount := 10

	var wg sync.WaitGroup
	wg.Add(concurrentReservesCount)

	var (
		success int32
		fail    int32
	)

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

	wg.Wait()

	assert.Equal(t, int32(1), success)
	assert.Equal(t, int32(9), fail)

	reservations, err := service.ListByRoom(context.Background(), roomID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(reservations))
}
