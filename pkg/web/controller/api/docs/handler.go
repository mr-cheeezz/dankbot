package docs

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

var docsPage = template.Must(template.New("api-docs").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DankBot API Docs</title>
  <style>
    :root {
      color-scheme: dark;
      --bg: #2b2b2b;
      --paper: #1f1f1f;
      --border: #3b3b3b;
      --text: #eeeeee;
      --muted: #a8a8a8;
      --primary: #4a89ff;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      background: var(--bg);
      color: var(--text);
      font-family: "Nunito Sans", "Segoe UI", sans-serif;
    }
    .shell {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    .appbar {
      min-height: 72px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 14px 28px;
      background: #1f1f1f;
      border-bottom: 1px solid #2f2f2f;
    }
    .brand {
      display: inline-flex;
      align-items: center;
      gap: 12px;
      user-select: none;
      min-width: 0;
    }
    .brand-mark {
      width: 34px;
      height: 34px;
      image-rendering: pixelated;
    }
    .brand-copy {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }
    .brand-title {
      font-size: 1.05rem;
      font-weight: 800;
      line-height: 1;
      color: var(--text);
    }
    .brand-subtitle {
      font-size: 0.68rem;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
    }
    .top-links {
      display: inline-flex;
      gap: 10px;
      align-items: center;
      flex-wrap: wrap;
    }
    .top-link {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-height: 38px;
      padding: 0 14px;
      border-radius: 12px;
      text-decoration: none;
      font-size: 0.95rem;
      font-weight: 700;
      color: #cdd4e7;
      border: 1px solid rgba(255,255,255,0.08);
      background: rgba(255,255,255,0.03);
    }
    .top-link:hover {
      background: rgba(255,255,255,0.05);
      border-color: rgba(255,255,255,0.16);
    }
    .hero {
      padding: 28px;
      border-bottom: 1px solid var(--border);
      background: var(--paper);
    }
    .hero-card {
      max-width: 1100px;
      margin: 0 auto;
      padding: 22px 24px;
      border: 1px solid var(--border);
      border-radius: 16px;
      background: #262626;
    }
    .hero-eyebrow {
      margin: 0 0 8px;
      color: var(--primary);
      font-size: 0.75rem;
      font-weight: 800;
      letter-spacing: 0.12em;
      text-transform: uppercase;
    }
    .hero h1 {
      margin: 0;
      font-size: clamp(28px, 5vw, 40px);
      line-height: 1.05;
    }
    .hero p {
      margin: 12px 0 0;
      max-width: 760px;
      color: var(--muted);
      line-height: 1.7;
      font-size: 0.98rem;
    }
    .docs-shell {
      flex: 1;
      padding: 24px 28px 28px;
    }
    #api-reference {
      max-width: 1100px;
      margin: 0 auto;
      border: 1px solid var(--border);
      border-radius: 14px;
      overflow: hidden;
      background: #181818;
    }
  </style>
</head>
<body>
  <div class="shell">
    <header class="appbar">
      <div class="brand">
        <img class="brand-mark" src="/brand/dankbot-mark.svg" alt="DankBot">
        <div class="brand-copy">
          <div class="brand-title">DANKBOT</div>
          <div class="brand-subtitle">API Docs</div>
        </div>
      </div>
      <div class="top-links">
        <a class="top-link" href="/">Home</a>
        <a class="top-link" href="/d">Dashboard</a>
      </div>
    </header>

    <section class="hero">
      <div class="hero-card">
        <p class="hero-eyebrow">Scalar + OpenAPI</p>
        <h1>DankBot API Reference</h1>
        <p>
          Core public and dashboard endpoints, plus the main OAuth entry routes, in one place.
          The spec is still served from <code>/api/openapi.json</code>, now rendered in Scalar.
        </p>
      </div>
    </section>

    <main class="docs-shell">
      <script
        id="api-reference"
        data-url="/api/openapi.json"
        data-theme="purple"
        data-dark-mode="true"
        data-layout="modern">
      </script>
    </main>
  </div>

  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`))

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}
	mux.Handle("/api/openapi.json", http.HandlerFunc(h.openapi))
	mux.Handle("/d/docs", http.HandlerFunc(h.docs))
	mux.Handle("/dashboard/docs", http.RedirectHandler("/d/docs", http.StatusMovedPermanently))
	mux.Handle("/docs", http.RedirectHandler("/d/docs", http.StatusMovedPermanently))
}

func (h handler) docs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = docsPage.Execute(w, nil)
}

func (h handler) openapi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(buildOpenAPISpec(h.appState))
}

func buildOpenAPISpec(appState *state.State) map[string]any {
	serverURL := "/"
	if appState != nil && appState.Config != nil {
		if value := strings.TrimSpace(appState.Config.Web.PublicURL); value != "" {
			serverURL = strings.TrimRight(value, "/")
		}
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "DankBot API",
			"version":     "0.9.1-beta3",
			"description": "Public site, dashboard, and OAuth entry endpoints for DankBot.",
		},
		"servers": []map[string]any{
			{"url": serverURL},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"sessionCookie": map[string]any{
					"type": "apiKey",
					"in":   "cookie",
					"name": session.CookieName,
				},
			},
		},
		"tags": []map[string]any{
			{"name": "Health"},
			{"name": "Auth"},
			{"name": "Public"},
			{"name": "Dashboard"},
		},
		"paths": map[string]any{
			"/healthz": map[string]any{
				"get": op("Health", "Health check", "Simple service health probe."),
			},
			"/api/auth/session": map[string]any{
				"get": op("Auth", "Current web session", "Returns the current signed-in Twitch user and dashboard flags."),
			},
			"/auth/login": map[string]any{
				"get": op("Auth", "Start Twitch site login", "Starts the Twitch OAuth flow for normal website sign-in."),
			},
			"/auth/streamer": map[string]any{
				"get": op("Auth", "Link Twitch streamer account", "Starts the Twitch OAuth flow for broadcaster-side Twitch features."),
			},
			"/auth/bot": map[string]any{
				"get": op("Auth", "Link Twitch bot account", "Starts the Twitch OAuth flow for the bot chat/moderation account."),
			},
			"/auth/spotify": map[string]any{
				"get": op("Auth", "Link Spotify", "Starts Spotify OAuth for now-playing and playback features."),
			},
			"/auth/roblox": map[string]any{
				"get": op("Auth", "Link Roblox", "Starts Roblox OAuth for game/profile features."),
			},
			"/auth/streamlabs": map[string]any{
				"get": op("Auth", "Link Streamlabs", "Starts Streamlabs OAuth for third-party donation alerts."),
			},
			"/auth/streamelements": map[string]any{
				"get": op("Auth", "Link StreamElements", "Starts StreamElements OAuth for third-party tip alerts."),
			},
			"/auth/discord": map[string]any{
				"get": op("Auth", "Install Discord bot", "Starts the Discord guild install flow."),
			},
			"/auth/logout": map[string]any{
				"post": op("Auth", "Logout", "Clears the current DankBot web session."),
			},
			"/api/public/summary": map[string]any{
				"get": op("Public", "Public home summary", "Returns stream, bot, link, and now-playing state for the public home page."),
			},
			"/api/public/commands": map[string]any{
				"get": op("Public", "Public command docs", "Returns viewer and moderator command documentation."),
			},
			"/api/public/quotes": map[string]any{
				"get": op("Public", "Public quotes", "Returns publicly visible quotes."),
			},
			"/api/public/users/{login}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Public"},
					"summary":     "Public user profile",
					"description": "Returns the public profile and user-facing stats for a Twitch login.",
					"parameters": []map[string]any{
						pathParam("login", "Raw Twitch login, for example mr_cheeezz."),
					},
					"responses": okResponses(),
				},
			},
			"/api/dashboard/summary": map[string]any{
				"get": authedOp("Dashboard", "Dashboard summary", "Returns moderator dashboard summary state."),
			},
			"/api/dashboard/bot-controls": map[string]any{
				"get":  authedOp("Dashboard", "Bot controls state", "Returns mode controls for the dashboard overview."),
				"post": authedOp("Dashboard", "Update bot controls", "Updates active bot mode or related dashboard controls."),
			},
			"/api/dashboard/audit-logs": map[string]any{
				"get": authedOp("Dashboard", "Audit logs", "Returns recent dashboard audit activity."),
			},
			"/api/dashboard/killswitch": map[string]any{
				"post": authedOp("Dashboard", "Toggle killswitch", "Updates the bot killswitch state."),
			},
			"/api/dashboard/modes": map[string]any{
				"get":    authedOp("Dashboard", "List modes", "Returns configured bot modes."),
				"post":   authedOp("Dashboard", "Create mode", "Creates a new mode."),
				"put":    authedOp("Dashboard", "Update mode", "Updates a mode."),
				"delete": authedOp("Dashboard", "Delete mode", "Deletes a mode."),
			},
			"/api/dashboard/public-home-settings": map[string]any{
				"get": authedOp("Dashboard", "Get channel settings", "Returns public-home and promo-link settings."),
				"put": authedOp("Dashboard", "Save channel settings", "Saves public-home and promo-link settings."),
			},
			"/api/dashboard/roles": map[string]any{
				"get":    authedOp("Dashboard", "List dashboard roles", "Returns assigned editor roles."),
				"post":   authedOp("Dashboard", "Assign editor", "Assigns manual editor access."),
				"delete": authedOp("Dashboard", "Remove editor", "Removes manual editor access."),
			},
			"/api/dashboard/twitch-user-search": map[string]any{
				"get": authedOp("Dashboard", "Search Twitch users", "Searches Twitch users for editor assignment."),
			},
			"/api/dashboard/moderation/blocked-terms": map[string]any{
				"get":    authedOp("Dashboard", "List blocked terms", "Returns DankBot-owned blocked terms."),
				"post":   authedOp("Dashboard", "Create blocked term", "Creates a new blocked term."),
				"put":    authedOp("Dashboard", "Update blocked term", "Updates a blocked term."),
				"delete": authedOp("Dashboard", "Delete blocked term", "Deletes a blocked term."),
			},
			"/api/dashboard/moderation/mass-action": map[string]any{
				"post": authedOp("Dashboard", "Mass moderation action", "Runs warn, timeout, ban, or unban on many usernames."),
			},
			"/api/dashboard/moderation/recent-followers": map[string]any{
				"get": authedOp("Dashboard", "Recent followers import", "Imports recent followers for mass moderation tools."),
			},
			"/api/dashboard/spam-filters": map[string]any{
				"get": authedOp("Dashboard", "List spam filters", "Returns spam filter settings."),
				"put": authedOp("Dashboard", "Update spam filters", "Saves spam filter settings."),
			},
			"/api/dashboard/discord-bot": map[string]any{
				"get": authedOp("Dashboard", "Discord bot settings", "Returns Discord bot website-managed settings."),
				"put": authedOp("Dashboard", "Save Discord bot settings", "Saves Discord bot website-managed settings."),
			},
			"/api/dashboard/modules": map[string]any{
				"get": authedOp("Dashboard", "Modules catalog", "Returns module metadata and settings schema from the database."),
			},
			"/api/dashboard/modules/game": map[string]any{
				"get": authedOp("Dashboard", "Game module settings", "Returns settings for the game module."),
				"put": authedOp("Dashboard", "Save game module settings", "Saves settings for the game module."),
			},
			"/api/dashboard/modules/now-playing": map[string]any{
				"get": authedOp("Dashboard", "Now playing module settings", "Returns settings for the now playing module."),
				"put": authedOp("Dashboard", "Save now playing module settings", "Saves settings for the now playing module."),
			},
			"/api/dashboard/modules/followers-only": map[string]any{
				"get": authedOp("Dashboard", "Followers-only module settings", "Returns auto followers-only settings."),
				"put": authedOp("Dashboard", "Save followers-only module settings", "Saves auto followers-only settings."),
			},
			"/api/dashboard/modules/new-chatter-greeting": map[string]any{
				"get": authedOp("Dashboard", "New chatter greeting module settings", "Returns first-message greeting settings."),
				"put": authedOp("Dashboard", "Save new chatter greeting module settings", "Saves first-message greeting settings."),
			},
			"/api/dashboard/modules/quotes": map[string]any{
				"get": authedOp("Dashboard", "Quote module settings", "Returns quote module settings."),
				"put": authedOp("Dashboard", "Save quote module settings", "Saves quote module settings."),
			},
			"/api/dashboard/modules/quotes/items": map[string]any{
				"get":    authedOp("Dashboard", "List quotes", "Returns website-managed quotes."),
				"post":   authedOp("Dashboard", "Create quote", "Creates a quote entry."),
				"put":    authedOp("Dashboard", "Update quote", "Updates a quote entry."),
				"delete": authedOp("Dashboard", "Delete quote", "Deletes a quote entry."),
			},
			"/api/dashboard/spotify": map[string]any{
				"get": authedOp("Dashboard", "Spotify dashboard state", "Returns dashboard playback state for the linked Spotify account."),
			},
			"/api/dashboard/spotify/search": map[string]any{
				"get": authedOp("Dashboard", "Spotify search", "Searches Spotify tracks from the dashboard."),
			},
			"/api/dashboard/spotify/queue": map[string]any{
				"post": authedOp("Dashboard", "Spotify queue action", "Queues a track on the linked Spotify account."),
			},
			"/api/dashboard/spotify/playback": map[string]any{
				"post": authedOp("Dashboard", "Spotify playback control", "Play, pause, next, or previous on the linked Spotify account."),
			},
			"/api/dashboard/integrations/unlink": map[string]any{
				"post": authedOp("Dashboard", "Unlink integration", "Unlinks a linked provider account."),
			},
		},
	}
}

func op(tag, summary, description string) map[string]any {
	return map[string]any{
		"tags":        []string{tag},
		"summary":     summary,
		"description": description,
		"responses":   okResponses(),
	}
}

func authedOp(tag, summary, description string) map[string]any {
	item := op(tag, summary, description)
	item["security"] = []map[string][]string{
		{"sessionCookie": {}},
	}
	return item
}

func okResponses() map[string]any {
	return map[string]any{
		"200": map[string]any{
			"description": "OK",
		},
		"401": map[string]any{
			"description": "Unauthorized",
		},
		"403": map[string]any{
			"description": "Forbidden",
		},
	}
}

func pathParam(name, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema": map[string]any{
			"type": "string",
		},
	}
}
