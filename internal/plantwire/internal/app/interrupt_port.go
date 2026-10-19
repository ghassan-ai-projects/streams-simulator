package app

import (
	"context"
	"slices"
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
	mu     sync.Mutex
	urgent []ControlCall
	queue  []ControlCall
	view   map[string]any
}

func NewControlPort() *ControlPort {
	return &ControlPort{view: domain.EmptyInterruptView()}
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
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dropAbandoned()
	list := &p.queue
	if request.Op() == domain.OpEStop {
		list = &p.urgent
	}
	if len(*list) >= controlBacklog {
		return call, false
	}
	*list = append(*list, call)
	return call, true
}

func (p *ControlPort) dropAbandoned() {
	live := func(call ControlCall) bool { return call.Ctx.Err() == nil }
	p.urgent = slices.DeleteFunc(p.urgent, func(call ControlCall) bool { return !live(call) })
	p.queue = slices.DeleteFunc(p.queue, func(call ControlCall) bool { return !live(call) })
}

func (p *ControlPort) busy() []byte {
	return domain.ControlReply(domain.RefusedReply(domain.ReasonNodeBusy, p.snapshot()))
}

func (p *ControlPort) take() (ControlCall, bool) {
	if p == nil {
		return ControlCall{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dropAbandoned()
	if call, ok := popFirst(&p.urgent); ok {
		return call, true
	}
	return popFirst(&p.queue)
}

func popFirst(list *[]ControlCall) (ControlCall, bool) {
	if len(*list) == 0 {
		return ControlCall{}, false
	}
	call := (*list)[0]
	*list = (*list)[1:]
	return call, true
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
