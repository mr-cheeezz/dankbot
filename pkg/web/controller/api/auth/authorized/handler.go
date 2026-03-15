package authorized

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	robloxoauth "github.com/mr-cheeezz/dankbot/pkg/roblox/oauth"
	spotifyapi "github.com/mr-cheeezz/dankbot/pkg/spotify/api"
	spotifyoauth "github.com/mr-cheeezz/dankbot/pkg/spotify/oauth"
	streamelementsapi "github.com/mr-cheeezz/dankbot/pkg/streamelements/api"
	streamelementsoauth "github.com/mr-cheeezz/dankbot/pkg/streamelements/oauth"
	streamlabsapi "github.com/mr-cheeezz/dankbot/pkg/streamlabs/api"
	streamlabsoauth "github.com/mr-cheeezz/dankbot/pkg/streamlabs/oauth"
	twitchEventSub "github.com/mr-cheeezz/dankbot/pkg/twitch/eventsub"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

type callbackResponse struct {
	Provider    string                        `json:"provider"`
	Flow        string                        `json:"flow"`
	Account     any                           `json:"account,omitempty"`
	Validation  *twitchoauth.ValidateResponse `json:"validation,omitempty"`
	UserInfo    any                           `json:"user_info,omitempty"`
	RequestedAt time.Time                     `json:"requested_at"`
}

