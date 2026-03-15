package eventsub

import (
	"io"
	"net/http"

	twitchEventSub "github.com/mr-cheeezz/dankbot/pkg/twitch/eventsub"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}
	mux.Handle("/api/twitch/eventsub/webhook", http.HandlerFunc(h.webhook))
}

func (h handler) webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
