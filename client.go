package streamlive

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	StartTracking(ctx context.Context, logins []string, checkRate time.Duration)
	AddLogin(login string)
	RemoveLogin(login string)
	IsLive(ctx context.Context, login string) (bool, error)
	OnStreamChange(func(e *StreamChangeEvent) error)
	OnRequestError(func(e *RequestErrorEvent) error)
}

type StreamChangeEvent struct {
	Event
	Channel string
	Live    bool
}

type RequestErrorEvent struct {
	Event
	Error error
}

type StreamLive struct {
	clients   []Client
	ctx       context.Context
	cancel    context.CancelFunc
	liveCache map[string]bool
	logins    map[string]struct{}
	mu        sync.RWMutex

	onStreamChange *Hook[*StreamChangeEvent]
	onRequestError *Hook[*RequestErrorEvent]
}

func New(clients ...Client) *StreamLive {
	return &StreamLive{
		clients:   clients,
		liveCache: make(map[string]bool),

		onStreamChange: &Hook[*StreamChangeEvent]{},
		onRequestError: &Hook[*RequestErrorEvent]{},
	}
}

// StartTracking will start a goroutine that will check for stream changes every checkRate.
// To subscribe to stream changes, use OnStreamChange.
func (s *StreamLive) StartTracking(ctx context.Context, logins []string, checkRate time.Duration) {
	if s.ctx != nil {
		panic("Stream tracker already started")
	}

	s.logins = make(map[string]struct{}, len(logins))
	for _, login := range logins {
		if login == "" {
			continue
		}
		s.logins[login] = struct{}{}
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(checkRate)

		if err := s.updateLiveCache(ctx); err != nil {
			_ = s.onRequestError.Trigger(&RequestErrorEvent{Error: err})
		}

		for {
			select {
			case <-ctx.Done():
				s.ctx = nil
				s.cancel = nil
				s.onStreamChange.UnbindAll()
				s.onRequestError.UnbindAll()
				return
			case <-ticker.C:
				if err := s.updateLiveCache(ctx); err != nil {
					_ = s.onRequestError.Trigger(&RequestErrorEvent{Error: err})
				}
			}
		}
	}()
}

func (s *StreamLive) StopTracking() {
	if s.ctx == nil {
		return
	}
	s.cancel()
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

func (s *StreamLive) OnRequestError(f func(e *RequestErrorEvent) error) {
	s.onRequestError.BindFunc(f)
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
