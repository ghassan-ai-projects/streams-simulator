package component

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

type Summary = layer.Summary

type Detail = layer.Detail

type Composed = layer.Composed

type ExampleSummary = layer.ExampleSummary

var ErrUnknownExample = layer.ErrUnknownExample

type DomainSummary = specs.Summary

var ErrNoLibrary = errors.New("component: a library built by NewLibrary is required")

const DirectoriesEnv = layer.DirectoriesEnv
