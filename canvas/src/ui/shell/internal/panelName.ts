import type { Route } from "../../../model/shell";

/** One name per right-hand panel, shared by the top-bar toggle and the keyboard help. */
export const rightPanelName: Readonly<Record<Route["screen"], string>> = {
  home: "Inspector",
  plant: "Inspector",
  run: "Run details",
  edit: "Build",
};
