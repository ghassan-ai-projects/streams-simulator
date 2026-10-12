package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (d *Director) RequireOperator(token string) error {
	_, err := d.ResolveOperator(token)
	return err
}

func (d *Director) OperatorNameplate(token string) (*rules.Nameplate, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.ReadNameplate(token)
}

func (d *Director) OperatorEffectors(token string) ([]rules.EffectorInfo, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.ListEffectors(token)
}

func (d *Director) OperatorInvoke(token, effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return nil, err
	}
	return v.Invoke(token, effector, entityID, commandID, args)
}

func (d *Director) OperatorReport(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) (string, error) {
	v, err := d.ResolveOperator(token)
	if err != nil {
		return "", err
	}
	return v.WorldID, v.Report(token, runID, quiescedThroughNS, verdict)
}