type twitchAccountResponse struct {
	Kind         string    `json:"kind"`
	TwitchUserID string    `json:"twitch_user_id"`
	Login        string    `json:"login"`
	Scopes       []string  `json:"scopes"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type robloxAccountResponse struct {
	Kind         string    `json:"kind"`
	RobloxUserID string    `json:"roblox_user_id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type spotifyAccountResponse struct {
	Kind          string    `json:"kind"`
	SpotifyUserID string    `json:"spotify_user_id"`
	DisplayName   string    `json:"display_name"`
	Product       string    `json:"product"`
	Country       string    `json:"country"`
	Scope         string    `json:"scope"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type streamlabsAccountResponse struct {
	Kind             string    `json:"kind"`
	StreamlabsUserID string    `json:"streamlabs_user_id"`
	DisplayName      string    `json:"display_name"`
	Scope            string    `json:"scope"`
	ExpiresAt        time.Time `json:"expires_at"`
	SocketToken      string    `json:"socket_token"`
}

type streamElementsAccountResponse struct {
	Kind        string    `json:"kind"`
	ChannelID   string    `json:"channel_id"`
	Provider    string    `json:"provider"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type siteLoginResult struct {
	sessionID          string
	canAccessDashboard bool
}

type authorizedPageData struct {
	Title          string
	Eyebrow        string
	Message        string
	AccountLabel   string
	AccountValue   string
	PrimaryLabel   string
	PrimaryHref    string
	SecondaryLabel string
	SecondaryHref  string
	RedirectHref   string
}

type authPageVariant string

const (
	authPageVariantPublic    authPageVariant = "public"
	authPageVariantDashboard authPageVariant = "dashboard"
)

var publicAuthorizedPage = template.Must(template.New("authorized-public").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DankBot - {{.Title}}</title>
  {{if .RedirectHref}}<meta http-equiv="refresh" content="1.8;url={{.RedirectHref}}">{{end}}
  <style>
    :root { color-scheme: dark; }
    * { box-sizing: border-box; }
    a { color: inherit; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Segoe UI", system-ui, sans-serif;
      background:
        radial-gradient(circle at top left, rgba(120, 87, 255, 0.18) 0%, rgba(120, 87, 255, 0) 30%),
        radial-gradient(circle at bottom right, rgba(74,137,255,0.18) 0%, rgba(74,137,255,0) 26%),
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
      background: rgba(35, 35, 35, 0.92);
      border-bottom: 1px solid rgba(255,255,255,0.08);
      backdrop-filter: blur(12px);
    }
    .brand {
      display: inline-flex;
      align-items: center;
      gap: 12px;
      user-select: none;
      min-width: 0;
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
      min-width: 0;
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
      transition: background-color 120ms ease, border-color 120ms ease, transform 120ms ease;
    }
    .top-link:hover {
      transform: translateY(-1px);
      border-color: rgba(255,255,255,0.16);
      background: rgba(255,255,255,0.05);
    }
    .content {
      flex: 1;
      display: grid;
      place-items: center;
      padding: 32px 24px 40px;
    }
    .card {
      width: min(680px, calc(100vw - 48px));
      padding: 32px;
      border-radius: 22px;
      background: rgba(24, 24, 24, 0.9);
      border: 1px solid rgba(255,255,255,0.08);
      box-shadow: 0 20px 55px rgba(0,0,0,0.28);
    }
    .eyebrow {
      margin: 0 0 10px;
      font-size: 13px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #8ff0b7;
    }
    h1 {
      margin: 0 0 12px;
      font-size: clamp(28px, 5vw, 40px);
      line-height: 1.05;
    }
    p {
      margin: 0;
      color: #c6cad7;
      line-height: 1.6;
      font-size: 16px;
    }
    .meta {
      margin-top: 18px;
      padding: 14px 16px;
      border-radius: 14px;
      background: rgba(255,255,255,0.04);
      border: 1px solid rgba(255,255,255,0.08);
    }
    .meta-label {
      display: block;
      margin-bottom: 4px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #9cabcf;
    }
    .meta-value {
      font-size: 16px;
      font-weight: 700;
      color: #ffffff;
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
      transition: transform 120ms ease, background-color 120ms ease, border-color 120ms ease, color 120ms ease;
    }
    .button:hover {
      transform: translateY(-1px);
    }
    .button-primary {
      background: #4a89ff;
      color: #ffffff;
    }
    .button-primary:hover {
      background: #5a95ff;
    }
    .button-secondary {
      border: 1px solid rgba(255,255,255,0.12);
      color: #e7ecfb;
      background: rgba(255,255,255,0.03);
    }
    .button-secondary:hover {
      border-color: rgba(255,255,255,0.2);
      background: rgba(255,255,255,0.06);
    }
    .redirect {
      margin-top: 18px;
      font-size: 14px;
      color: #9cabcf;
    }
    .footer {
      padding: 0 28px 24px;
      color: #97a0b7;
      font-size: 13px;
    }
    @media (max-width: 760px) {
      .appbar {
        padding: 14px 16px;
        flex-direction: column;
        align-items: flex-start;
      }
      .content {
        padding: 24px 16px 32px;
      }
      .card {
        width: min(680px, calc(100vw - 32px));
        padding: 24px;
      }
      .footer {
        padding: 0 16px 20px;
      }
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
          <div class="brand-subtitle">Connected</div>
        </div>
      </div>
      <nav class="top-links" aria-label="Auth navigation">
        <a class="top-link" href="/">Home</a>
        <a class="top-link" href="/commands">Commands</a>
        <a class="top-link" href="/quotes">Quotes</a>
      </nav>
    </header>

    <main class="content">
      <section class="card">
        <p class="eyebrow">{{.Eyebrow}}</p>
        <h1>{{.Title}}</h1>
        <p>{{.Message}}</p>
        {{if .AccountValue}}
          <div class="meta">
            <span class="meta-label">{{.AccountLabel}}</span>
            <span class="meta-value">{{.AccountValue}}</span>
          </div>
        {{end}}
        <div class="actions">
          {{if .PrimaryHref}}<a class="button button-primary" href="{{.PrimaryHref}}">{{.PrimaryLabel}}</a>{{end}}
          {{if .SecondaryHref}}<a class="button button-secondary" href="{{.SecondaryHref}}">{{.SecondaryLabel}}</a>{{end}}
        </div>
        {{if .RedirectHref}}<p class="redirect">Redirecting now. If nothing happens, use the button above.</p>{{end}}
      </section>
    </main>

    <footer class="footer">
      DankBot account linking keeps your streamer tools, alerts, and integrations connected in one place.
    </footer>
  </div>
</body>
</html>`))

var dashboardAuthorizedPage = template.Must(template.New("authorized-dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DankBot - {{.Title}}</title>
  {{if .RedirectHref}}<meta http-equiv="refresh" content="1.8;url={{.RedirectHref}}">{{end}}
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
      min-width: 0;
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
      min-width: 0;
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
      transition: background-color 120ms ease, border-color 120ms ease, transform 120ms ease;
    }
    .top-link:hover {
      transform: translateY(-1px);
      border-color: rgba(255,255,255,0.16);
      background: rgba(255,255,255,0.05);
    }
    .content {
      flex: 1;
      display: grid;
      place-items: center;
      padding: 32px 24px 40px;
    }
    .card {
      width: min(680px, calc(100vw - 48px));
      padding: 32px;
      border-radius: 22px;
      background: rgba(24, 24, 24, 0.96);
      border: 1px solid rgba(255,255,255,0.08);
      box-shadow: 0 20px 55px rgba(0,0,0,0.28);
    }
    .eyebrow {
      margin: 0 0 10px;
      font-size: 13px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #8ff0b7;
    }
    h1 {
      margin: 0 0 12px;
      font-size: clamp(28px, 5vw, 40px);
      line-height: 1.05;
    }
    p {
      margin: 0;
      color: #c6cad7;
      line-height: 1.6;
      font-size: 16px;
    }
    .meta {
      margin-top: 18px;
      padding: 14px 16px;
      border-radius: 14px;
      background: rgba(255,255,255,0.04);
      border: 1px solid rgba(255,255,255,0.08);
    }
    .meta-label {
      display: block;
      margin-bottom: 4px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #9cabcf;
    }
    .meta-value {
      font-size: 16px;
      font-weight: 700;
      color: #ffffff;
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
      transition: transform 120ms ease, background-color 120ms ease, border-color 120ms ease, color 120ms ease;
    }
    .button:hover {
      transform: translateY(-1px);
    }
    .button-primary {
      background: #4a89ff;
      color: #ffffff;
    }
    .button-primary:hover {
      background: #5a95ff;
    }
    .button-secondary {
      border: 1px solid rgba(255,255,255,0.12);
      color: #e7ecfb;
      background: rgba(255,255,255,0.03);
    }
    .button-secondary:hover {
      border-color: rgba(255,255,255,0.2);
      background: rgba(255,255,255,0.06);
    }
    .redirect {
      margin-top: 18px;
      font-size: 14px;
      color: #9cabcf;
    }
    .footer {
      padding: 0 28px 24px;
      color: #97a0b7;
      font-size: 13px;
    }
    @media (max-width: 760px) {
      .appbar {
        padding: 14px 16px;
        flex-direction: column;
        align-items: flex-start;
      }
      .content {
        padding: 24px 16px 32px;
      }
      .card {
        width: min(680px, calc(100vw - 32px));
        padding: 24px;
      }
      .footer {
        padding: 0 16px 20px;
      }
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
      <nav class="top-links" aria-label="Auth navigation">
        <a class="top-link" href="/dashboard">Dashboard</a>
        <a class="top-link" href="/dashboard/integrations">Integrations</a>
        <a class="top-link" href="/">Home</a>
      </nav>
    </header>

    <main class="content">
      <section class="card">
        <p class="eyebrow">{{.Eyebrow}}</p>
        <h1>{{.Title}}</h1>
        <p>{{.Message}}</p>
        {{if .AccountValue}}
          <div class="meta">
            <span class="meta-label">{{.AccountLabel}}</span>
            <span class="meta-value">{{.AccountValue}}</span>
          </div>
        {{end}}
        <div class="actions">
          {{if .PrimaryHref}}<a class="button button-primary" href="{{.PrimaryHref}}">{{.PrimaryLabel}}</a>{{end}}
          {{if .SecondaryHref}}<a class="button button-secondary" href="{{.SecondaryHref}}">{{.SecondaryLabel}}</a>{{end}}
        </div>
        {{if .RedirectHref}}<p class="redirect">Redirecting now. If nothing happens, use the button above.</p>{{end}}
      </section>
    </main>

    <footer class="footer">
      DankBot keeps sign-in and broadcaster tools separate so the public site and dashboard stay clean.
    </footer>
  </div>
</body>
</html>`))

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}
	mux.Handle("/authorized", http.HandlerFunc(h.authorizedCallback))
	mux.Handle("/connected", http.HandlerFunc(h.connectedCallback))
}

