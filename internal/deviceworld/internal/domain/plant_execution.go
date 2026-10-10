package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (p *Plant) targetBinding(target string) (Binding, error) {
	binding, ok := p.bindings[target]
	if !ok {
		return Binding{}, fmt.Errorf("%w: no binding for target %q", device.ErrPlantUnavailable, target)
	}
	return binding, nil
}

func (p *Plant) prepareCommand(binding Binding, cmd device.PlantCommand) (int64, map[string]any, error) {
	if p.w == nil {
		return 0, nil, fmt.Errorf("%w: world is unavailable", device.ErrPlantUnavailable)
	}
	atNS, err := p.advance(cmd.AtMicros)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: advance world: %w", device.ErrPlantUnavailable, err)
	}
	args, err := binding.args(cmd.Params)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: build effector arguments: %w", device.ErrPlantUnavailable, err)
	}
	return atNS, args, nil
}

func (p *Plant) invokeWorld(effector, entity, command string, args map[string]any, atNS int64, missing string) (*world.InvokeResult, error) {
	result, err := p.w.InvokeEffector(effector, entity, command, args, atNS)
	if err != nil || result == nil {
		if errors.Is(err, world.ErrInterlockRefused) {
			return nil, fmt.Errorf("%w: %w", device.ErrPlantInterlocked, err)
		}
		if err == nil {
			err = fmt.Errorf("%s", missing)
		}
		return nil, fmt.Errorf("%w: %w", device.ErrPlantUnavailable, err)
	}
	return result, nil
}

func (p *Plant) completeCommand(binding Binding, result *world.InvokeResult, atNS int64) (device.PlantEffect, error) {
	if _, _, err := p.w.Advance(atNS + 1); err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world after effector: %w", device.ErrPlantUnavailable, err)
	}
	effect := device.PlantEffect{Energized: result.EffectApplied}
	if binding.valueState != "" {
		effect.Value = p.w.StateValue(binding.entity, binding.valueState, p.w.Clock())
	}
	return effect, nil
}

func (p *Plant) prepareSafeStop(binding Binding, atMicros int64) (int64, map[string]any, error) {
	if p.w == nil {
		return 0, nil, fmt.Errorf("%w: world is unavailable", device.ErrPlantUnavailable)
	}
	atNS, err := p.advance(atMicros)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: advance world for safe stop: %w", device.ErrPlantUnavailable, err)
	}
	args, err := binding.safeStopArgsForEntity()
	if err != nil {
		return 0, nil, fmt.Errorf("%w: build safe-stop arguments: %w", device.ErrPlantUnavailable, err)
	}
	return atNS, args, nil
}

func (p *Plant) invokeSafeStop(binding Binding, target string, atNS int64, args map[string]any) (device.PlantEffect, error) {
	if p.safeStops == nil {
		p.safeStops = map[string]int{}
	}
	p.safeStops[target]++
	command := fmt.Sprintf("safe-stop/%s/%d", target, p.safeStops[target])
	result, err := p.invokeWorld(binding.safeStopEffector, binding.entity, command, args, atNS, "world returned no safe-stop result")
	if err != nil {
		return device.PlantEffect{}, err
	}
	return p.completeSafeStop(binding, result, atNS)
}

func (p *Plant) completeSafeStop(binding Binding, result *world.InvokeResult, atNS int64) (device.PlantEffect, error) {
	if !result.Accepted {
		return device.PlantEffect{}, fmt.Errorf("%w: safe stop was not accepted: %s", device.ErrPlantUnavailable, result.Reason)
	}
	if !result.EffectApplied {
		return device.PlantEffect{}, fmt.Errorf("%w: safe stop was accepted without applying its effect: %s", device.ErrPlantUnavailable, result.Mode)
	}
	if _, _, err := p.w.Advance(atNS + 1); err != nil {
		return device.PlantEffect{}, fmt.Errorf("%w: advance world after safe stop: %w", device.ErrPlantUnavailable, err)
	}
	return device.PlantEffect{Value: stateValue(p.w, binding), Energized: false}, nil
}
