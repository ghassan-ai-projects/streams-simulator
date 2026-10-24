import { counted } from "../../shared/words";
import { convertedNote, rangeWords } from "./instruments";
import type { Instrument } from "./twinTypes";

export const reproduced: readonly string[] = [
  "Each reading of the device with its name, type, unit, range and resolution, and the instrument chosen for it.",
  "Each command with its arguments, risk class and safe stop, and the response chosen for it: none, direct or delayed.",
  "That the twin offers exactly what the device offers (interface conformance).",
];

export const notReproduced: readonly string[] = [
  "Physical dynamics beyond the chosen first-order response.",
  "Saturation, unless it is declared as a range.",
  "Device firmware behaviour and timing jitter.",
  "Any device setting a twin does not accept, such as a scan class. A device that has one is left out of the twin.",
];

export const standInWords = "The response is a simple stand-in until it is compared with recorded field data.";

export const calibrationWords = "Calibration: not performed";

export const hardwareWords = "Hardware validation: not run";

export const hardwareNote =
  "This check does not read twin reports. A comparison with a recorded board run is shown in Review as agrees, diverges or cannot judge.";

export const suggestedWords = "Suggested";

export const chosenWords = "Your choice";

export interface Claims {
  readonly interface: "passed" | "failed" | "running";
  readonly plausibility: "passed" | "warnings" | "failed" | "running" | "unavailable";
  readonly warnings: number;
  readonly errors: number;
}

function interfacePart(state: Claims["interface"]): string {
  return `interface check ${state}`;
}

function plausibilityPart({ plausibility, warnings, errors }: Claims): string {
  if (plausibility === "unavailable") return "plausibility check not run";
  if (plausibility === "failed") return `plausibility check found ${counted(errors, "error")}`;
  if (plausibility === "warnings") return `plausibility check passed with ${counted(warnings, "warning")}`;
  return `plausibility check ${plausibility}`;
}

export function claimSummary(claims: Claims): string {
  const parts = [
    interfacePart(claims.interface),
    plausibilityPart(claims),
    "not calibrated",
    "not compared with a board",
  ];
  return `${parts.join("; ").replace(/^./, (first) => first.toUpperCase())}.`;
}

export function partialWords(excluded: readonly { readonly name: string; readonly reason: string }[]): string {
  if (excluded.length === 0) return "";
  return `This twin excludes ${excluded.map(({ name, reason }) => `${name} (${reason})`).join(", ")}.`;
}

export function suggestionNote(count: number): string | null {
  if (count === 0) return null;
  const verb = count === 1 ? "is still the suggested default" : "are still the suggested defaults";
  return `${counted(count, "choice")} ${verb}: review the instrument and response of each reading.`;
}

export function instrumentAssumptions(instrument: Instrument | undefined): string {
  if (instrument === undefined) return "Nothing is added to the true value.";
  const stated = [instrument.unit, rangeWords(instrument.min, instrument.max), convertedNote(instrument)];
  return `On record: ${stated.filter((part) => part !== "").join(", ")}.`;
}
