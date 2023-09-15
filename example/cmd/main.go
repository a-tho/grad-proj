package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	storage "github.com/a-tho/grad-proj/example/internal/employee"
	"github.com/a-tho/grad-proj/example/transport"
)

func main() {
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	level := zerolog.InfoLevel
	log.Logger = zerolog.New(out).Level(level).With().Timestamp().Logger()

	log.Info().Msg("starting server...")
	defer log.Info().Msg("shutting down server")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT)

	serviceEmployeeStorage := storage.New()

	server := transport.NewServer(log.Logger.With().Str("module", "server").Logger())
	server.InitEmployeeStorageServer(serviceEmployeeStorage)

	go func() {
		log.Info().Str("bind", ":9000").Msg("listen on")
		if err := http.ListenAndServe(":9000", server); err != nil {
			log.Panic().Err(err).Stack().Msg("server error")
		}
	}()

	<-shutdown
}
