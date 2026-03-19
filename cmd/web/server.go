package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	"github.com/mr-cheeezz/dankbot/pkg/web/router"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

const shutdownTimeout = 5 * time.Second

type server struct {
	httpServer *http.Server
	listener   net.Listener
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
		if strings.HasPrefix(strings.TrimSpace(s.httpServer.Addr), "unix:") {
			socketPath := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s.httpServer.Addr), "unix:"))
			if socketPath == "" {
				errCh <- errors.New("web.bind_addr unix socket path is empty")
				return
			}
			_ = os.Remove(socketPath)
			ln, err := net.Listen("unix", socketPath)
			if err != nil {
				errCh <- err
				return
			}
			s.listener = ln
			errCh <- s.httpServer.Serve(ln)
			return
		}

		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := s.httpServer.Shutdown(shutdownCtx)
		if s.listener != nil {
			_ = s.listener.Close()
		}
		return err
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
