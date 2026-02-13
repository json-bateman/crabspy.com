package eventbus

import "sync"

type Bus struct {
	mu    sync.Mutex
	rooms map[string]map[chan struct{}]struct{}
}

func NewBus() *Bus {
	return &Bus{
		rooms: make(map[string]map[chan struct{}]struct{}),
	}
}

func (b *Bus) SubscribeRoom(code string) chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.rooms[code] == nil {
		b.rooms[code] = make(map[chan struct{}]struct{})
	}

	ch := make(chan struct{}, 1)
	b.rooms[code][ch] = struct{}{}
	return ch
}

func (b *Bus) UnsubscribeRoom(code string, ch chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.rooms[code]
	if subs == nil {
		return
	}

	delete(subs, ch)
	close(ch)

	if len(subs) == 0 {
		delete(b.rooms, code)
	}
}

func (b *Bus) NotifyRoom(code string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.rooms[code] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
