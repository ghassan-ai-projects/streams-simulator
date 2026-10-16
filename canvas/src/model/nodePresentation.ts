import type { Pins } from "./ports";

export type RowCount = (pins: Pins, channels: readonly string[]) => number;

export interface ViewPresentation {
  readonly rows: RowCount;
  readonly note: (lines: number) => number;
}

export function plainRows(pins: Pins): number {
  return Math.max(2, pins.ins.length, pins.outs.length);
}
