import type { ConnectState } from "../app/launch";
import type {
  ControlReply,
  InterruptRecord,
  InterruptState,
  PendingApproval,
  SessionDevice,
  SessionLive,
} from "../model/review";
import { interruptsWire, liveWire } from "./liveFixture";

export const emptyState: InterruptState = { holds: [], approve: [], pending: [], devices: [] };

export const fanWaits: PendingApproval = {
  commandId: "cmd-7",
  device: "fan",
  command: "set_fan",
  deadlineNs: 130e9,
  queuedAtNs: 10e9,
  clock: "world",
  contentDigest: `sha256:${"a".repeat(64)}`,
  args: [
    { name: "duty_permille", value: "300" },
    { name: "lease_ms", value: "5000" },
  ],
  argsTruncated: false,
};

export const sensorWaits: PendingApproval = {
  commandId: "cmd-8",
  device: "sensor",
  command: "calibrate",
  deadlineNs: 1_700_000_045e9,
  queuedAtNs: 1,
  clock: "wall",
  contentDigest: `sha256:${"b".repeat(64)}`,
  args: [],
  argsTruncated: false,
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
    unreachable: [],
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

export const offeredSession: Extract<ConnectState, { readonly phase: "live" }> = {
  phase: "live",
  domain: "mini",
  run: "c-1",
  check: null,
  binding: null,
  stopping: false,
  controls: true,
  note: null,
};

const fanWire = {
  device: "fan",
  board: "mega-01",
  instance: "fan-01",
  environment: "simulated",
  boot_id: "1",
  health: "simulated",
  reboots: 0,
  readings: [],
};

export function sessionSnapshot(interrupts?: object) {
  const wire = liveWire(10e9);
  return { ...wire, devices: [...wire.devices, fanWire], ...(interrupts === undefined ? {} : { interrupts }) };
}

const [waitingFan] = interruptsWire().pending;

export const busyInterrupts = interruptsWire({
  holds: [
    { kind: "board", id: "uno-07", latched: false },
    { kind: "device", id: "fan", latched: true },
  ],
  pending: [
    { ...waitingFan, deadline_ns: 130e9 },
    {
      ...waitingFan,
      command_id: "cmd-8",
      device: "sensor",
      command: "calibrate",
      deadline_ns: 1_700_000_045e9,
      clock: "wall",
    },
  ],
});
