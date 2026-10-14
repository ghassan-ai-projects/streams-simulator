import { describe, expect, it } from "vitest";
import { PanelStore } from "./panelStore";

function preferences(stored: Record<string, string> = {}) {
  const values = new Map(Object.entries(stored));
  return {
    values,
    read: (key: string) => values.get(key) ?? null,
    write: (key: string, value: string) => {
      values.set(key, value);
    },
  };
}

describe("panels", () => {
  it("opens both side panels by default, toggles each and remembers the choice", () => {
    const stored = preferences();
    const panels = new PanelStore(stored);
    expect(panels.snapshot()).toEqual({ left: true, right: true });
    panels.toggle("left");
    expect(panels.snapshot()).toEqual({ left: false, right: true });
    expect(new PanelStore(stored).snapshot()).toEqual({ left: false, right: true });
    panels.toggle("left");
    panels.toggle("right");
    expect(stored.values.get("streamsim.canvas.panel.right")).toBe("closed");
    expect(new PanelStore(preferences({ "streamsim.canvas.panel.left": "open" })).snapshot().left).toBe(true);
  });
});
