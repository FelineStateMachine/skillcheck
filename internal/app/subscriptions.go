package app

import "sync"

type OperationID uint64

type OperationEvent[T any] struct {
	ID       OperationID
	Progress *Progress
	Result   *T
	Err      error
}

// Subscription is a bounded, non-blocking bridge between application work and
// presentation loops. Progress may be coalesced; terminal events are retained.
type Subscription[T any] struct {
	C      <-chan OperationEvent[T]
	ch     chan OperationEvent[T]
	once   sync.Once
	closed chan struct{}
}

func NewSubscription[T any](capacity int) *Subscription[T] {
	if capacity < 1 {
		capacity = 1
	}
	ch := make(chan OperationEvent[T], capacity)
	return &Subscription[T]{C: ch, ch: ch, closed: make(chan struct{})}
}

func (s *Subscription[T]) Publish(event OperationEvent[T]) {
	select {
	case <-s.closed:
		return
	default:
	}
	select {
	case s.ch <- event:
	default:
		select {
		case <-s.ch:
		default:
		}
		s.ch <- event
	}
}

func (s *Subscription[T]) Close() { s.once.Do(func() { close(s.closed) }) }
