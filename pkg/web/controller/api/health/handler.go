package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type response struct {
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

func NewHandler(appState *state.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(response{
			Status:    "ok",
			StartedAt: appState.StartedAt,
		})
	})
}
