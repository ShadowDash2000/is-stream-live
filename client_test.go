package streamlive

import (
	"context"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func Test_ClientStartTracking(t *testing.T) {
	clientID := envOrSkip(t, "TWITCH_CLIENT_ID")
	clientSecret := envOrSkip(t, "TWITCH_CLIENT_SECRET")
	login := envOrSkip(t, "TWITCH_USER_LOGIN")

	c := New(
		NewTwitch(clientID, clientSecret),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.StartTracking(ctx, []string{login}, time.Minute)
}
