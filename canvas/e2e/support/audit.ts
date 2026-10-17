export interface Audit {
  readonly considered: number;
  readonly problems: readonly string[];
  readonly met: boolean;
}

export function guard(tested: Audit): readonly string[] {
  if (tested.met) return tested.problems;
  return [`the audit examined ${String(tested.considered)} elements, below its floor`, ...tested.problems];
}