func (h handler) authorizedCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.oauthCallback(w, r, authPageVariantDashboard)
	case http.MethodPost:
		h.eventSubCallback(w, r)
	default:
		writeMethodNotAllowed(w, authPageVariantDashboard, http.MethodGet, http.MethodPost)
	}
}

func (h handler) connectedCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, authPageVariantPublic, http.MethodGet)
		return
	}

	h.oauthCallback(w, r, authPageVariantPublic)
}

func (h handler) oauthCallback(w http.ResponseWriter, r *http.Request, variant authPageVariant) {
	query := r.URL.Query()
	if authError := strings.TrimSpace(query.Get("error")); authError != "" {
		message := authError
		if description := strings.TrimSpace(query.Get("error_description")); description != "" {
			message += ": " + description
		}
		writeOAuthErrorPage(w, message, http.StatusBadRequest, variant)
		return
	}

	stateKey := strings.TrimSpace(query.Get("state"))
	code := strings.TrimSpace(query.Get("code"))
	if stateKey == "" {
		writeOAuthErrorPage(w, "missing oauth state", http.StatusBadRequest, variant)
		return
	}
	if code == "" {
		writeOAuthErrorPage(w, "missing oauth code", http.StatusBadRequest, variant)
		return
	}

	twitchResult, err := h.appState.TwitchOAuth.HandleCallback(r.Context(), stateKey, code)
	switch {
	case err == nil:
		if twitchResult.Flow == twitchoauth.FlowSiteLogin {
			loginResult, err := h.persistSiteLoginSession(r.Context(), twitchResult)
			if err != nil {
				writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, authPageVariantDashboard)
				return
			}

			session.SetCookie(w, loginResult.sessionID, isSecureCookie(h.appState))
			page := authorizedPageData{
				Title:        "Signed in with Twitch",
				Eyebrow:      "Authorization complete",
				Message:      "You are signed in and DankBot is getting your session ready now.",
				AccountLabel: "Twitch account",
				AccountValue: preferredLoginName(twitchResult),
				PrimaryLabel: "Continue to home page",
				PrimaryHref:  "/",
				RedirectHref: "/",
			}
			if loginResult.canAccessDashboard {
				page.SecondaryLabel = "Open dashboard"
				page.SecondaryHref = "/dashboard"
			}
			writeAuthorizedPage(w, http.StatusOK, authPageVariantDashboard, page)
			return
		}

		response, err := h.finishTwitchLinkedCallback(r.Context(), twitchResult)
		if err != nil {
			writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
			return
		}
		writeAuthorizedPage(w, http.StatusOK, variant, buildAuthorizedPageData(response))
		return
	case errors.Is(err, twitchoauth.ErrStateNotFound):
	default:
		writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
		return
	}

	var response *callbackResponse

	response, err = h.handleRobloxCallback(r.Context(), stateKey, code)
	switch {
	case err == nil:
		writeAuthorizedPage(w, http.StatusOK, variant, buildAuthorizedPageData(response))
		return
	case errors.Is(err, robloxoauth.ErrStateNotFound):
	default:
		writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
		return
	}

	response, err = h.handleSpotifyCallback(r.Context(), stateKey, code)
	switch {
	case err == nil:
		writeAuthorizedPage(w, http.StatusOK, variant, buildAuthorizedPageData(response))
		return
	case errors.Is(err, spotifyoauth.ErrStateNotFound):
	default:
		writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
		return
	}

	response, err = h.handleStreamlabsCallback(r.Context(), stateKey, code)
	switch {
	case err == nil:
		writeAuthorizedPage(w, http.StatusOK, variant, buildAuthorizedPageData(response))
		return
	case errors.Is(err, streamlabsoauth.ErrStateNotFound):
	default:
		writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
		return
	}

	response, err = h.handleStreamElementsCallback(r.Context(), stateKey, code)
	switch {
	case err == nil:
		writeAuthorizedPage(w, http.StatusOK, variant, buildAuthorizedPageData(response))
		return
	case errors.Is(err, streamelementsoauth.ErrStateNotFound):
		writeOAuthErrorPage(w, "unknown oauth state", http.StatusBadRequest, variant)
	default:
		writeOAuthErrorPage(w, err.Error(), http.StatusBadRequest, variant)
	}
}

