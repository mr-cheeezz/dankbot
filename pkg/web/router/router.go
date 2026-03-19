package router

import (
	"net/http"
	"strings"

	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	authorized "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/authorized"
	discordauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/discord"
	robloxauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/roblox"
	sessionauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/session"
	spotifyauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/spotify"
	streamelementsauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/streamelements"
	streamlabsauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/streamlabs"
	twitchauth "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/auth/twitch"
	"github.com/mr-cheeezz/dankbot/pkg/web/controller/api/dashboard"
	apidocs "github.com/mr-cheeezz/dankbot/pkg/web/controller/api/docs"
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
	apidocs.Register(mux, appState)
	publicsummary.Register(mux, appState)

	// Serve SPA index with the correct HTTP status codes without URL redirects.
	mux.Handle("/403", static.NewIndex("web/dist", func(r *http.Request) int { return http.StatusForbidden }))
	mux.Handle("/404", static.NewIndex("web/dist", func(r *http.Request) int { return http.StatusNotFound }))

	dashboardIndex := static.NewIndex("web/dist", func(r *http.Request) int {
		_, access, err := webaccess.LoadDashboardSession(r.Context(), r, appState)
		if err != nil {
			return http.StatusForbidden
		}

		view := ""
		path := strings.TrimPrefix(strings.TrimSpace(r.URL.Path), "/d")
		path = strings.TrimPrefix(path, "/")
		if path != "" {
			if segment, _, ok := strings.Cut(path, "/"); ok {
				view = strings.TrimSpace(segment)
			} else {
				view = strings.TrimSpace(path)
			}
		}

		switch view {
		case "integrations":
			if !webaccess.CanManageIntegrations(access) {
				return http.StatusForbidden
			}
		case "channel-points", "giveaways", "discord", "modes", "blocked-terms", "mass-moderation":
			if !webaccess.CanAccessEditorFeatures(access) {
				return http.StatusForbidden
			}
		}

		return http.StatusOK
	})
	mux.Handle("/d", dashboardIndex)
	mux.Handle("/d/", dashboardIndex)

	mux.Handle("/", static.New("web/dist"))
	return mux
}
