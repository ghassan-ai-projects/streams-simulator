package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type ReplayOptions struct {
	Run        string
	HostSocket string
	SessionID  string
	BarrierCap time.Duration
}

type ReplayStats struct {
	Records       int    `json:"records"`
	Skipped       int    `json:"skipped"`
	Barriers      int    `json:"barriers"`
	Refused       int    `json:"commands_refused"`
	FieldTrace    string `json:"field_trace_digest"`
	HostSessionID string `json:"session_id"`
}

type replayer struct {
	link    Link
	out     *domain.Transmitter
	in      domain.Receiver
	plan    domain.ReplayPlan
	session string
	version int
	stats   ReplayStats
	over    bool
}

func ReplayTrace(ctx context.Context, options ReplayOptions) (ReplayStats, error) {
	if options.Run == "" || options.HostSocket == "" || options.BarrierCap <= 0 || domain.WireID(options.SessionID) != options.SessionID {
		return ReplayStats{}, fmt.Errorf("%w: a replay needs a run folder, the host's socket, a barrier cap and a wire-safe session id", ErrBadOptions)
	}
	trace, digest, err := run.ReadFieldTrace(options.Run)
	if err != nil {
		return ReplayStats{}, fmt.Errorf("plantwire replay: %w", err)
	}
	plan, err := domain.PlanReplay(trace)
	if err != nil {
		return ReplayStats{}, fmt.Errorf("plantwire replay: %w", err)
	}
	return playPlan(ctx, options, plan, digest)
}

func playPlan(ctx context.Context, options ReplayOptions, plan domain.ReplayPlan, digest string) (ReplayStats, error) {
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	link, err := uds.Dial(ctx, options.HostSocket, options.BarrierCap)
	if err != nil {
		return ReplayStats{}, fmt.Errorf("plantwire replay: %w", err)
	}
	p := &replayer{link: link, out: domain.NewTransmitter(options.SessionID), plan: plan, session: options.SessionID}
	p.stats = ReplayStats{Skipped: plan.Skipped, FieldTrace: digest, HostSessionID: options.SessionID}
	played := p.play(ctx)
	if err := errors.Join(played, link.Close()); err != nil {
		return p.stats, fmt.Errorf("plantwire replay: %w", err)
	}
	return p.stats, nil
}

func (p *replayer) play(ctx context.Context) error {
	if err := p.handshake(ctx); err != nil {
		return err
	}
	for i, chunk := range p.plan.Chunks {
		if err := p.chunk(ctx, chunk, i == 0); err != nil {
			return err
		}
		if p.over {
			break
		}
	}
	return p.send(domain.NodeDeath(p.plan.Births-1, "shutdown"))
}

func (p *replayer) handshake(ctx context.Context) error {
	if err := p.send(p.plan.Chunks[0].Records[0]); err != nil {
		return err
	}
	inbound, met, err := p.link.AwaitHandshake(ctx, isType("STATE"))
	if err != nil || !met || len(inbound) != 1 || inbound[0].Invalid != nil {
		return fmt.Errorf("plantwire replay: no valid STATE first after NBIRTH: %w", errors.Join(err, ErrHostLost))
	}
	return p.adopt(inbound[0])
}

func (p *replayer) adopt(state domain.Inbound) error {
	version, err := domain.Negotiate(state.Message, domain.Version)
	if err != nil {
		return fmt.Errorf("plantwire replay: handshake: %w", err)
	}
	if p.in.Accept(state.Seq) != domain.InOrder {
		return fmt.Errorf("plantwire replay: %w: the first host record has seq %d", ErrHostGap, state.Seq)
	}
	p.out.UseVersion(version)
	p.version = version
	return nil
}

func (p *replayer) chunk(ctx context.Context, chunk domain.ReplayChunk, first bool) error {
	records := chunk.Records
	if first {
		records = records[1:]
	}
	if len(records) == 0 {
		return nil
	}
	if err := p.send(records...); err != nil {
		return err
	}
	return p.barrier(ctx, chunk.WorldTimeNS)
}

func (p *replayer) send(records ...domain.Record) error {
	for _, record := range records {
		p.out.Queue(record)
	}
	sent := p.out.Drain()
	if err := p.link.Send(sent); err != nil {
		return fmt.Errorf("plantwire replay: send: %w", err)
	}
	p.stats.Records += len(records)
	return nil
}

func (p *replayer) barrier(ctx context.Context, worldNS int64) error {
	if err := p.send(domain.Sync(p.out.LastSeq(), worldNS)); err != nil {
		return err
	}
	through := p.out.LastSeq()
	inbound, met, err := p.link.Await(ctx, func(m domain.Message) bool { return m.Type == "HSYNC" && m.ThroughSeq >= through })
	if err != nil || !met {
		return fmt.Errorf("plantwire replay: barrier: %w", errors.Join(err, ErrHostLost))
	}
	p.stats.Barriers++
	return p.answer(inbound)
}
