export function liveWire(atNs = 10e9) {
  return {
    type: "live",
    seq: 10,
    stream: "observer",
    at_ns: atNs,
    host: "online",
    gateway: "online",
    marks: [{ at_ns: atNs, kind: "command", label: "set_fan accepted", entity: "zone-01" }],
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
