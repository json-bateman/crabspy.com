package eventbus

import "sync"

type Bus struct {
	mu    sync.Mutex
	rooms map[int64]map[chan struct{}]struct{}
}

func NewBus() *Bus {
	return &Bus{
		rooms: make(map[int64]map[chan struct{}]struct{}),
	}
}

func (b *Bus) SubscribeRoom(roomID int64) chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.rooms[roomID] == nil {
		b.rooms[roomID] = make(map[chan struct{}]struct{})
	}

	ch := make(chan struct{}, 1)
	b.rooms[roomID][ch] = struct{}{}
	return ch
}

func (b *Bus) UnsubscribeRoom(roomID int64, ch chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.rooms[roomID]
	if subs == nil {
		return
	}

	delete(subs, ch)
	close(ch)

	if len(subs) == 0 {
		delete(b.rooms, roomID)
	}
}

func (b *Bus) NotifyRoom(roomID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.rooms[roomID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
