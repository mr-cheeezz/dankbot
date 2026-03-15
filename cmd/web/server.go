package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	"github.com/mr-cheeezz/dankbot/pkg/web/router"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

const shutdownTimeout = 5 * time.Second

type server struct {
	httpServer *http.Server
}

func newServer(cfg *config.Config, appState *state.State) *server {
	return &server{
		httpServer: &http.Server{
			Addr:    cfg.Web.BindAddr,
			Handler: router.New(appState),
		},
	}
}

func (s *server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
