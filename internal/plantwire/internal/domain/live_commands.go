package domain

import "slices"

func (l *Live) command(id string) *LiveCommand {
	index := slices.IndexFunc(l.commands, func(c *LiveCommand) bool { return c.CommandID == id })
	if index < 0 {
		return nil
	}
	return l.commands[index]
}

func (l *Live) sent(r map[string]any, at int64) bool {
	if !l.remember(text(r, "command_id")) {
		return false
	}
	c := &LiveCommand{CommandID: text(r, "command_id"), Device: text(r, "device"), Environment: l.environmentOf(text(r, "device")),
		Command: text(r, "command"), BootID: text(r, "expected_boot_id"), Sent: true, Receipt: "pending", Report: "pending",
		Effect: "unknown", SentAtNS: at, UpdatedAtNS: at, policy: text(r, "policy_digest")}
	l.commands = append(l.commands, c)
	if len(l.commands) > maxLiveCommands {
		l.commands = slices.Clone(l.commands[1:])
		l.truncated = true
	}
	l.mark(at, "command", "command", c.Device+" command "+c.Command+" sent", c.Device)
	return true
}

func (l *Live) remember(id string) bool {
	if l.seen[id] {
		return false
	}
	l.seen[id] = true
	l.seenOrder = append(l.seenOrder, id)
	if len(l.seenOrder) > maxSeenCommands {
		delete(l.seen, l.seenOrder[0])
		l.seenOrder = l.seenOrder[1:]
	}
	return true
}

func (l *Live) environmentOf(device string) string {
	if d := l.devices[device]; d != nil {
		return d.Environment
	}
	return "unknown"
}

func (l *Live) answer(r map[string]any, at int64) bool {
	c := l.command(text(r, "command_id"))
	if c == nil || (text(r, "device") != "" && text(r, "device") != c.Device) {
		return false
	}
	if text(r, "type") == "DRESULT" {
		return c.BootID == text(r, "boot_id") && l.result(c, r, at)
	}
	return (c.BootID == text(r, "boot_id") || refusedBeforeBoot(r)) && l.receipt(c, r, at)
}

func refusedBeforeBoot(r map[string]any) bool {
	return text(r, "status") == "refused" && text(r, "boot_id") == noBoot
}

func (l *Live) receipt(c *LiveCommand, r map[string]any, at int64) bool {
	status := text(r, "status")
	if c.Receipt != "pending" || (status != "accepted" && status != "refused") {
		return false
	}
	c.Receipt, c.Reason, c.UpdatedAtNS = status, text(r, "reason"), at
	c.PolicyDigest = refusedPolicy(c)
	l.mark(at, "command", "command", c.Device+" command "+c.Command+" "+status, c.Device)
	return true
}

func refusedPolicy(c *LiveCommand) string {
	if c.Receipt == "refused" && c.Reason == "policy_not_allowed" && policyDigest.MatchString(c.policy) {
		return c.policy
	}
	return ""
}

func (l *Live) result(c *LiveCommand, r map[string]any, at int64) bool {
	if c.Receipt != "accepted" || c.Report != "pending" {
		return false
	}
	c.Report, c.UpdatedAtNS = text(r, "reported_outcome"), at
	l.mark(at, "command", "command", c.Device+" reported command "+c.Command+" "+c.Report, c.Device)
	return true
}
