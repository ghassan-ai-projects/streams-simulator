package device

// admit returns "" when the command may execute, or a device reject_code. All
// target/operation/bound knowledge comes from the data-defined capability
// catalog, never a code branch.
func (d *Device) admit(command map[string]any, now int64) string {
	if boot, _ := command["expected_boot_id"].(string); boot != d.bootID {
		return "wrong_boot"
	}
	notBefore, _ := command["not_before_mono_us"].(float64)
	expiresAfter, _ := command["expires_after_ms"].(float64)
	// Zero is the immediate-dispatch sentinel emitted by Agentic Stream.
	// Nonzero values are boot-relative freshness anchors.
	if notBefore > 0 {
		if float64(now) < notBefore {
			return "not_ready"
		}
		if float64(now) > notBefore+expiresAfter*1000 {
			return "expired"
		}
	}
	rawParams, validParams := strictParams(command["parameters"])
	if !validParams {
		return "out_of_range"
	}
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	if operation == "safe_stop" {
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
	capa, ok := d.capabilities.target(target)
	if !ok {
		return "wrong_target"
	}
	if capa.ExpiresAfterMS > 0 && expiresAfter > float64(capa.ExpiresAfterMS) {
		return "expired"
	}
	if operation != capa.Operation {
		return "unknown_operation"
	}
	if len(rawParams) != len(capa.Bounds)+len(capa.StringValues) {
		return "out_of_range"
	}
	for name, bounds := range capa.Bounds {
		v, ok := rawParams[name].(float64)
		if !ok || v < bounds[0] || v > bounds[1] {
			return "out_of_range"
		}
	}
	for name, allowed := range capa.StringValues {
		value, ok := rawParams[name].(string)
		if !ok {
			return "out_of_range"
		}
		if _, ok := allowed[value]; !ok {
			return "out_of_range"
		}
	}
	return ""
}
