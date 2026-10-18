package app

import (
	"context"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const controlBacklog = 16

type ControlCall struct {
	Entry domain.InterruptEntry
	Reply chan []byte
	Ctx   context.Context
}

type ControlPort struct {
	calls chan ControlCall
	mu    sync.Mutex
	view  map[string]any
}

func NewControlPort() *ControlPort {
	return &ControlPort{calls: make(chan ControlCall, controlBacklog), view: domain.EmptyInterruptView()}
}

func (p *ControlPort) Handle(ctx context.Context, line []byte) []byte {
	request, err := domain.DecodeControl(line)
	switch {
	case err != nil:
		return domain.ControlReply(domain.RefusedReply(domain.ReasonMalformed, p.snapshot()))
	case request.Op() == domain.OpState:
		return domain.ControlReply(domain.StateReply(p.snapshot()))
	}
	return p.submit(ctx, request)
}

func (p *ControlPort) submit(ctx context.Context, request domain.InterruptEntry) []byte {
	call, queued := p.enqueue(ctx, request)
	if !queued {
		return p.busy()
	}
	select {
	case reply := <-call.Reply:
		return reply
	case <-ctx.Done():
		return p.busy()
	}
}

func (p *ControlPort) enqueue(ctx context.Context, request domain.InterruptEntry) (ControlCall, bool) {
	call := ControlCall{Entry: request, Reply: make(chan []byte, 1), Ctx: ctx}
	select {
	case p.calls <- call:
		return call, true
	case <-ctx.Done():
		return call, false
	}
}

func (p *ControlPort) busy() []byte {
	return domain.ControlReply(domain.RefusedReply(domain.ReasonNodeBusy, p.snapshot()))
}

func (p *ControlPort) take() (ControlCall, bool) {
	if p == nil {
		return ControlCall{}, false
	}
	for {
		select {
		case call := <-p.calls:
			if call.Ctx.Err() == nil {
				return call, true
			}
		default:
			return ControlCall{}, false
		}
	}
}

func (p *ControlPort) Publish(view map[string]any) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.view = view
}

func (p *ControlPort) snapshot() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.view
}
