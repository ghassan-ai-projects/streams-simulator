import { describe, expect, it } from "vitest";
import type { Watched } from "./watchList";
import { canWatch, isWatched, toggled, watchKey, watchLimit } from "./watchList";

const basin: Watched = { entity: "CT-1", channel: "cooling_tower.basin_temp" };
const power: Watched = { entity: "CH-1", channel: "chiller.power" };

describe("the watch list", () => {
  it("adds a channel at the end and removes it when toggled again", () => {
    const one = toggled([], basin);
    expect(one).toEqual([basin]);
    expect(toggled(one, power)).toEqual([basin, power]);
    expect(toggled(toggled(one, power), basin)).toEqual([power]);
    expect(isWatched(one, basin)).toBe(true);
    expect(isWatched(one, power)).toBe(false);
  });

  it("tells channels apart by instance and by channel", () => {
    expect(watchKey(basin)).not.toBe(watchKey({ entity: "CT-2", channel: basin.channel }));
    expect(isWatched([basin], { entity: "CT-1", channel: "cooling_tower.fan_power" })).toBe(false);
  });

  it("refuses a channel beyond the limit but still removes one when the list is full", () => {
    const full = Array.from({ length: watchLimit }, (_, at) => ({ entity: `E-${String(at)}`, channel: "x.y" }));
    expect(canWatch(full, basin)).toBe(false);
    expect(toggled(full, basin)).toBe(full);
    expect(canWatch(full, full[0] ?? basin)).toBe(true);
    expect(toggled(full, full[0] ?? basin)).toHaveLength(watchLimit - 1);
  });

  it("never changes the list it was given", () => {
    const list = [basin];
    toggled(list, power);
    expect(list).toEqual([basin]);
  });
});
