# Is Stream Live
Is Stream Live is a simple library for checking if a Twitch stream is live.\
For now, only Twitch is supported.

#### Install:

```bash
go get -u github.com/ShadowDash2000/is-stream-live
```

#### How to use:

```go
package main

import (
	"context"
	"time"

	streamlive "github.com/ShadowDash2000/is-stream-live"
)

func main() {
	streamTracker := streamlive.New(
		streamlive.NewTwitch("TWITCH_CLIENT_ID", "TWITCH_CLIENT_SECRET"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	if err := streamTracker.StartTracking(
		ctx,
		[]string{"twitch_user_login"},
		time.Minute,
	); err != nil {
		// error handling
	}

	// ...
}

```

### API keys

To get API keys, visit:
- Twitch API: https://dev.twitch.tv/console
