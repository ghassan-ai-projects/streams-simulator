export interface Range {
  readonly min: number;
  readonly max: number;
}

export interface AssumedReading {
  readonly channel: string;
  readonly unit: string;
  readonly valueType: string;
  readonly range: Range | null;
  readonly resolution: number;
  readonly noise: { readonly model: string; readonly sigma: number };
  readonly cadence: { readonly mode: string; readonly periodS: number | null };
  readonly instrument: string;
  readonly basis: string;
  readonly recordedIn: string;
  readonly suggestion: boolean | null;
}

export interface AssumedDrive {
  readonly from: string;
  readonly kind: string;
  readonly metric: string;
  readonly gain: number;
  readonly timeConstantS: number;
}

export interface AssumedCommand {
  readonly name: string;
  readonly riskClass: string;
  readonly safeStop: boolean;
  readonly response: string;
  readonly recordedIn: string;
  readonly drives: readonly AssumedDrive[];
  readonly deadTimeS: number;
  readonly timeConstantS: number;
  readonly statement: string;
  readonly suggestion: boolean | null;
}

export interface AssumedDynamics {
  readonly target: string;
  readonly form: string;
  readonly gain: number | null;
  readonly timeConstantS: number | null;
  readonly clamp: Range | null;
}

export interface AssumedFault {
  readonly id: string;
  readonly description: string;
  readonly onsetShape: string;
  readonly affects: readonly string[];
}

export interface AssumedKind {
  readonly kind: string;
  readonly implements: string;
  readonly instances: readonly string[];
  readonly readings: readonly AssumedReading[];
  readonly dynamics: readonly AssumedDynamics[];
  readonly commands: readonly AssumedCommand[];
  readonly faults: readonly AssumedFault[];
}

export interface NotModeled {
  readonly code: string;
  readonly scope: string;
  readonly sentence: string;
}

export interface Assumptions {
  readonly domain: { readonly id: string; readonly version: string; readonly digest: string };
  readonly origin: {
    readonly kind: "twin" | "composition" | "domain";
    readonly twin: { readonly id: string; readonly version: string; readonly status: string } | null;
  };
  readonly kinds: readonly AssumedKind[];
  readonly notModeled: readonly NotModeled[];
}
