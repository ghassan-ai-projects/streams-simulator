import type { Route } from "../../../model/shell";
import { rightPanelName } from "./panelName";

type Screen = Route["screen"];

interface KeyRow {
  readonly keys: readonly string[];
  readonly does: string;
}

export interface KeyGroup {
  readonly scope: string;
  readonly screens: readonly Screen[];
  readonly rows: readonly KeyRow[];
}

const mod = "Ctrl or ⌘";

function everywhere(screen: Screen): KeyGroup {
  const panel = rightPanelName[screen].toLowerCase();
  return {
    scope: "Everywhere",
    screens: ["workspace", "plant", "review", "edit", "device"],
    rows: [
      { keys: ["Tab", "Shift Tab"], does: `move between the top bar, the outline, the centre and the ${panel} panel` },
      { keys: ["[", "]"], does: `hide or show the outline panel, the ${panel} panel` },
      { keys: ["?"], does: "show or hide this list" },
      { keys: ["Escape"], does: "close this list" },
    ],
  };
}

const screenGroups: readonly KeyGroup[] = [
  {
    scope: "Plant graph",
    screens: ["plant", "review", "edit"],
    rows: [
      { keys: ["↑", "↓", "←", "→"], does: "move to the nearest instance in that direction" },
      { keys: ["Space"], does: "select the instance or hall" },
      { keys: ["Enter"], does: "select the instance, or open the hall" },
      { keys: ["Escape"], does: "clear the selection, then leave the hall" },
      { keys: ["+", "−"], does: "zoom in, zoom out" },
      { keys: ["0"], does: "fit the plant in view" },
    ],
  },
  {
    scope: "Run review",
    screens: ["review"],
    rows: [
      { keys: ["Space"], does: "play or pause" },
      { keys: [",", "."], does: "go to the previous event, the next event" },
      { keys: ["←", "→"], does: "on the time slider, step one second" },
      { keys: ["Shift ←", "Shift →"], does: "on the time slider, step one minute" },
      { keys: ["Page Up", "Page Down"], does: "on the time slider, step further" },
    ],
  },
  {
    scope: "Editor",
    screens: ["edit"],
    rows: [
      { keys: [`${mod} Z`], does: "undo" },
      { keys: [`${mod} Shift Z`, `${mod} Y`], does: "redo" },
      { keys: [`${mod} S`], does: "save" },
    ],
  },
  {
    scope: "Device editor",
    screens: ["device"],
    rows: [{ keys: [`${mod} S`], does: "save the device and ask the binary to check it" }],
  },
];

export function keyGroups(screen: Screen): readonly KeyGroup[] {
  return [everywhere(screen), ...screenGroups];
}
