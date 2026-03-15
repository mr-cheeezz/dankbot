package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	twitchhelix "github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

type application struct {
	config  *config.Config
	runtime *runtime
}

func newApplication(configPath string) (*application, error) {
	cfg, err := loadBotConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &application{
		config:  cfg,
		runtime: newRuntime(cfg),
	}, nil
}

func (a *application) Run(ctx context.Context) error {
	fmt.Printf("starting bot runtime for streamer %s on instance %s\n", a.startupStreamerLabel(ctx), a.config.Worker.InstanceID)

	return a.runtime.Run(ctx)
}

func (a *application) startupStreamerLabel(ctx context.Context) string {
	streamerID := strings.TrimSpace(a.config.Main.StreamerID)
	if streamerID == "" {
		return "[streamer id or twitch client id is empty]"
	}

	if label := a.streamerLabelFromDB(ctx); label != "" {
		return label
	}

	if label := a.streamerLabelFromHelix(ctx, streamerID); label != "" {
		return label
	}

	return streamerID
}

func (a *application) streamerLabelFromDB(ctx context.Context) string {
	client := postgres.NewClient(a.config.Main.DB)
	defer client.Close()

	store := postgres.NewTwitchAccountStore(client)
	account, err := store.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return ""
	}

	if strings.TrimSpace(account.DisplayName) != "" {
		return account.DisplayName
	}

	return strings.TrimSpace(account.Login)
}

func (a *application) streamerLabelFromHelix(ctx context.Context, streamerID string) string {
	oauthClient := twitchoauth.NewClient(nil, a.config.Twitch.ClientID, a.config.Twitch.ClientSecret, strings.TrimSpace(a.config.Twitch.RedirectURI))
	oauthService := twitchoauth.NewService(oauthClient, nil)

	token, err := oauthService.AppToken(ctx)
	if err != nil || token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return ""
	}

	helixClient := twitchhelix.NewClient(a.config.Twitch.ClientID, token.AccessToken)
	users, err := helixClient.GetUsersByIDs(ctx, []string{streamerID})
	if err != nil || len(users) == 0 {
		return ""
	}

	if strings.TrimSpace(users[0].DisplayName) != "" {
		return users[0].DisplayName
	}

	return strings.TrimSpace(users[0].Login)
}
