package component

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Summary = layer.Summary

type Issue = model.Issue

type Detail = layer.Detail

type Composed = layer.Composed

type ExampleSummary = layer.ExampleSummary

type NewPlant = layer.NewPlant

var (
	ErrUnknownExample = layer.ErrUnknownExample
	ErrNoKindFiles    = layer.ErrNoKindFiles
)

type DomainSummary = specs.Summary

var ErrNoLibrary = errors.New("component: a library built by NewLibrary is required")

const DirectoriesEnv = layer.DirectoriesEnv
