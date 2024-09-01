package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ynuraddi/test-kami/config"
	"github.com/ynuraddi/test-kami/internal/application"
	repository "github.com/ynuraddi/test-kami/internal/infrastructure/postgres"
	"github.com/ynuraddi/test-kami/internal/transport"
	httpserver "github.com/ynuraddi/test-kami/pkg/httpServer"
	"github.com/ynuraddi/test-kami/pkg/postgres"
)

func init() {
	time.Local = time.UTC
}

func main() {
	configFilePath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.Load(*configFilePath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	psg, err := postgres.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		panic(err)
	}

	if err := postgres.Migrate(cfg.Postgres.MigrationURL, cfg.Postgres.DSN); err != nil {
		panic(err)
	}

	txManager := repository.NewTxManager(psg)
	repo := repository.NewReservations(psg)

	service := application.NewReservationService(repo, txManager)

	handler := transport.NewRouter(service)
	server := httpserver.New(handler, cfg.HTTP.PORT)

	gracefullShutdown(func() {
		if err := server.Shutdown(); err != nil {
			log.Println("shutdown server error:", err.Error())
			return
		}
		psg.Close()
	})

	err = <-server.Notify()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println("server stopped with:", err.Error())
		return
	}
	log.Println("server stopped gracefully")
}

func gracefullShutdown(shutdownFunc func()) {
	osC := make(chan os.Signal, 1)
	signal.Notify(osC, os.Interrupt)

	go func() {
		log.Println(<-osC)
		shutdownFunc()
	}()
}
