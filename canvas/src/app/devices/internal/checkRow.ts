export interface CheckRow {
  readonly id: "device" | "twin" | "scope" | "implements" | "lint" | "calibration" | "hardware";
  readonly state: "ok" | "warn" | "bad" | "pending" | "idle" | "neutral";
  readonly status: string;
  readonly text: string;
  readonly notes: readonly string[];
  readonly details: readonly string[];
}
