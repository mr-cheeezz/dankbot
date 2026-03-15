package socket

import "encoding/json"

const (
	DefaultSocketURL = "https://sockets.streamlabs.com"
	EventName        = "event"
)

type Message struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}
