package streamlive

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func envOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("env %s is not set; skipping", key)
	}
	return v
}

func Test_IsLive(t *testing.T) {
	clientID := envOrSkip(t, "TWITCH_CLIENT_ID")
	clientSecret := envOrSkip(t, "TWITCH_CLIENT_SECRET")
	login := envOrSkip(t, "TWITCH_USER_LOGIN")

	c := NewTwitch(clientID, clientSecret)
	_, err := c.IsLive(context.Background(), login)
	if err != nil {
		t.Fatalf("Test_IsLive: IsLive failed: %v", err)
	}
}

func Test_StartTracking(t *testing.T) {
	clientID := envOrSkip(t, "TWITCH_CLIENT_ID")
	clientSecret := envOrSkip(t, "TWITCH_CLIENT_SECRET")
	login := envOrSkip(t, "TWITCH_USER_LOGIN")

	c := NewTwitch(clientID, clientSecret)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.StartTracking(ctx, []string{login}, time.Minute)
}
