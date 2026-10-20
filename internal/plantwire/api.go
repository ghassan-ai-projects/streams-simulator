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

type DeviceFiles = domain.DeviceFiles

type TwinRules = domain.TwinRules

type TwinReportOptions = app.TwinReportOptions

type TwinReportDocument = domain.TwinReport

type Capture = domain.Capture

var (
	ErrKindFileMissing   = domain.ErrKindFileMissing
	ErrKindDrift         = domain.ErrKindDrift
	ErrNotCanonical      = domain.ErrNotCanonical
	ErrTwinRequirements  = domain.ErrTwinRequirements
	ErrDuplicateMember   = domain.ErrDuplicateMember
	ErrRefused           = domain.ErrRefused
	ErrRunFolderTaken    = domain.ErrRunFolderTaken
	ErrNoBirth           = domain.ErrNoBirth
	ErrBirthRefused      = domain.ErrBirthRefused
	ErrSealedStream      = domain.ErrSealedStream
	ErrNoReplayBirth     = domain.ErrNoReplayBirth
	ErrNotObserver       = domain.ErrNotObserver
	ErrPlanRefused       = app.ErrPlanRefused
	ErrHostPolicy        = app.ErrHostPolicy
	ErrHostInput         = app.ErrHostInput
	ErrHostLost          = app.ErrHostLost
	ErrHostGap           = app.ErrHostGap
	ErrSocketNotPrivate  = uds.ErrSocketNotPrivate
	ErrTwinNoJournal     = app.ErrTwinNoJournal
	ErrTwinJournalDigest = app.ErrTwinJournalDigest
	ErrTwinJournal       = domain.ErrTwinJournal
	ErrTwinTolerance     = domain.ErrTwinTolerance
	ErrTwinKinds         = domain.ErrTwinKinds
	ErrTwinNoDevices     = domain.ErrTwinNoDevices
	ErrTwinSpan          = domain.ErrTwinSpan
)
