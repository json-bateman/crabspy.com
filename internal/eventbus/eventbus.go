package eventbus

import "sync"

type Bus struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func NewBus() *Bus {
	// Empty structs takes zero bytes of memory
	// Which makes them ideal for notifications
	return &Bus{subs: make(map[chan struct{}]struct{})}
}

func (b *Bus) Subscribe() chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan struct{}, 1)
	b.subs[ch] = struct{}{}
	return ch
}

func (b *Bus) Unsubscribe(ch chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subs, ch)
	close(ch)
}

func (b *Bus) Notify() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- struct{}{}:
		default: // don't block if client is slow
		}
	}
}
