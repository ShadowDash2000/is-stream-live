package streamlive

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	StartTracking(ctx context.Context, logins []string, checkRate time.Duration) error
	IsLive(ctx context.Context, login string) (bool, error)
	OnStreamChange(func(e *StreamChangeEvent) error)
}

type StreamChangeEvent struct {
	Event
	Channel string
	Live    bool
}

type StreamLive struct {
	clients   []Client
	liveCache map[string]bool
	mu        sync.RWMutex

	onStreamChange *Hook[*StreamChangeEvent]
}

func New(clients ...Client) *StreamLive {
	return &StreamLive{
		clients:   clients,
		liveCache: make(map[string]bool),

		onStreamChange: &Hook[*StreamChangeEvent]{},
	}
}

// StartTracking will start a goroutine that will check for stream changes every checkRate.
// To subscribe to stream changes, use OnStreamChange.
func (s *StreamLive) StartTracking(ctx context.Context, logins []string, checkRate time.Duration) error {
	if err := s.updateLiveCache(ctx, logins); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(checkRate)

		for {
			select {
			case <-ctx.Done():
				s.onStreamChange.UnbindAll()
				return
			case <-ticker.C:
				_ = s.updateLiveCache(ctx, logins)
			}
		}
	}()

	return nil
}

// OnStreamChange
// Note: event will be triggered if at least one client reports that channel is live.
func (s *StreamLive) OnStreamChange(f func(e *StreamChangeEvent) error) {
	s.onStreamChange.BindFunc(f)
}

func (s *StreamLive) updateLiveCache(ctx context.Context, logins []string) error {
	clientsCache := make(map[string]map[int]bool, len(logins))
	for clientIndex, client := range s.clients {
		for _, login := range logins {
			live, err := client.IsLive(ctx, login)
			if err != nil {
				return err
			}

			if clientsCache[login] == nil {
				clientsCache[login] = make(map[int]bool, len(s.clients))
			}

			clientsCache[login][clientIndex] = live
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for login, clients := range clientsCache {
		live := false
		for _, clientLive := range clients {
			if clientLive {
				live = true
				break
			}
		}

		if l, ok := s.liveCache[login]; ok && l == live {
			continue
		}

		s.liveCache[login] = live
		if err := s.onStreamChange.Trigger(&StreamChangeEvent{
			Channel: login,
			Live:    live,
		}); err != nil {
			return err
		}
	}

	return nil
}
