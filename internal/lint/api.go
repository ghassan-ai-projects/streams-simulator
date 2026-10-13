package lint

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/lint/internal/domain"
)

type Options = layer.Options

type Report = layer.Report

type Finding = layer.Finding

var ErrNoSpec = layer.ErrNoSpec

var ErrInvalidHours = layer.ErrInvalidHours

const (
	DefaultHours = layer.DefaultHours
	DefaultSeed  = layer.DefaultSeed
)
