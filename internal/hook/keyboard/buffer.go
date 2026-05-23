package keyboard

import "sync"

// InputBuffer is a fixed-size ring buffer for KeyEvent values.
type InputBuffer struct {
	buf  [512]KeyEvent
	head uint32
	tail uint32
	mu   sync.Mutex
}

func (b *InputBuffer) Push(e KeyEvent) bool {
	b.mu.Lock()
	next := (b.tail + 1) % uint32(len(b.buf))
	if next == b.head {
		b.mu.Unlock()
		return false
	}
	b.buf[b.tail] = e
	b.tail = next
	b.mu.Unlock()
	return true
}

func (b *InputBuffer) Pop() (KeyEvent, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.head == b.tail {
		return KeyEvent{}, false
	}
	e := b.buf[b.head]
	b.head = (b.head + 1) % uint32(len(b.buf))
	return e, true
}
