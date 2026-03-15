package coordinator

import "time"

type Heartbeat struct {
	InstanceID string
	Interval   time.Duration
	LastSeenAt time.Time
}
