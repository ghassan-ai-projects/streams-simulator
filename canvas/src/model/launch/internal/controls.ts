import { rateWords } from "./launch";

export type ControlName = "pause" | "resume" | "step" | "rate" | "fault" | "call";

export interface RunControl {
  readonly pending: ControlName | null;
  readonly refusal: string;
}

export const idleControl: RunControl = { pending: null, refusal: "" };

interface SpeedChoice {
  readonly value: string;
  readonly label: string;
}

const presets: readonly number[] = [1, 10, 60, 300, 3600];

const fastestValue = "0";

export function speedValue(rate: number | null): string {
  return rate === null ? fastestValue : String(rate);
}

export function speedChoices(rate: number | null): readonly SpeedChoice[] {
  const rates = rate === null || presets.includes(rate) ? presets : [...presets, rate].sort((a, b) => a - b);
  return [
    ...rates.map((each) => ({ value: String(each), label: rateWords(each) })),
    { value: fastestValue, label: "As fast as possible" },
  ];
}

const pendingText: Record<ControlName, string> = {
  pause: "Pausing…",
  resume: "Resuming…",
  step: "Stepping…",
  rate: "Changing the speed…",
  fault: "Injecting the fault…",
  call: "Calling the effector…",
};

export function pendingWords(name: ControlName | null): string {
  return name === null ? "" : pendingText[name];
}

export function refusalWords(message: string): string {
  const text = message.trim().replace(/[.\s]+$/, "");
  return `The simulator did not do that: ${text === "" ? "it gave no reason" : text}.`;
}

export const exploreOnlyFaultsWords = "Faults can only be injected live in Explore runs.";
