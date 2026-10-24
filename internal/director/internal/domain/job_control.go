package domain

func CheckJobRate(rate float64) error {
	if !validRate(rate) {
		return ToolErrorf(CodeInvalidArgs, "job: %s", rateProblem)
	}
	return nil
}

func CheckImmediateFault(entityID, fault string) error {
	return checkImmediate(entityID, "fault", fault)
}

func CheckImmediateCall(entityID, effector string) error {
	return checkImmediate(entityID, "effector", effector)
}

func checkImmediate(entityID, nameField, name string) error {
	switch {
	case entityID == "":
		return ToolErrorf(CodeInvalidArgs, "job: entity_id is required")
	case name == "":
		return ToolErrorf(CodeInvalidArgs, "job: %s is required", nameField)
	}
	return nil
}

type JobImmediate struct {
	JobStatus
	FaultID string          `json:"fault_id,omitempty"`
	Call    *JobCallOutcome `json:"call,omitempty"`
}
