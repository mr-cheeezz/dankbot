package router

import (
	"net/http"

	authorized "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/authorized"
	discordauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/discord"
	robloxauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/roblox"
	sessionauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/session"
	spotifyauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/spotify"
	streamelementsauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/streamelements"
	streamlabsauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/streamlabs"
	twitchauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/twitch"
	"github.com/mr-cheeezz/dankbot/pkg/web/controller/api/dashboard"
	"github.com/mr-cheeezz/dankbot/pkg/web/controller/api/health"
	publicsummary "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/public"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
	"github.com/mr-cheeezz/dankbot/pkg/web/static"
)

func New(appState *state.State) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", health.NewHandler(appState))
	authorized.Register(mux, appState)
	discordauth.Register(mux, appState)
	sessionauth.Register(mux, appState)
	robloxauth.Register(mux, appState)
	spotifyauth.Register(mux, appState)
	streamlabsauth.Register(mux, appState)
	streamelementsauth.Register(mux, appState)
	twitchauth.Register(mux, appState)
	dashboard.Register(mux, appState)
	publicsummary.Register(mux, appState)
	mux.Handle("/", static.New("web/dist"))
	return mux
}
