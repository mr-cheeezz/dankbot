package coordinator

import "time"

type Lease struct {
	OwnerID   string
	TTL       time.Duration
	ExpiresAt time.Time
}
