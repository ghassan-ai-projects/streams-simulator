export interface SessionReading {
  readonly name: string;
  readonly unit: string;
  readonly value: number | string | boolean | null;
  readonly quality: string;
  readonly state: "good" | "stale" | "disconnected" | "invalid" | "untrusted";
}

export interface SessionDevice {
  readonly device: string;
  readonly board: string;
  readonly instance: string;
  readonly environment: "simulated" | "hardware" | "unknown";
  readonly bootId: string;
  readonly health: "online" | "offline";
  readonly reboots: number;
  readonly readings: readonly SessionReading[];
}

export interface SessionCommand {
  readonly sentAtNs: number;
  readonly updatedAtNs: number;
  readonly commandId: string;
  readonly device: string;
  readonly command: string;
  readonly bootId: string;
  readonly receipt: "pending" | "accepted" | "refused";
  readonly reason: string;
  readonly report: "pending" | "succeeded" | "failed" | "unknown";
  readonly effect: "unknown" | "observed";
  readonly effectEvidence: string;
}

export interface SessionLive {
  readonly commandsTruncated: boolean;
  readonly marks: readonly {
    readonly atNs: number;
    readonly kind: "command" | "connection";
    readonly label: string;
    readonly entity: string;
  }[];
  readonly atNs: number;
  readonly host: "unknown" | "online" | "offline";
  readonly gateway: "not_applicable" | "unknown" | "online" | "offline";
  readonly devices: readonly SessionDevice[];
  readonly commands: readonly SessionCommand[];
}

export function deviceReadings(device: SessionDevice) {
  return device.readings.map((reading) => ({
    channel: reading.name,
    value: reading.state === "good" && reading.value !== null ? reading.value : reading.state,
    unit: reading.state === "good" ? reading.unit : "",
    atNs: 0,
  }));
}