func (h handler) eventSubCallback(w http.ResponseWriter, r *http.Request) {
	if h.appState == nil || h.appState.EventSub == nil {
		http.Error(w, "eventsub is not configured", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	resp, err := h.appState.EventSub.HandleWebhook(r.Context(), twitchEventSub.HeadersFromRequest(r), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if resp.ContentType != "" {
		w.Header().Set("Content-Type", resp.ContentType)
	}
	w.WriteHeader(resp.StatusCode)
	if len(resp.Body) > 0 {
		_, _ = w.Write(resp.Body)
	}
}

func (h handler) finishTwitchLinkedCallback(ctx context.Context, result *twitchoauth.CallbackResult) (*callbackResponse, error) {
	response := &callbackResponse{
		Provider:    "twitch",
		Flow:        string(result.Flow),
		Validation:  result.Validation,
		UserInfo:    result.UserInfo,
		RequestedAt: result.RequestedAt,
	}

	if result.Flow == twitchoauth.FlowStreamerConnect || result.Flow == twitchoauth.FlowBotConnect {
		account, err := h.persistTwitchLinkedAccount(ctx, result)
		if err != nil {
			return nil, err
		}
		response.Account = account
	}

	return response, nil
}

func (h handler) persistSiteLoginSession(ctx context.Context, result *twitchoauth.CallbackResult) (*siteLoginResult, error) {
	if h.appState == nil || h.appState.Sessions == nil {
		return nil, fmt.Errorf("session store is not configured")
	}
	if result.Validation == nil {
		return nil, fmt.Errorf("missing twitch token validation payload")
	}

	login := strings.TrimSpace(result.Validation.Login)
	displayName := login
	avatarURL := ""
	if result.UserInfo != nil {
		if preferred := strings.TrimSpace(result.UserInfo.PreferredUsername); preferred != "" {
			displayName = preferred
		}
		avatarURL = strings.TrimSpace(result.UserInfo.Picture)
	}
	if displayName == "" {
		displayName = strings.TrimSpace(result.Validation.UserID)
	}

	access, err := webaccess.EvaluateDashboardAccess(ctx, h.appState, result.Validation.UserID)
	if err != nil {
		// Site login should still succeed even if dashboard access can't be evaluated yet.
		// Access can be recomputed later by the session status endpoint once dependencies recover.
		access = webaccess.DashboardAccess{}
	}

	sessionID, err := h.appState.Sessions.Create(ctx, session.UserSession{
		UserID:             result.Validation.UserID,
		Login:              login,
		DisplayName:        displayName,
		AvatarURL:          avatarURL,
		IsModerator:        access.IsModerator,
		IsBroadcaster:      access.IsBroadcaster,
		IsBotAccount:       access.IsBotAccount,
		IsEditor:           access.IsEditor,
		IsAdmin:            access.IsAdmin,
		CanAccessDashboard: access.CanAccessDashboard,
	})
	if err != nil {
		return nil, err
	}

	return &siteLoginResult{
		sessionID:          sessionID,
		canAccessDashboard: access.CanAccessDashboard,
	}, nil
}

func (h handler) handleRobloxCallback(ctx context.Context, stateKey, code string) (*callbackResponse, error) {
	if h.appState == nil || h.appState.Config == nil || !h.appState.Config.Roblox.Enabled {
		return nil, robloxoauth.ErrStateNotFound
	}

	result, err := h.appState.RobloxOAuth.HandleCallback(ctx, stateKey, code)
	if err != nil {
		return nil, err
	}

	account, err := h.persistRobloxLinkedAccount(ctx, result)
	if err != nil {
		return nil, err
	}

	return &callbackResponse{
		Provider:    "roblox",
		Flow:        string(result.Flow),
		Account:     account,
		UserInfo:    result.UserInfo,
		RequestedAt: result.RequestedAt,
	}, nil
}

func (h handler) handleSpotifyCallback(ctx context.Context, stateKey, code string) (*callbackResponse, error) {
	if h.appState == nil || h.appState.Config == nil || !h.appState.Config.Spotify.Enabled {
		return nil, spotifyoauth.ErrStateNotFound
	}

	result, err := h.appState.SpotifyOAuth.HandleCallback(ctx, stateKey, code)
	if err != nil {
		return nil, err
	}

	account, profile, err := h.persistSpotifyLinkedAccount(ctx, result)
	if err != nil {
		return nil, err
	}

	return &callbackResponse{
		Provider:    "spotify",
		Flow:        string(result.Flow),
		Account:     account,
		UserInfo:    profile,
		RequestedAt: result.RequestedAt,
	}, nil
}

func (h handler) handleStreamlabsCallback(ctx context.Context, stateKey, code string) (*callbackResponse, error) {
	if h.appState == nil || h.appState.Config == nil || !h.appState.Config.Streamlabs.Enabled {
		return nil, streamlabsoauth.ErrStateNotFound
	}

	result, err := h.appState.StreamlabsOAuth.HandleCallback(ctx, stateKey, code)
	if err != nil {
		return nil, err
	}

	account, err := h.persistStreamlabsLinkedAccount(ctx, result)
	if err != nil {
		return nil, err
	}

	return &callbackResponse{
		Provider:    "streamlabs",
		Flow:        string(result.Flow),
		Account:     account,
		RequestedAt: result.RequestedAt,
	}, nil
}

func (h handler) handleStreamElementsCallback(ctx context.Context, stateKey, code string) (*callbackResponse, error) {
	if h.appState == nil || h.appState.Config == nil || !h.appState.Config.StreamElements.Enabled {
		return nil, streamelementsoauth.ErrStateNotFound
	}

	result, err := h.appState.StreamElementsOAuth.HandleCallback(ctx, stateKey, code)
	if err != nil {
		return nil, err
	}

	account, err := h.persistStreamElementsLinkedAccount(ctx, result)
	if err != nil {
		return nil, err
	}

	return &callbackResponse{
		Provider:    "streamelements",
		Flow:        string(result.Flow),
		Account:     account,
		RequestedAt: result.RequestedAt,
	}, nil
}

func (h handler) persistTwitchLinkedAccount(ctx context.Context, result *twitchoauth.CallbackResult) (*twitchAccountResponse, error) {
	if result.Validation == nil {
		return nil, fmt.Errorf("missing twitch token validation payload")
	}

	var (
		kind       postgres.TwitchAccountKind
		expectedID string
	)

	switch result.Flow {
	case twitchoauth.FlowStreamerConnect:
		kind = postgres.TwitchAccountKindStreamer
		expectedID = h.appState.Config.Main.StreamerID
	case twitchoauth.FlowBotConnect:
		kind = postgres.TwitchAccountKindBot
		expectedID = h.appState.Config.Main.BotID
	default:
		return nil, fmt.Errorf("unsupported linked account flow %q", result.Flow)
	}

	if expectedID != "" && result.Validation.UserID != expectedID {
		return nil, fmt.Errorf("connected twitch account %s does not match configured %s id %s", result.Validation.UserID, kind, expectedID)
	}

	account := postgres.TwitchAccount{
		Kind:            kind,
		TwitchUserID:    result.Validation.UserID,
		Login:           result.Validation.Login,
		DisplayName:     result.Validation.Login,
		AccessToken:     result.Token.AccessToken,
		RefreshToken:    result.Token.RefreshToken,
		Scopes:          append([]string(nil), result.Token.Scope...),
		TokenType:       result.Token.TokenType,
		ExpiresAt:       result.Token.ExpiresAt(),
		LastValidatedAt: result.RequestedAt,
	}

	if err := h.appState.TwitchAccounts.Save(ctx, account); err != nil {
		return nil, err
	}

	return &twitchAccountResponse{
		Kind:         string(kind),
		TwitchUserID: account.TwitchUserID,
		Login:        account.Login,
		Scopes:       account.Scopes,
		ExpiresAt:    account.ExpiresAt,
	}, nil
}

func (h handler) persistRobloxLinkedAccount(ctx context.Context, result *robloxoauth.CallbackResult) (*robloxAccountResponse, error) {
	if result.UserInfo == nil {
		return nil, fmt.Errorf("missing roblox userinfo payload")
	}

	username := strings.TrimSpace(result.UserInfo.PreferredUsername)
	if username == "" {
		username = strings.TrimSpace(result.UserInfo.Name)
	}
	if username == "" {
		username = strings.TrimSpace(result.UserInfo.Nickname)
	}
	if username == "" {
		return nil, fmt.Errorf("roblox userinfo missing username")
	}

	displayName := strings.TrimSpace(result.UserInfo.Name)
	if displayName == "" {
		displayName = username
	}

	account := postgres.RobloxAccount{
		Kind:         postgres.RobloxAccountKindStreamer,
		RobloxUserID: strings.TrimSpace(result.UserInfo.Subject),
		Username:     username,
		DisplayName:  displayName,
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		Scope:        strings.TrimSpace(result.Token.Scope),
		TokenType:    strings.TrimSpace(result.Token.TokenType),
		ExpiresAt:    result.Token.ExpiresAt(),
	}

	if account.RobloxUserID == "" {
		return nil, fmt.Errorf("roblox userinfo missing subject")
	}

	if err := h.appState.RobloxAccounts.Save(ctx, account); err != nil {
		return nil, err
	}

	return &robloxAccountResponse{
		Kind:         string(account.Kind),
		RobloxUserID: account.RobloxUserID,
		Username:     account.Username,
		DisplayName:  account.DisplayName,
		Scope:        account.Scope,
		ExpiresAt:    account.ExpiresAt,
	}, nil
}

func (h handler) persistSpotifyLinkedAccount(ctx context.Context, result *spotifyoauth.CallbackResult) (*spotifyAccountResponse, *spotifyapi.UserProfile, error) {
	client := spotifyapi.NewClient(nil, result.Token.AccessToken)
	profile, err := client.GetCurrentUserProfile(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch spotify profile: %w", err)
	}

	account := postgres.SpotifyAccount{
		Kind:          postgres.SpotifyAccountKindStreamer,
		SpotifyUserID: profile.ID,
		DisplayName:   profile.DisplayName,
		Email:         "",
		Product:       profile.Product,
		Country:       profile.Country,
		AccessToken:   result.Token.AccessToken,
		RefreshToken:  result.Token.RefreshToken,
		Scope:         strings.TrimSpace(result.Token.Scope),
		TokenType:     strings.TrimSpace(result.Token.TokenType),
		ExpiresAt:     result.Token.ExpiresAt(),
	}

	if err := h.appState.SpotifyAccounts.Save(ctx, account); err != nil {
		return nil, nil, err
	}

	return &spotifyAccountResponse{
		Kind:          string(account.Kind),
		SpotifyUserID: account.SpotifyUserID,
		DisplayName:   account.DisplayName,
		Product:       account.Product,
		Country:       account.Country,
		Scope:         account.Scope,
		ExpiresAt:     account.ExpiresAt,
	}, profile, nil
}

func (h handler) persistStreamlabsLinkedAccount(ctx context.Context, result *streamlabsoauth.CallbackResult) (*streamlabsAccountResponse, error) {
	var (
		displayName string
		userID      string
		socketToken string
	)

	client := streamlabsapi.NewClient(nil, result.Token.AccessToken)
	if socket, err := client.GetSocketToken(ctx); err == nil && socket != nil {
		socketToken = strings.TrimSpace(socket.SocketToken)
	}

	if profile, err := client.GetUser(ctx); err == nil {
		if value, ok := profile["display_name"].(string); ok {
			displayName = strings.TrimSpace(value)
		}
		if displayName == "" {
			if value, ok := profile["name"].(string); ok {
				displayName = strings.TrimSpace(value)
			}
		}
		if value, ok := profile["id"].(string); ok {
			userID = strings.TrimSpace(value)
		}
	}

	account := postgres.StreamlabsAccount{
		Kind:             postgres.StreamlabsAccountKindStreamer,
		StreamlabsUserID: userID,
		DisplayName:      displayName,
		AccessToken:      result.Token.AccessToken,
		RefreshToken:     result.Token.RefreshToken,
		Scope:            strings.TrimSpace(result.Token.Scope),
		TokenType:        strings.TrimSpace(result.Token.TokenType),
		ExpiresAt:        result.Token.ExpiresAt(),
		SocketToken:      socketToken,
	}

	if err := h.appState.StreamlabsAccounts.Save(ctx, account); err != nil {
		return nil, err
	}

	return &streamlabsAccountResponse{
		Kind:             string(account.Kind),
		StreamlabsUserID: account.StreamlabsUserID,
		DisplayName:      account.DisplayName,
		Scope:            account.Scope,
		ExpiresAt:        account.ExpiresAt,
		SocketToken:      account.SocketToken,
	}, nil
}

func (h handler) persistStreamElementsLinkedAccount(ctx context.Context, result *streamelementsoauth.CallbackResult) (*streamElementsAccountResponse, error) {
	var (
		channelID   string
		provider    string
		username    string
		displayName string
	)

	client := streamelementsapi.NewClient(nil, result.Token.AccessToken)
	if channel, err := client.GetChannelMe(ctx); err == nil && channel != nil {
		channelID = strings.TrimSpace(channel.ID)
		provider = strings.TrimSpace(channel.Provider)
		username = strings.TrimSpace(channel.Username)
		displayName = strings.TrimSpace(channel.DisplayName)
		if displayName == "" {
			displayName = strings.TrimSpace(channel.Alias)
		}
	}
	if username == "" || displayName == "" {
		if user, err := client.GetCurrentUser(ctx); err == nil && user != nil {
			if username == "" {
				username = strings.TrimSpace(user.Username)
			}
			if displayName == "" {
				displayName = strings.TrimSpace(user.DisplayName)
			}
			if channelID == "" {
				channelID = strings.TrimSpace(user.ID)
			}
		}
	}

	account := postgres.StreamElementsAccount{
		Kind:         postgres.StreamElementsAccountKindStreamer,
		ChannelID:    channelID,
		Provider:     provider,
		Username:     username,
		DisplayName:  displayName,
		AccessToken:  result.Token.AccessToken,
		RefreshToken: result.Token.RefreshToken,
		Scope:        strings.TrimSpace(result.Token.Scope),
		TokenType:    strings.TrimSpace(result.Token.TokenType),
		ExpiresAt:    result.Token.ExpiresAt(),
	}

	if err := h.appState.StreamElementsAccounts.Save(ctx, account); err != nil {
		return nil, err
	}

	return &streamElementsAccountResponse{
		Kind:        string(account.Kind),
		ChannelID:   account.ChannelID,
		Provider:    account.Provider,
		Username:    account.Username,
		DisplayName: account.DisplayName,
		Scope:       account.Scope,
		ExpiresAt:   account.ExpiresAt,
	}, nil
}

func writeMethodNotAllowed(w http.ResponseWriter, variant authPageVariant, allowedMethods ...string) {
	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	writeOAuthErrorPage(w, "method not allowed", http.StatusMethodNotAllowed, variant)
}

func isSecureCookie(appState *state.State) bool {
	if appState == nil || appState.Config == nil {
		return false
	}

	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(appState.Config.Web.PublicURL)), "https://")
}

func writeAuthorizedPage(w http.ResponseWriter, statusCode int, variant authPageVariant, page authorizedPageData) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(statusCode)
	switch variant {
	case authPageVariantDashboard:
		_ = dashboardAuthorizedPage.Execute(w, page)
	default:
		_ = publicAuthorizedPage.Execute(w, page)
	}
}

