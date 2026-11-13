package domain

import "math"

// weightedError is |RUL^ − RUL*| / RUL* at time at, a late prediction (failure predicted after the true one) counting
// prognosisLateWeight times.
func weightedError(p prediction, failure, at int64) float64 {
	miss := float64(p.failureNS - failure)
	if miss > 0 {
		miss *= prognosisLateWeight
	}
	return math.Abs(miss) / float64(failure-at)
}

// withinAlpha holds when the share of the prediction inside ±α of the true RUL reaches β; a point prediction is all
// inside or all outside.
func withinAlpha(weighted float64) bool {
	inside := 0.0
	if weighted <= prognosisAlpha {
		inside = 1
	}
	return inside >= prognosisBeta
}

func horizon(predictions []prediction, failure int64) int64 {
	from := len(predictions)
	for from > 0 && withinAlpha(weightedError(predictions[from-1], failure, predictions[from-1].atNS)) {
		from--
	}
	if from == len(predictions) {
		return 0
	}
	return failure - predictions[from].atNS
}

func lambdaMetric(predictions []prediction, failure, until int64, lambda float64) LambdaMetric {
	first := predictions[0].atNS
	at := first + int64(lambda*float64(failure-first))
	m := LambdaMetric{Lambda: lambda, AtNS: at, Graded: at < until}
	if !m.Graded {
		return m
	}
	weighted := weightedError(inForceAt(predictions, at), failure, at)
	accuracy := 1 - weighted
	m.AlphaLambda, m.RelativeAccuracy = withinAlpha(weighted), &accuracy
	return m
}

func inForceAt(predictions []prediction, at int64) prediction {
	current := predictions[0]
	for _, p := range predictions {
		if p.atNS <= at {
			current = p
		}
	}
	return current
}
