package discordauth

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	discordoauth "github.com/mr-cheeezz/dankbot/pkg/discord/oauth"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

var joinedPage = template.Must(template.New("joined").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DankBot - Discord Bot Joined</title>
  <style>
    :root { color-scheme: dark; }
    * { box-sizing: border-box; }
    a { color: inherit; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Segoe UI", system-ui, sans-serif;
      background:
        linear-gradient(180deg, rgba(255,255,255,0.03), rgba(255,255,255,0)),
        #232323;
      color: #f4f7ff;
    }
    .shell {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    .appbar {
      position: sticky;
      top: 0;
      z-index: 10;
      min-height: 72px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 18px;
      padding: 14px 28px;
      background: rgba(35, 35, 35, 0.96);
      border-bottom: 1px solid rgba(255,255,255,0.08);
      backdrop-filter: blur(12px);
    }
    .brand {
      display: inline-flex;
      align-items: center;
      gap: 12px;
      user-select: none;
    }
    .brand-mark {
      width: 36px;
      height: 36px;
      image-rendering: pixelated;
      display: block;
    }
    .brand-copy {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }
    .brand-title {
      font-size: 1.1rem;
      font-weight: 800;
      line-height: 1;
      color: #f4f7ff;
    }
    .brand-subtitle {
      font-size: 0.68rem;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: #a8afc4;
    }
    .top-links {
      display: inline-flex;
      align-items: center;
      gap: 10px;
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
      font-weight: 700;
      font-size: 0.95rem;
      color: #cdd4e7;
      border: 1px solid rgba(255,255,255,0.08);
      background: rgba(255,255,255,0.03);
    }
    .content {
      flex: 1;
      display: grid;
      place-items: center;
      padding: 32px 24px 40px;
    }
    .card {
      width: min(640px, calc(100vw - 48px));
      padding: 32px;
      border-radius: 22px;
      background: rgba(24, 24, 24, 0.96);
      border: 1px solid rgba(255,255,255,0.08);
      box-shadow: 0 20px 55px rgba(0,0,0,0.28);
    }
    h1 { margin: 0 0 12px; font-size: clamp(28px, 5vw, 40px); line-height: 1.05; }
    p { margin: 0; color: #c8d0ea; line-height: 1.6; }
    .ok {
      margin: 0 0 10px;
      color: #8ff0b7;
      font-weight: 700;
      font-size: 13px;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }
    .meta {
      margin-top: 18px;
      padding: 14px 16px;
      border-radius: 14px;
      background: rgba(255,255,255,0.04);
      border: 1px solid rgba(255,255,255,0.08);
      font-size: 14px;
      color: #9cabcf;
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      margin-top: 24px;
    }
    .button {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-height: 44px;
      padding: 0 18px;
      border-radius: 12px;
      text-decoration: none;
      font-weight: 700;
    }
    .button-primary {
      background: #4a89ff;
      color: #ffffff;
    }
    .button-secondary {
      border: 1px solid rgba(255,255,255,0.12);
      color: #e7ecfb;
      background: rgba(255,255,255,0.03);
    }
    code {
      padding: 2px 6px;
      border-radius: 8px;
      background: rgba(255,255,255,0.08);
      color: #fff;
    }
  </style>
</head>
<body>
  <div class="shell">
    <header class="appbar">
      <div class="brand">
        <img class="brand-mark" src="/brand/dankbot-mark.svg" alt="dankbot">
        <div class="brand-copy">
          <div class="brand-title">DANKBOT</div>
          <div class="brand-subtitle">Dashboard Access</div>
        </div>
      </div>
      <nav class="top-links" aria-label="Discord join navigation">
        <a class="top-link" href="/d">Dashboard</a>
        <a class="top-link" href="/d/discord">Discord Bot</a>
        <a class="top-link" href="/">Home</a>
      </nav>
    </header>

    <main class="content">
      <section class="card">
        <p class="ok">Discord OAuth complete</p>
        <h1>Discord bot joined successfully.</h1>
        {{if .UserName}}<p>{{.UserName}}, the bot install completed and the server is ready for Discord-side commands once the bot runtime is online.</p>{{else}}<p>The bot install completed and the server is ready for Discord-side commands once the bot runtime is online.</p>{{end}}
        {{if .GuildID}}<p class="meta">Guild ID: <code>{{.GuildID}}</code></p>{{end}}
        <div class="actions">
          <a class="button button-primary" href="/d/discord">Open Discord Bot</a>
          <a class="button button-secondary" href="/d/integrations">Back to integrations</a>
        </div>
      </section>
    </main>
  </div>
</body>
</html>`))

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}
	mux.Handle("/auth/discord", http.HandlerFunc(h.join))
	mux.Handle("/joined", http.HandlerFunc(h.joined))
}

func (h handler) enabled() bool {
	return h.appState != nil && h.appState.Config != nil && h.appState.Config.Discord.Enabled && h.appState.DiscordOAuth != nil
}

func (h handler) join(w http.ResponseWriter, r *http.Request) {
	if !h.enabled() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if !h.requireIntegrationAccess(w, r) {
		return
	}

	url, err := h.appState.DiscordOAuth.JoinURL(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h handler) joined(w http.ResponseWriter, r *http.Request) {
	if !h.enabled() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	query := r.URL.Query()
	if authError := strings.TrimSpace(query.Get("error")); authError != "" {
		description := strings.TrimSpace(query.Get("error_description"))
		if description != "" {
			redirectJoinedResult(w, r, "error", "Discord connection failed", authError+": "+description)
			return
		}
		redirectJoinedResult(w, r, "error", "Discord connection failed", authError)
		return
	}

	stateKey := strings.TrimSpace(query.Get("state"))
	code := strings.TrimSpace(query.Get("code"))
	if stateKey == "" || code == "" {
		redirectJoinedResult(w, r, "error", "Discord connection failed", "missing discord oauth state or code")
		return
	}

	result, err := h.appState.DiscordOAuth.HandleCallback(
		r.Context(),
		stateKey,
		code,
		query.Get("guild_id"),
		query.Get("permissions"),
	)
	if err != nil {
		if err == discordoauth.ErrStateNotFound {
			redirectJoinedResult(w, r, "error", "Discord connection failed", "this discord join link has expired or was already used")
			return
		}
		redirectJoinedResult(w, r, "error", "Discord connection failed", err.Error())
		return
	}

	userName := ""
	userID := ""
	if result.User != nil {
		userID = strings.TrimSpace(result.User.ID)
		userName = strings.TrimSpace(result.User.GlobalName)
		if userName == "" {
			userName = strings.TrimSpace(result.User.Username)
		}
	}

	if h.appState != nil && h.appState.DiscordBotInstallation != nil {
		_, saveErr := h.appState.DiscordBotInstallation.Save(r.Context(), postgres.DiscordBotInstallation{
			GuildID:           result.GuildID,
			InstallerUserID:   userID,
			InstallerUsername: userName,
			Permissions:       result.Permissions,
		})
		if saveErr != nil {
			redirectJoinedResult(w, r, "error", "Discord connection failed", saveErr.Error())
			return
		}
	}

	message := "The Discord bot install completed and the server is ready for Discord-side commands."
	if userName != "" {
		message = userName + ", the Discord bot install completed and the server is ready for Discord-side commands."
	}
	redirectJoinedResult(w, r, "success", "Discord bot joined", message)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeErrorPage(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte("<!doctype html><html><body style=\"margin:0;min-height:100vh;font-family:system-ui;background:#232323;color:#f4f7ff;display:grid;place-items:center;padding:24px;\"><main style=\"width:min(640px,calc(100vw - 48px));padding:32px;border-radius:22px;background:rgba(24,24,24,0.96);border:1px solid rgba(255,255,255,0.08);box-shadow:0 20px 55px rgba(0,0,0,0.28)\"><p style=\"margin:0 0 10px;color:#ff9aa2;font-weight:700;font-size:13px;letter-spacing:.08em;text-transform:uppercase\">Discord OAuth error</p><h1 style=\"margin:0 0 12px;font-size:clamp(28px,5vw,40px);line-height:1.05\">Discord join failed</h1><p style=\"margin:0;color:#c8d0ea;line-height:1.6\">" + template.HTMLEscapeString(message) + "</p><div style=\"display:flex;gap:12px;flex-wrap:wrap;margin-top:24px\"><a href=\"/d/integrations\" style=\"display:inline-flex;align-items:center;justify-content:center;min-height:44px;padding:0 18px;border-radius:12px;text-decoration:none;font-weight:700;background:#4a89ff;color:#fff\">Back to integrations</a><a href=\"/d\" style=\"display:inline-flex;align-items:center;justify-content:center;min-height:44px;padding:0 18px;border-radius:12px;text-decoration:none;font-weight:700;border:1px solid rgba(255,255,255,0.12);background:rgba(255,255,255,0.03);color:#e7ecfb\">Open dashboard</a></div></main></body></html>"))
}

func redirectJoinedResult(w http.ResponseWriter, r *http.Request, status, title, message string) {
	params := url.Values{}
	params.Set("oauth_status", strings.TrimSpace(status))
	if trimmed := strings.TrimSpace(title); trimmed != "" {
		params.Set("oauth_title", trimmed)
	}
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		params.Set("oauth_message", trimmed)
	}

	target := "/d/integrations"
	if encoded := params.Encode(); encoded != "" {
		target += "?" + encoded
	}

	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (h handler) requireIntegrationAccess(w http.ResponseWriter, r *http.Request) bool {
	_, access, err := webaccess.LoadDashboardSession(r.Context(), r, h.appState)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return false
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	if !webaccess.CanLinkStreamerIntegrations(access) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}
