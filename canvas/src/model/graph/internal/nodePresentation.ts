import type {} from "./types";
export type { ViewPresentation } from "./types";
import type { Pins } from "../../plant";

export function plainRows(pins: Pins): number {
  return Math.max(2, pins.ins.length, pins.outs.length);
}
