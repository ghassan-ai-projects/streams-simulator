import type {
  ControlReply,
  InterruptRecord,
  InterruptState,
  PendingApproval,
  SessionDevice,
  SessionLive,
} from "../model/review";

export const emptyState: InterruptState = { holds: [], approve: [], pending: [], devices: [] };

export const fanWaits: PendingApproval = {
  commandId: "cmd-7",
  device: "fan",
  command: "set_fan",
  deadlineNs: 130e9,
  queuedAtNs: 10e9,
  clock: "world",
};

export const sensorWaits: PendingApproval = {
  commandId: "cmd-8",
  device: "sensor",
  command: "calibrate",
  deadlineNs: 1_700_000_045e9,
  queuedAtNs: 1,
  clock: "wall",
};

export const busyState: InterruptState = {
  holds: [
    { scope: { kind: "board", id: "uno-07" }, latched: false },
    { scope: { kind: "device", id: "fan" }, latched: true },
  ],
  approve: [{ kind: "all", id: "" }],
  pending: [fanWaits, sensorWaits],
  devices: [
    { device: "fan", board: "mega-01", mode: "hold", latched: true },
    { device: "sensor", board: "uno-07", mode: "hold", latched: false },
    { device: "led", board: "mega-01", mode: "approve", latched: false },
  ],
};

export function replyOf(overrides: Partial<ControlReply> = {}): ControlReply {
  return {
    result: "applied",
    reason: "",
    op: "hold",
    commandId: "",
    stopped: [],
    skipped: [],
    state: busyState,
    ...overrides,
  };
}

export function recordOf(overrides: Partial<InterruptRecord> = {}): InterruptRecord {
  return {
    event: "control",
    nodeNs: 10e9,
    clock: "world",
    op: "hold",
    scope: { kind: "board", id: "mega-01" },
    enabled: null,
    commandId: "",
    device: "",
    command: "",
    verdict: null,
    resolution: null,
    result: null,
    reason: "",
    ...overrides,
  };
}

export const sensorDevice: SessionDevice = {
  device: "sensor",
  board: "uno-07",
  instance: "zone-01",
  environment: "hardware",
  bootId: "61",
  health: "online",
  reboots: 0,
  readings: [],
};

const fanDevice: SessionDevice = {
  device: "fan",
  board: "mega-01",
  instance: "fan-01",
  environment: "simulated",
  bootId: "1",
  health: "simulated",
  reboots: 0,
  readings: [],
};

export function liveOf(overrides: Partial<SessionLive> = {}): SessionLive {
  return {
    key: "k",
    commandsTruncated: false,
    marks: [],
    atNs: 10e9,
    host: "online",
    gateway: "online",
    devices: [sensorDevice, fanDevice],
    commands: [],
    interrupts: busyState,
    ...overrides,
  };
}
