package main

import (
	"context"
	"fmt"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type application struct {
	config *config.Config
	state  *state.State
	server *server
}

func newApplication(configPath string) (*application, error) {
	cfg, err := loadWebConfig(configPath)
	if err != nil {
		return nil, err
	}

	postgresClient := postgres.NewClient(cfg.Main.DB)
	redisClient := redispkg.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.KeyPrefix)

	appState := state.New(cfg, postgresClient, redisClient)

	return &application{
		config: cfg,
		state:  appState,
		server: newServer(cfg, appState),
	}, nil
}

func (a *application) Run(ctx context.Context) error {
	fmt.Printf("starting web server on %s\n", a.config.Web.BindAddr)
	a.state.EventSub.Start(ctx)

	err := a.server.Run(ctx)
	closeErr := a.close()
	if err != nil {
		return err
	}
	return closeErr
}

func (a *application) close() error {
	if err := a.state.Redis.Close(); err != nil {
		return err
	}

	return a.state.Postgres.Close()
}
