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

type Scripted = host.Scripted

type Verify = host.Verify

type WireFault = domain.WireFault

var (
	ErrRefused          = domain.ErrRefused
	ErrHostLost         = app.ErrHostLost
	ErrHostGap          = app.ErrHostGap
	ErrSocketNotPrivate = uds.ErrSocketNotPrivate
)
