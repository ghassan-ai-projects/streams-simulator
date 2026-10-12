package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// RequireOperator refuses a token that names no world, so a surface can
// check the capability before it decodes anything else of the call.
func (d *Director) RequireOperator(token string) error {
	_, err := d.ResolveOperator(token)
	return err
}

// OperatorNameplate returns the static nameplate of the token's world.
func (d *Director) OperatorNameplate(token string) (*rules.Nameplate, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.ReadNameplate(token)
}

// OperatorEffectors returns the effectors the token's world declares.
func (d *Director) OperatorEffectors(token string) ([]rules.EffectorInfo, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.ListEffectors(token)
}

// OperatorInvoke actuates an effector of the token's world; the consumer gets
// the acknowledgement only.
func (d *Director) OperatorInvoke(token, effector, entityID, commandID string, args map[string]any, atNS int64) (*model.EffectorAcknowledgement, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.Invoke(token, effector, entityID, commandID, args, atNS)
}

// OperatorReport records the consumer's quiescence and verdict and returns
// the id of the world it was recorded against. It never returns a score.
func (d *Director) OperatorReport(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) (string, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return "", err
	}
	return v.WorldID, v.Report(token, runID, quiescedThroughNS, verdict)
}
