import { describe, expect, it, vi } from "vitest";
import { WatchStore } from "./watchStore";

const basin = { entity: "CT-1", channel: "cooling_tower.basin_temp" };
const power = { entity: "CH-1", channel: "chiller.power" };

describe("the watch list of a plant", () => {
  it("is empty for a plant nobody watched anything of", () => {
    expect(new WatchStore().listOf("factory")).toEqual([]);
  });

  it("keeps a separate list for each plant", () => {
    const store = new WatchStore();
    store.toggle("factory", basin);
    store.toggle("boiler", power);
    expect(store.listOf("factory")).toEqual([basin]);
    expect(store.listOf("boiler")).toEqual([power]);
  });

  it("adds, removes and clears, and tells subscribers once for each real change", () => {
    const store = new WatchStore();
    const heard = vi.fn();
    store.subscribe(heard);
    store.toggle("factory", basin);
    store.toggle("factory", power);
    store.toggle("factory", basin);
    expect(store.listOf("factory")).toEqual([power]);
    store.clear("factory");
    store.clear("factory");
    expect(store.listOf("factory")).toEqual([]);
    expect(heard).toHaveBeenCalledTimes(4);
  });

  it("remembers whether the panel is open and the window chosen, for the whole session", () => {
    const store = new WatchStore();
    expect(store.snapshot()).toMatchObject({ open: false, choice: null });
    store.show();
    store.hide();
    store.toggleOpen();
    store.setChoice("30");
    expect(store.snapshot()).toMatchObject({ open: true, choice: "30" });
  });
});
