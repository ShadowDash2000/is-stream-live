package streamlive

type Client interface {
	IsLive(channel string) (bool, error)
	OnStreamChange(func(e *StreamChangeEvent) error)
}

type StreamChangeEvent struct {
	Event
	Channel string
	Live    bool
}
