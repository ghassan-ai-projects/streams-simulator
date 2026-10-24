import type { InstanceOrigin } from "../../graph";
import type { OriginKind } from "../../shared/origin";
import { codeTitle, withUnit } from "../../shared/units";
import type { InterruptState } from "./interrupts";

type ReadingState = "good" | "simulated" | "stale" | "disconnected" | "invalid" | "untrusted";
export type DeviceHealth = "online" | "offline" | "simulated";
export type LinkState = "unknown" | "online" | "offline" | "not_applicable";
type DeviceEnvironment = "simulated" | "hardware" | "unknown";

export interface SessionReading {
  readonly name: string;
  readonly unit: string;
  readonly value: number | string | boolean | null;
  readonly quality: string;
  readonly state: ReadingState;
}

export interface SessionDevice {
  readonly device: string;
  readonly board: string;
  readonly instance: string;
  readonly environment: DeviceEnvironment;
  readonly bootId: string;
  readonly health: DeviceHealth;
  readonly reboots: number;
  readonly readings: readonly SessionReading[];
}

export interface SessionCommand {
  readonly sentAtNs: number;
  readonly updatedAtNs: number;
  readonly commandId: string;
  readonly device: string;
  readonly environment: DeviceEnvironment;
  readonly command: string;
  readonly bootId: string;
  readonly receipt: "pending" | "accepted" | "refused";
  readonly reason: string;
  readonly report: "pending" | "succeeded" | "failed" | "unknown";
  readonly effect: "unknown" | "observed";
  readonly effectEvidence: string;
  readonly policyDigest: string;
}

export interface LiveMark {
  readonly atNs: number;
  readonly kind: "command" | "connection";
  readonly subject: "host" | "gateway" | "device" | "command";
  readonly label: string;
  readonly entity: string;
}

export interface SessionLive {
  readonly key: string;
  readonly commandsTruncated: boolean;
  readonly marks: readonly LiveMark[];
  readonly atNs: number;
  readonly host: LinkState;
  readonly gateway: LinkState;
  readonly devices: readonly SessionDevice[];
  readonly commands: readonly SessionCommand[];
  readonly interrupts: InterruptState | null;
}

interface ReadingView {
  readonly channel: string;
  readonly value: number | string | boolean;
  readonly unit: string;
}

interface ReadingWords {
  readonly valueText: string;
  readonly stateWords: string;
  readonly unitTitle: string;
}

const environmentOrigins: Readonly<Record<DeviceEnvironment, OriginKind>> = {
  hardware: "real",
  simulated: "simulated",
  unknown: "unknown",
};

export function originKindOf(environment: DeviceEnvironment): OriginKind {
  return environmentOrigins[environment];
}

export function originOf(device: SessionDevice): InstanceOrigin | null {
  if (device.instance === "" || device.environment === "unknown") return null;
  const { instance, board, bootId } = device;
  return { instance, kind: originKindOf(device.environment), board, device: device.device, bootId };
}

export function deviceOf(live: SessionLive | null, instance: string): SessionDevice | null {
  return live?.devices.find((device) => device.instance === instance) ?? null;
}

const readingStateWords: Readonly<Record<ReadingState, string>> = {
  good: "good",
  simulated: "simulated",
  stale: "stale",
  disconnected: "disconnected",
  invalid: "invalid",
  untrusted: "untrusted",
};

function qualityWords(quality: string): string {
  return quality.replaceAll("_", " ");
}

function carriesValue(reading: SessionReading): boolean {
  return reading.state === "good" || reading.state === "simulated";
}

export function readingWords(reading: SessionReading): ReadingWords {
  const state = readingStateWords[reading.state];
  const plain = reading.quality === "" || reading.quality === reading.state;
  const quality = plain ? "" : ` · ${qualityWords(reading.quality)}`;
  const stateWords = `${state}${quality}`;
  if (!carriesValue(reading) || reading.value === null) return { valueText: "—", stateWords, unitTitle: "" };
  return {
    valueText: withUnit(String(reading.value), reading.unit),
    stateWords,
    unitTitle: codeTitle(reading.unit),
  };
}

export function deviceReadings(device: SessionDevice): readonly ReadingView[] {
  return device.readings.map((reading) => {
    const value = reading.value;
    if (carriesValue(reading) && value !== null) return { channel: reading.name, value, unit: reading.unit };
    return { channel: reading.name, value: readingStateWords[reading.state], unit: "" };
  });
}
