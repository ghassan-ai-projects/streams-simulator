package plantwire

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type Options = app.Options

type HostOptions = host.Options

type HostReport = host.Report

type RelayOptions = app.RelayOptions

type RelayStats = domain.RelayStats

type SimulatedOptions = app.SimulatedOptions

type ReplayOptions = app.ReplayOptions

type ReplayStats = app.ReplayStats
type ConnectOptions = app.ConnectOptions
type ConnectCheck = app.ConnectCheck
type ConnectResult = app.ConnectResult

type Scripted = host.Scripted

type Verify = host.Verify

type WireFault = domain.WireFault

type KindSummary = domain.KindSummary

type TwinRules = domain.TwinRules

type Capture = domain.Capture

var (
	ErrKindFileMissing  = domain.ErrKindFileMissing
	ErrKindDrift        = domain.ErrKindDrift
	ErrNotCanonical     = domain.ErrNotCanonical
	ErrTwinRequirements = domain.ErrTwinRequirements
	ErrDuplicateMember  = domain.ErrDuplicateMember
	ErrRefused          = domain.ErrRefused
	ErrNoBirth          = domain.ErrNoBirth
	ErrBirthRefused     = domain.ErrBirthRefused
	ErrSealedStream     = domain.ErrSealedStream
	ErrNoReplayBirth    = domain.ErrNoReplayBirth
	ErrNotObserver      = domain.ErrNotObserver
	ErrPlanRefused      = app.ErrPlanRefused
	ErrHostLost         = app.ErrHostLost
	ErrHostGap          = app.ErrHostGap
	ErrSocketNotPrivate = uds.ErrSocketNotPrivate
)
