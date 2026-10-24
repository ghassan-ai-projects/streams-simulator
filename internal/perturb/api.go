package perturb

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

type Delivered = layer.Delivered

type KindInfo = layer.KindInfo

type ParamInfo = layer.ParamInfo

const CategoryDelivery = layer.CategoryDelivery

var ErrNoSpec = errors.New("perturb: a domain spec is required")
