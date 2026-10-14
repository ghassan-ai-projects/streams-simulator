interface KeyRow {
  readonly keys: readonly string[];
  readonly does: string;
}

export interface KeyGroup {
  readonly scope: string;
  readonly rows: readonly KeyRow[];
}

const mod = "Ctrl or ⌘";

export const keyGroups: readonly KeyGroup[] = [
  {
    scope: "Everywhere",
    rows: [
      { keys: ["Tab", "Shift Tab"], does: "move between the top bar, the outline, the centre and the details panel" },
      { keys: ["[", "]"], does: "hide or show the outline panel, the details panel" },
      { keys: ["?"], does: "show or hide this list" },
      { keys: ["Escape"], does: "close this list" },
    ],
  },
  {
    scope: "Plant graph",
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
    rows: [
      { keys: ["Space"], does: "play or pause" },
      { keys: ["←", "→"], does: "on the time slider, step one second" },
      { keys: ["Shift ←", "Shift →"], does: "on the time slider, step one minute" },
      { keys: ["Page Up", "Page Down"], does: "on the time slider, step further" },
    ],
  },
  {
    scope: "Editor",
    rows: [
      { keys: [`${mod} Z`], does: "undo" },
      { keys: [`${mod} Shift Z`, `${mod} Y`], does: "redo" },
      { keys: [`${mod} S`], does: "save" },
    ],
  },
];
