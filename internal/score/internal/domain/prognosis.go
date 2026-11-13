package domain

import (
	"cmp"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	prognosisAlpha      = 0.2
	prognosisBeta       = 0.5
	prognosisLateWeight = 2.0
)

var prognosisLambdas = []float64{0.5, 0.75}

type PrognosisMetrics struct {
	FailureNS     int64          `json:"failure_ns"`
	GradedUntilNS int64          `json:"graded_until_ns"`
	Predictions   int            `json:"predictions"`
	Alpha         float64        `json:"alpha"`
	Beta          float64        `json:"beta"`
	LateWeight    float64        `json:"late_weight"`
	HorizonNS     int64          `json:"horizon_ns,omitempty"`
	Lambdas       []LambdaMetric `json:"lambdas"`
}

type LambdaMetric struct {
	Lambda           float64  `json:"lambda"`
	AtNS             int64    `json:"at_ns"`
	Graded           bool     `json:"graded"`
	AlphaLambda      bool     `json:"alpha_lambda"`
	RelativeAccuracy *float64 `json:"relative_accuracy,omitempty"`
}

type prediction struct{ atNS, failureNS int64 }

func prognosis(detections []model.Detection, ev Evidence, gt *model.GroundTruthRecord) *PrognosisMetrics {
	failure, fails := failureOf(ev.DoNothing, gt.EntityID)
	if !fails {
		return nil
	}
	until := min(failure, firstEffectiveCall(ev.Calls, failure))
	predictions := gradedPredictions(detections, gt, until)
	if len(predictions) == 0 {
		return nil
	}
	return prognosisOf(predictions, failure, until)
}

func prognosisOf(predictions []prediction, failure, until int64) *PrognosisMetrics {
	m := &PrognosisMetrics{FailureNS: failure, GradedUntilNS: until, Predictions: len(predictions), Alpha: prognosisAlpha,
		Beta: prognosisBeta, LateWeight: prognosisLateWeight, HorizonNS: horizon(predictions, failure)}
	for _, lambda := range prognosisLambdas {
		m.Lambdas = append(m.Lambdas, lambdaMetric(predictions, failure, until, lambda))
	}
	return m
}

func failureOf(nothing *model.OutcomeFacts, entity string) (int64, bool) {
	if nothing == nil {
		return 0, false
	}
	at, fails := int64(math.MaxInt64), false
	for _, outcome := range nothing.Entities {
		for _, trip := range outcome.Trips {
			if outcome.EntityID == entity && trip.FirstNS < at {
				at, fails = trip.FirstNS, true
			}
		}
	}
	return at, fails
}

func firstEffectiveCall(calls []world.EffectorCall, failure int64) int64 {
	first := failure
	for _, call := range calls {
		if call.EffectApplied && call.AtNS < first {
			first = call.AtNS
		}
	}
	return first
}

func gradedPredictions(detections []model.Detection, gt *model.GroundTruthRecord, until int64) []prediction {
	var predictions []prediction
	for _, detection := range detections {
		at, atErr := model.ParseTime(detection.DetectedAt)
		failure, failureErr := model.ParseTime(detection.PredictedFailureAt)
		if detection.EntityID == gt.EntityID && atErr == nil && failureErr == nil && at >= gt.FirstObservableTimeNS && at < until {
			predictions = append(predictions, prediction{atNS: at, failureNS: failure})
		}
	}
	slices.SortStableFunc(predictions, func(a, b prediction) int { return cmp.Compare(a.atNS, b.atNS) })
	return predictions
}
