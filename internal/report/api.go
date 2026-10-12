package report

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

type Report = layer.Report

const (
	FormatJSON     = app.JSONFormat
	FormatMarkdown = app.Markdown
)

var ErrUnknownFormat = app.ErrUnknownFormat
