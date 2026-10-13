package domain

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain/internal/files"
)

type Compiled = layer.Compiled

type FormatResult = files.FormatResult

type Entry = layer.Entry

type CoverageReport = layer.CoverageReport

var ErrUnknownDomain = layer.ErrUnknownDomain

var ErrNoCatalog = errors.New("domain: a catalog built by NewCatalog is required")
