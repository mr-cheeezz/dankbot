package public

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

const overlaySocketWriteTimeout = 5 * time.Second
const overlaySocketPingInterval = 25 * time.Second

func (h handler) overlaySocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()
	if err := h.writeOverlaySnapshot(ctx, conn); err != nil {
		_ = conn.Close(websocket.StatusInternalError, "initial snapshot failed")
		return
	}

	if h.appState == nil || h.appState.Redis == nil {
		h.keepOverlaySocketAlive(ctx, conn)
		return
	}

	subscription, err := h.appState.Redis.Subscribe(ctx, publicOverlayUpdatesChannel)
	if err != nil {
		h.keepOverlaySocketAlive(ctx, conn)
		return
	}
	defer subscription.Close()

	ticker := time.NewTicker(overlaySocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.writeOverlaySnapshot(ctx, conn); err != nil {
				return
			}
		case _, ok := <-subscription.Messages():
			if !ok {
				return
			}
			if err := h.writeOverlaySnapshot(ctx, conn); err != nil {
				return
			}
		}
	}
}

func (h handler) keepOverlaySocketAlive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(overlaySocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.writeOverlaySnapshot(ctx, conn); err != nil {
				return
			}
		}
	}
}

func (h handler) writeOverlaySnapshot(ctx context.Context, conn *websocket.Conn) error {
	payload := h.buildOverlaySnapshot(ctx)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	writeCtx, cancel := context.WithTimeout(ctx, overlaySocketWriteTimeout)
	defer cancel()

	return conn.Write(writeCtx, websocket.MessageText, body)
}