func writeOAuthErrorPage(w http.ResponseWriter, message string, statusCode int, variant authPageVariant) {
	page := authorizedPageData{
		Title:          "Authorization failed",
		Eyebrow:        "OAuth error",
		Message:        message,
		PrimaryLabel:   "Back to home",
		PrimaryHref:    "/",
		SecondaryLabel: "Open integrations",
		SecondaryHref:  "/dashboard/integrations",
	}
	if variant == authPageVariantDashboard {
		page.PrimaryLabel = "Back to dashboard"
		page.PrimaryHref = "/dashboard"
	}
	writeAuthorizedPage(w, statusCode, variant, page)
}

func buildAuthorizedPageData(response *callbackResponse) authorizedPageData {
	page := authorizedPageData{
		Title:          "Authorization complete",
		Eyebrow:        "OAuth success",
		Message:        "The account was linked successfully and DankBot can use it now.",
		PrimaryLabel:   "Open integrations",
		PrimaryHref:    "/dashboard/integrations",
		SecondaryLabel: "Back to home",
		SecondaryHref:  "/",
	}

	switch response.Provider {
	case "twitch":
		page.Eyebrow = "Twitch linked"
		page.AccountLabel = "Twitch account"
		switch response.Account.(type) {
		case *twitchAccountResponse:
			account := response.Account.(*twitchAccountResponse)
			page.AccountValue = account.Login
			if response.Flow == string(twitchoauth.FlowStreamerConnect) {
				page.Title = "Streamer account linked"
				page.Message = "DankBot can now use your streamer account for broadcaster-side Twitch features."
			} else {
				page.Title = "Bot account linked"
				page.Message = "DankBot can now use your bot account for chat, moderation, and message sending."
			}
		default:
			page.Title = "Twitch account linked"
		}
	case "spotify":
		page.Title = "Spotify linked"
		page.Eyebrow = "Spotify connected"
		page.Message = "Spotify is connected and the now-playing module can use the linked account."
		page.AccountLabel = "Spotify account"
		if account, ok := response.Account.(*spotifyAccountResponse); ok {
			if strings.TrimSpace(account.DisplayName) != "" {
				page.AccountValue = account.DisplayName
			} else {
				page.AccountValue = "linked streamer account"
			}
		}
	case "roblox":
		page.Title = "Roblox linked"
		page.Eyebrow = "Roblox connected"
		page.Message = "Roblox is connected and DankBot can use the linked account for game and profile features."
		page.AccountLabel = "Roblox account"
		if account, ok := response.Account.(*robloxAccountResponse); ok {
			if strings.TrimSpace(account.DisplayName) != "" {
				page.AccountValue = account.DisplayName
			} else {
				page.AccountValue = account.Username
			}
		}
	case "streamlabs":
		page.Title = "Streamlabs linked"
		page.Eyebrow = "Streamlabs connected"
		page.Message = "Streamlabs is connected and DankBot can use the linked account for alerts and realtime events."
		page.AccountLabel = "Streamlabs account"
		if account, ok := response.Account.(*streamlabsAccountResponse); ok {
			page.AccountValue = strings.TrimSpace(account.DisplayName)
			if page.AccountValue == "" {
				page.AccountValue = "linked streamer account"
			}
		}
	case "streamelements":
		page.Title = "StreamElements linked"
		page.Eyebrow = "StreamElements connected"
		page.Message = "StreamElements is connected and DankBot can use the linked account for alerts and realtime events."
		page.AccountLabel = "StreamElements account"
		if account, ok := response.Account.(*streamElementsAccountResponse); ok {
			page.AccountValue = strings.TrimSpace(account.DisplayName)
			if page.AccountValue == "" {
				page.AccountValue = strings.TrimSpace(account.Username)
			}
			if page.AccountValue == "" {
				page.AccountValue = "linked streamer account"
			}
		}
	}

	return page
}

func preferredLoginName(result *twitchoauth.CallbackResult) string {
	if result == nil || result.Validation == nil {
		return ""
	}

	if result.UserInfo != nil {
		if preferred := strings.TrimSpace(result.UserInfo.PreferredUsername); preferred != "" {
			return preferred
		}
	}

	return strings.TrimSpace(result.Validation.Login)
}
