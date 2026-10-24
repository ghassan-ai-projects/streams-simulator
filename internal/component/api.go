package component

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/files"
	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Summary = layer.Summary

type Issue = model.Issue

type Detail = layer.Detail

type Composed = layer.Composed

type Fit = layer.Fit

type ExampleSummary = layer.ExampleSummary

type NewPlant = layer.NewPlant

type DomainAssumptions = layer.Assumptions

type AssumptionsInput = layer.AssumptionsInput

type TwinFiles = layer.TwinFiles

type DomainLineage = layer.Lineage

type DomainDiff = layer.Diff

type WorkspaceDirs = files.WorkspaceDirs

var (
	ErrUnknownExample = layer.ErrUnknownExample
	ErrNoKindFiles    = layer.ErrNoKindFiles
	ErrReadingRange   = layer.ErrReadingRange
)

const (
	FitUnitIncompatible = layer.FitUnitIncompatible
	FitRangeUnstated    = layer.FitRangeUnstated
	FitRangeNarrower    = layer.FitRangeNarrower
	FitReadingNotNumber = layer.FitReadingNotNumber
)

type DomainSummary = specs.Summary

var ErrNoLibrary = errors.New("component: a library built by NewLibrary is required")

const DirectoriesEnv = layer.DirectoriesEnv
