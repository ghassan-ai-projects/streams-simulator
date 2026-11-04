import { describe, expect, it } from "vitest";
import { PlaybackIndex } from "./playback";
import { probeKey } from "./trendProbe";

const watched = { entity: "CT-1", channel: "cooling_tower.basin_temp" };
const span = { startNs: 0, endNs: 600e9 };

function indexed(): PlaybackIndex {
  const index = new PlaybackIndex();
  [100, 200, 300].forEach((second) => {
    index.add({ atNs: second * 1e9, ...watched, value: second, unit: "degC" });
  });
  return index;
}

const press = (key: string, shift = false) => ({ key, shift });

describe("moving the probe on a plot with the keyboard", () => {
  it("steps to the next and the previous reading of that channel", () => {
    const index = indexed();
    expect(probeKey(index, watched, span, 150e9, press("ArrowRight"))).toEqual({ kind: "move", atNs: 200e9 });
    expect(probeKey(index, watched, span, 150e9, press("ArrowLeft"))).toEqual({ kind: "move", atNs: 100e9 });
    expect(probeKey(index, watched, span, 200e9, press("ArrowRight"))).toEqual({ kind: "move", atNs: 300e9 });
  });

  it("stays where it is at the first and last reading", () => {
    const index = indexed();
    expect(probeKey(index, watched, span, 300e9, press("ArrowRight"))).toBeNull();
    expect(probeKey(index, watched, span, 100e9, press("ArrowLeft"))).toBeNull();
  });

  it("moves by a twentieth of the window with Shift or Page keys, never outside the window", () => {
    const index = indexed();
    expect(probeKey(index, watched, span, 300e9, press("ArrowRight", true))).toEqual({ kind: "move", atNs: 330e9 });
    expect(probeKey(index, watched, span, 300e9, press("PageUp"))).toEqual({ kind: "move", atNs: 270e9 });
    expect(probeKey(index, watched, span, 590e9, press("PageDown"))).toEqual({ kind: "move", atNs: 600e9 });
    expect(probeKey(index, watched, span, 10e9, press("ArrowLeft", true))).toEqual({ kind: "move", atNs: 0 });
  });

  it("jumps to the ends of the window, commits on Enter or Space and clears on Escape", () => {
    const index = indexed();
    expect(probeKey(index, watched, span, 300e9, press("Home"))).toEqual({ kind: "move", atNs: 0 });
    expect(probeKey(index, watched, span, 300e9, press("End"))).toEqual({ kind: "move", atNs: 600e9 });
    expect(probeKey(index, watched, span, 300e9, press("Enter"))).toEqual({ kind: "commit" });
    expect(probeKey(index, watched, span, 300e9, press(" "))).toEqual({ kind: "commit" });
    expect(probeKey(index, watched, span, 300e9, press("Escape"))).toEqual({ kind: "clear" });
  });

  it("leaves every other key alone", () => {
    expect(probeKey(indexed(), watched, span, 300e9, press("a"))).toBeNull();
    expect(probeKey(indexed(), watched, span, 300e9, press("Tab"))).toBeNull();
  });
});
