package domain

// admit returns "" when the command may execute, or a device reject_code. All
// target/operation/bound knowledge comes from the data-defined capability
// catalog, never a code branch.
func (d *Device) admit(command map[string]any, now int64) string {
	if reject := d.admitFreshness(command, now); reject != "" {
		return reject
	}
	return d.admitCommandParameters(command)
}

func (d *Device) admitFreshness(command map[string]any, now int64) string {
	if boot, _ := command["expected_boot_id"].(string); boot != d.bootID {
		return "wrong_boot"
	}
	notBefore, _ := command["not_before_mono_us"].(float64)
	expiresAfter, _ := command["expires_after_ms"].(float64)
	return freshnessWindow(notBefore, expiresAfter, now)
}

func (d *Device) admitSafeStop(target string, rawParams map[string]any, expiresAfter float64) string {
	if d.capabilities == nil || !d.capabilities.hasSafeStop(target) {
		return "wrong_target"
	}
	if len(rawParams) != 0 {
		return "out_of_range"
	}
	stop := d.capabilities.safeStops[target]
	if stop.expiresAfterMS > 0 && expiresAfter > float64(stop.expiresAfterMS) {
		return "expired"
	}
	return ""
}

func admitCapabilityParams(capa TargetCapability, rawParams map[string]any) string {
	if len(rawParams) != len(capa.Bounds)+len(capa.StringValues) {
		return "out_of_range"
	}
	if reject := admitNumericBounds(capa.Bounds, rawParams); reject != "" {
		return reject
	}
	return admitStringValues(capa.StringValues, rawParams)
}

func (d *Device) admitTargetOperation(target, operation string, params map[string]any, expires float64) string {
	capability, ok := d.capabilities.target(target)
	if !ok {
		return "wrong_target"
	}
	if capability.ExpiresAfterMS > 0 && expires > float64(capability.ExpiresAfterMS) {
		return "expired"
	}
	if operation != capability.Operation {
		return "unknown_operation"
	}
	return admitCapabilityParams(capability, params)
}

func freshnessWindow(notBefore, expires float64, now int64) string {
	// Zero is the immediate-dispatch sentinel; other values are boot-relative.
	if notBefore > 0 {
		if float64(now) < notBefore {
			return "not_ready"
		}
		if float64(now) > notBefore+expires*1000 {
			return "expired"
		}
	}
	return ""
}

func admitNumericBounds(bounds map[string][2]float64, params map[string]any) string {
	for name, bound := range bounds {
		value, ok := params[name].(float64)
		if !ok || value < bound[0] || value > bound[1] {
			return "out_of_range"
		}
	}
	return ""
}

func admitStringValues(allowed map[string]map[string]struct{}, params map[string]any) string {
	for name, values := range allowed {
		value, ok := params[name].(string)
		if !ok {
			return "out_of_range"
		}
		if _, ok := values[value]; !ok {
			return "out_of_range"
		}
	}
	return ""
}

func (d *Device) admitCommandParameters(command map[string]any) string {
	expiresAfter, _ := command["expires_after_ms"].(float64)
	params, valid := strictParams(command["parameters"])
	if !valid {
		return "out_of_range"
	}
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	if operation == "safe_stop" {
		return d.admitSafeStop(target, params, expiresAfter)
	}
	return d.admitTargetOperation(target, operation, params, expiresAfter)
}
