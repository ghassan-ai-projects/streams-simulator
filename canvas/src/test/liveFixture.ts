export function liveWire(atNs = 10e9) {
  return {
    type: "live",
    seq: 10,
    stream: "observer",
    at_ns: atNs,
    host: "online",
    gateway: "online",
    marks: [{ at_ns: atNs, kind: "command", subject: "command", label: "fan command set_fan accepted", entity: "fan" }],
    devices: [
      {
        device: "sensor",
        board: "uno-07",
        instance: "zone-01",
        environment: "hardware",
        boot_id: "61",
        health: "online",
        reboots: 0,
        readings: [{ name: "temperature_c", unit: "degC", value: 28, quality: "good", state: "good" }],
      },
    ],
    commands: [
      {
        command_id: "cmd-1",
        device: "fan",
        environment: "simulated",
        command: "set_fan",
        boot_id: "61",
        sent: true,
        sent_at_ns: 1e9,
        updated_at_ns: atNs,
        receipt: "accepted",
        reason: "",
        report: "succeeded",
        effect: "unknown",
        effect_evidence: "",
      },
    ],
  };
}

export function interruptsWire(overrides: Record<string, unknown> = {}) {
  return {
    holds: [{ kind: "board", id: "uno-07", latched: false }],
    approve: [{ kind: "all", id: "" }],
    pending: [
      {
        command_id: "cmd-7",
        device: "fan",
        command: "set_fan",
        deadline_ns: 1767225780e9,
        queued_at_ns: 1767225660e9,
        clock: "world",
      },
    ],
    devices: [
      { device: "sensor", board: "uno-07", mode: "hold", latched: false },
      { device: "fan", board: "mega-01", mode: "approve", latched: false },
    ],
    ...overrides,
  };
}

export function interruptWire(fields: Record<string, unknown>, atNs = 10e9) {
  return { type: "interrupt", seq: 11, stream: "observer", at_ns: atNs, node_ns: atNs, clock: "world", ...fields };
}

export function controlReplyWire(overrides: Record<string, unknown> = {}) {
  return { result: "applied", op: "hold", state: interruptsWire(), ...overrides };
}
