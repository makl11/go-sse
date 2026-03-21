package channel_mux

import (
	"log/slog"
	"sync"
)

type ChannelMux[T any] interface {
	NewOut() chan T
	AppendOut(out chan T)
	RemoveOut(out chan T)
}

type channelMux[T any] struct {
	rwm  *sync.RWMutex
	in   chan T
	outs map[chan T]struct{}
}

func NewChannelMux[T any](in chan T) ChannelMux[T] {
	mux := channelMux[T]{rwm: &sync.RWMutex{}, in: in, outs: make(map[chan T]struct{})}
	go func() {
		for x := range in {
			slog.Debug("Send message to clients", "clientCount", len(mux.outs))
			mux.rwm.Lock()
			for out := range mux.outs {
				out <- x
			}
			mux.rwm.Unlock()
		}
	}()
	return &mux
}

func (cm channelMux[T]) NewOut() chan T {
	ch := make(chan T)
	cm.AppendOut(ch)
	return ch
}

func (cm channelMux[T]) AppendOut(out chan T) {
	cm.rwm.Lock()
	cm.outs[out] = struct{}{}
	cm.rwm.Unlock()
}

func (cm channelMux[T]) RemoveOut(out chan T) {
	cm.rwm.Lock()
	delete(cm.outs, out)
	cm.rwm.Unlock()
}
