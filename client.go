package streamlive

import (
	"context"
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
