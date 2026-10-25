import { describe, expect, it, vi } from "vitest";
import type { PreferenceStore } from "../../ports";
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

function viewerStorage(initial: Record<string, string> = {}) {
  const kept = new Map(Object.entries(initial));
  const store: PreferenceStore = {
    read: (key) => kept.get(key) ?? null,
    write: (key, value) => {
      kept.set(key, value);
    },
  };
  return { kept, store };
}

describe("the trends a viewer keeps across a reload", () => {
  const key = "streamsim.trends.factory";
  const exists = () => true;

  it("saves the pins, the panel state and the size of the plant under its own key as they change", () => {
    const { kept, store } = viewerStorage();
    const watch = new WatchStore(store);
    watch.toggle("factory", basin);
    watch.show();
    watch.setSize("tall");
    expect(JSON.parse(kept.get(key) ?? "null")).toEqual({ open: true, size: "tall", watched: [basin] });
    watch.clear("factory");
    expect(JSON.parse(kept.get(key) ?? "null")).toMatchObject({ watched: [] });
  });

  it("brings them back once for the plant, and drops pins that no longer exist", () => {
    const { store } = viewerStorage({
      [key]: JSON.stringify({ open: true, size: "full", watched: [basin, power] }),
    });
    const watch = new WatchStore(store);
    watch.restore("factory", (pin) => pin.entity === "CT-1");
    expect(watch.listOf("factory")).toEqual([basin]);
    expect(watch.snapshot()).toMatchObject({ open: true, size: "full" });
    watch.toggle("factory", power);
    watch.restore("factory", exists);
    expect(watch.listOf("factory")).toEqual([basin, power]);
  });

  it("does not replace pins the user made before the restore ran", () => {
    const { store } = viewerStorage({ [key]: JSON.stringify({ open: false, size: null, watched: [basin] }) });
    const watch = new WatchStore(store);
    watch.toggle("factory", power);
    watch.restore("factory", exists);
    expect(watch.listOf("factory")).toEqual([power]);
  });

  it("starts empty and closed when nothing, or something unreadable, was kept", () => {
    const empty = new WatchStore(viewerStorage().store);
    empty.restore("factory", exists);
    expect(empty.snapshot()).toMatchObject({ open: false, size: null });
    const broken = new WatchStore(viewerStorage({ [key]: "{not json" }).store);
    broken.restore("factory", exists);
    expect(broken.listOf("factory")).toEqual([]);
  });

  it("works without any storage", () => {
    const watch = new WatchStore();
    watch.restore("factory", exists);
    watch.toggle("factory", basin);
    watch.setSize("compact");
    expect(watch.snapshot()).toMatchObject({ size: "compact" });
    expect(watch.listOf("factory")).toEqual([basin]);
  });
});
