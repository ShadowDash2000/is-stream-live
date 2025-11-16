package streamlive

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	StartTracking(ctx context.Context, logins []string, checkRate time.Duration) error
	AddLogin(login string)
	RemoveLogin(login string)
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
	started   bool
	liveCache map[string]bool
	logins    map[string]struct{}
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
	if s.started {
		panic("Stream tracker already started")
	}

	s.logins = make(map[string]struct{}, len(logins))
	for _, login := range logins {
		if login == "" {
			continue
		}
		s.logins[login] = struct{}{}
	}

	if err := s.updateLiveCache(ctx); err != nil {
		return err
	}

	s.started = true
	go func() {
		ticker := time.NewTicker(checkRate)

		for {
			select {
			case <-ctx.Done():
				s.started = false
				s.onStreamChange.UnbindAll()
				return
			case <-ticker.C:
				_ = s.updateLiveCache(ctx)
			}
		}
	}()

	return nil
}

func (s *StreamLive) AddLogin(login string) {
	if login == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.logins[login] = struct{}{}
}

func (s *StreamLive) RemoveLogin(login string) {
	if login == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.logins, login)
	delete(s.liveCache, login)
}

// OnStreamChange
// Note: event will be triggered if at least one client reports that channel is live.
func (s *StreamLive) OnStreamChange(f func(e *StreamChangeEvent) error) {
	s.onStreamChange.BindFunc(f)
}

func (s *StreamLive) updateLiveCache(ctx context.Context) error {
	clientsCache := make(map[string]map[int]bool, len(s.logins))
	for clientIndex, client := range s.clients {
		for login := range s.logins {
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
