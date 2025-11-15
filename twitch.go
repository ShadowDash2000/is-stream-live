package streamlive

import (
	"context"
	"time"
)

type TwitchChannelName string

type Twitch struct {
	token     *twitchToken
	limiter   *rateLimiter
	checkRate time.Duration

	onStreamChange *Hook[*StreamChangeEvent]
}

func NewTwitch(clientID, clientSecret string) Client {
	return &Twitch{
		token:   newTwitchToken(clientID, clientSecret),
		limiter: newRateLimiter(4),

		onStreamChange: &Hook[*StreamChangeEvent]{},
	}
}

func (t *Twitch) Start(ctx context.Context, channels []TwitchChannelName) error {

	return nil
}

func (t *Twitch) IsLive(channel string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *Twitch) OnStreamChange(f func(e *StreamChangeEvent) error) {
	t.onStreamChange.BindFunc(f)
}
