package streamlive

type Handler[T Resolver] struct {
	Func func(T) error
	id   string
}

type Hook[T Resolver] struct {
	handlers []*Handler[T]
}

type Unsubscribe func()

func (h *Hook[T]) Bind(handler *Handler[T]) Unsubscribe {
	handler.id = generateHookId()
	h.handlers = append(h.handlers, handler)

	return func() {
		h.Unbind(handler.id)
	}
}

func (h *Hook[T]) BindFunc(fn func(e T) error) Unsubscribe {
	return h.Bind(&Handler[T]{
		Func: fn,
	})
}

func (h *Hook[T]) Unbind(idsToRemove ...string) {
	for _, id := range idsToRemove {
		for i := len(h.handlers) - 1; i >= 0; i-- {
			if h.handlers[i].id == id {
				h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
				break
			}
		}
	}
}

func (h *Hook[T]) Trigger(event T) error {
	event.setNextFunc(nil) // reset in case the event is being reused

	for i := len(h.handlers) - 1; i >= 0; i-- {
		handler := h.handlers[i]
		old := event.nextFunc()
		event.setNextFunc(func() error {
			event.setNextFunc(old)
			return handler.Func(event)
		})

	}

	err := event.Next()

	return err
}

func generateHookId() string {
	return PseudorandomString(20)
}
