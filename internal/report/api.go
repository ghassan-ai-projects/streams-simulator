package report

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type Report = layer.Report

type Comparison = layer.Comparison

type Directories = run.Directories

type Request = app.Request

const (
	FormatJSON     = app.JSONFormat
	FormatMarkdown = app.Markdown
)

var (
	ErrUnknownFormat   = app.ErrUnknownFormat
	ErrUnknownDocument = app.ErrUnknownDocument
	ErrRunNotEnded     = layer.ErrNotEnded
)
