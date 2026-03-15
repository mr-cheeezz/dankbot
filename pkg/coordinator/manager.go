package coordinator

import (
	"context"
	"time"

	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
)

type Manager struct {
	redis     *redispkg.Client
	lease     Lease
	heartbeat Heartbeat
}

func NewManager(instanceID string, leaseTTL, heartbeatInterval time.Duration, redisClient *redispkg.Client) *Manager {
	now := time.Now()

	return &Manager{
		redis: redisClient,
		lease: Lease{
			OwnerID:   instanceID,
			TTL:       leaseTTL,
			ExpiresAt: now.Add(leaseTTL),
		},
		heartbeat: Heartbeat{
			InstanceID: instanceID,
			Interval:   heartbeatInterval,
			LastSeenAt: now,
		},
	}
}

func (m *Manager) Start(ctx context.Context) error {
	_ = ctx
	m.heartbeat.LastSeenAt = time.Now()
	m.lease.ExpiresAt = time.Now().Add(m.lease.TTL)
	return nil
}

func (m *Manager) Lease() Lease {
	return m.lease
}

func (m *Manager) Heartbeat() Heartbeat {
	return m.heartbeat
}
