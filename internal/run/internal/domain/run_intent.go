package domain

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func CheckRunPlan(plan model.RunPlan) error {
	if plan.RequestedEndNS <= 0 {
		return fmt.Errorf("run: plan requested_end_ns %d must be positive", plan.RequestedEndNS)
	}
	if plan.Rate < 0 || math.IsNaN(plan.Rate) || math.IsInf(plan.Rate, 0) {
		return fmt.Errorf("run: plan rate %v must be a finite number, 0 or more", plan.Rate)
	}
	if err := model.CheckPlanText(plan.Name, plan.Purpose); err != nil {
		return fmt.Errorf("run: plan %w", err)
	}
	return nil
}

func CheckTermination(termination model.Termination) error {
	if !model.IsTerminationReason(termination.Reason) {
		return fmt.Errorf("run: termination reason %q is not one of %v", termination.Reason, model.TerminationReasons())
	}
	return nil
}
