import type { Route } from "../../../model/shell";

export const rightPanelName: Readonly<Record<Route["screen"], string>> = {
  workspace: "Inspector",
  plant: "Inspector",
  review: "Run details",
  edit: "Build",
  device: "Device details",
  capture: "Inspector",
};
