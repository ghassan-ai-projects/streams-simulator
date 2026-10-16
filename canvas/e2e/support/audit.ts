/** What a geometry audit examined, what it found, and whether it examined enough for its floor to mean anything. */
export interface Audit {
  readonly considered: number;
  readonly problems: readonly string[];
  readonly met: boolean;
}

/** Fails a clean-expectation audit when its selector matched too few elements to prove anything. */
export function guard(tested: Audit): readonly string[] {
  if (tested.met) return tested.problems;
  return [`the audit examined ${String(tested.considered)} elements, below its floor`, ...tested.problems];
}
