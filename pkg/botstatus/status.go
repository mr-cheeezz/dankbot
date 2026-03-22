package botstatus

import (
	"encoding/json"
	"time"
)

const RedisKey = "bot:runtime"

type Heartbeat struct {
	StartedAt     time.Time `json:"started_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	BotLogin      string    `json:"bot_login"`
	StreamerLogin string    `json:"streamer_login"`
	Version       string    `json:"version,omitempty"`
	Branch        string    `json:"branch,omitempty"`
	Revision      string    `json:"revision,omitempty"`
	CommitTime    string    `json:"commit_time,omitempty"`
}

func (h Heartbeat) Marshal() (string, error) {
	payload, err := json.Marshal(h)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func Unmarshal(value string) (*Heartbeat, error) {
	if value == "" {
		return nil, nil
	}

	var heartbeat Heartbeat
	if err := json.Unmarshal([]byte(value), &heartbeat); err != nil {
		return nil, err
	}

	return &heartbeat, nil
}
