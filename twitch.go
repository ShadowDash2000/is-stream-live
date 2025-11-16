package streamlive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Twitch struct {
	token      *twitchToken
	limiter    *rateLimiter
	httpClient *http.Client
	started    bool
	liveCache  map[string]bool
	logins     map[string]struct{}
	mu         sync.RWMutex

	onStreamChange *Hook[*StreamChangeEvent]
}

func NewTwitch(clientID, clientSecret string) Client {
	return &Twitch{
		token:      newTwitchToken(clientID, clientSecret),
		limiter:    newRateLimiter(10),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		liveCache:  make(map[string]bool),

		onStreamChange: &Hook[*StreamChangeEvent]{},
	}
}

func (t *Twitch) StartTracking(ctx context.Context, logins []string, checkRate time.Duration) error {
	if t.started {
		panic("Twitch client already started")
	}

	t.logins = make(map[string]struct{}, len(logins))
	for _, login := range logins {
		if login == "" {
			continue
		}
		t.logins[login] = struct{}{}
	}

	if err := t.updateLiveCache(ctx); err != nil {
		return err
	}

	t.started = true
	go func() {
		ticker := time.NewTicker(checkRate)

		for {
			select {
			case <-ctx.Done():
				t.started = false
				t.onStreamChange.UnbindAll()
				return
			case <-ticker.C:
				_ = t.updateLiveCache(ctx)
			}
		}
	}()

	return nil
}

func (t *Twitch) AddLogin(login string) {
	if login == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.logins[login] = struct{}{}
}

func (t *Twitch) RemoveLogin(login string) {
	if login == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.logins, login)
}

func (t *Twitch) OnStreamChange(f func(e *StreamChangeEvent) error) {
	t.onStreamChange.BindFunc(f)
}

func (t *Twitch) IsLive(ctx context.Context, login string) (bool, error) {
	base, _ := url.Parse("https://api.twitch.tv/helix/streams")
	q := base.Query()
	q.Set("user_login", login)
	base.RawQuery = q.Encode()

	res, err := t.request(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("streams query failed: %s", res.Status)
	}
	var data struct {
		Data []any `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return false, err
	}

	return len(data.Data) > 0, nil
}

func (t *Twitch) request(ctx context.Context, method, url string, body any) (*http.Response, error) {
	t.limiter.wait()

	token, err := t.token.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	var reqBodyReader *strings.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBodyReader = strings.NewReader(string(b))
	} else {
		reqBodyReader = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", t.token.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return t.httpClient.Do(req)
}

func (t *Twitch) updateLiveCache(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for login := range t.logins {
		live, err := t.IsLive(ctx, login)
		if err != nil {
			return err
		}

		if l, ok := t.liveCache[login]; ok && l == live {
			continue
		}

		t.liveCache[login] = live
		if err = t.onStreamChange.Trigger(&StreamChangeEvent{
			Channel: login,
			Live:    live,
		}); err != nil {
			return err
		}
	}
	return nil
}
