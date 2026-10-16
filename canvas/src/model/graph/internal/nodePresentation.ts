import type {} from "../contract";
export type { ViewPresentation } from "../contract";
import type { Pins } from "../../ports";

export function plainRows(pins: Pins): number {
  return Math.max(2, pins.ins.length, pins.outs.length);
}
