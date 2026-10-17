import { describe, expect, it } from "vitest";
import { NavigationStore } from "./navigationStore";
import { memoryLocation } from "../../../test/location";

describe("navigation", () => {
  it("starts from the address and writes every move to it", () => {
    const location = memoryLocation("#/review?world=w-2");
    const navigation = new NavigationStore(location);
    expect(navigation.snapshot()).toEqual({ screen: "review", worldId: "w-2" });
    navigation.go({ screen: "plant", domain: "domains/a.domain.json", hall: null });
    expect(location.hash).toBe("#/plant?domain=domains%2Fa.domain.json");
    expect(location.replaced).toBe(false);
  });

  it("replaces the address without a new history entry when asked", () => {
    const location = memoryLocation("#/plant?domain=d&launch=1");
    const navigation = new NavigationStore(location);
    navigation.replace({ screen: "plant", domain: "d", hall: null });
    expect([location.hash, location.replaced]).toEqual(["#/plant?domain=d", true]);
    expect(navigation.snapshot()).toEqual({ screen: "plant", domain: "d", hall: null });
  });

  it("follows the address when the user goes back", () => {
    const location = memoryLocation("#/");
    const navigation = new NavigationStore(location);
    location.hash = "#/plant?domain=d&hall=h01";
    location.fire();
    expect(navigation.snapshot()).toEqual({ screen: "plant", domain: "d", hall: "h01" });
  });

  it("holds a move away from an edited composition until the user decides", () => {
    const location = memoryLocation("#/edit?composition=c1");
    const navigation = new NavigationStore(location);
    let dirty = true;
    navigation.guardUnsaved(() => dirty);
    navigation.go({ screen: "workspace" });
    expect([navigation.snapshot(), location.hash]).toEqual([
      { screen: "edit", composition: "c1" },
      "#/edit?composition=c1",
    ]);
    expect(navigation.leaving.snapshot()).toEqual({ route: { screen: "workspace" }, replace: false });
    navigation.stay();
    expect([navigation.snapshot(), navigation.leaving.snapshot()]).toEqual([
      { screen: "edit", composition: "c1" },
      null,
    ]);
    navigation.go({ screen: "workspace" });
    dirty = false;
    navigation.leave();
    expect([navigation.snapshot(), location.hash]).toEqual([{ screen: "workspace" }, "#/"]);
  });

  it("does not hold a move when nothing is unsaved or the composition stays the same", () => {
    const location = memoryLocation("#/edit?composition=c1");
    const navigation = new NavigationStore(location);
    navigation.guardUnsaved(() => true);
    navigation.go({ screen: "edit", composition: "c1" });
    expect(navigation.leaving.snapshot()).toBeNull();
    navigation.guardUnsaved(() => false);
    navigation.go({ screen: "workspace" });
    expect(navigation.snapshot()).toEqual({ screen: "workspace" });
  });

  it("removes a route's unsaved guard when its disposer runs", () => {
    const navigation = new NavigationStore(memoryLocation("#/edit?composition=c1"));
    const dispose = navigation.guardUnsaved(() => true);
    navigation.go({ screen: "workspace" });
    expect(navigation.leaving.snapshot()).not.toBeNull();
    navigation.stay();
    dispose();
    navigation.go({ screen: "workspace" });
    expect(navigation.snapshot()).toEqual({ screen: "workspace" });
  });

  it("does not let an old guard disposer remove a replacement guard", () => {
    const navigation = new NavigationStore(memoryLocation("#/edit?composition=c1"));
    const disposeOld = navigation.guardUnsaved(() => false);
    navigation.guardUnsaved(() => true);
    disposeOld();
    navigation.go({ screen: "workspace" });
    expect(navigation.leaving.snapshot()).not.toBeNull();
  });

  it("puts the address back and holds the move when the user goes back with unsaved changes", () => {
    const location = memoryLocation("#/edit?composition=c1");
    const navigation = new NavigationStore(location);
    navigation.guardUnsaved(() => true);
    location.hash = "#/";
    location.fire();
    expect([navigation.snapshot(), location.hash, location.replaced]).toEqual([
      { screen: "edit", composition: "c1" },
      "#/edit?composition=c1",
      true,
    ]);
    expect(navigation.leaving.snapshot()).toEqual({ route: { screen: "workspace" }, replace: false });
  });

  it("publishes a move once although the address announces it again", () => {
    const location = memoryLocation("#/");
    const navigation = new NavigationStore(location);
    let published = 0;
    navigation.subscribe(() => {
      published += 1;
    });
    navigation.go({ screen: "review", worldId: "w-2" });
    location.fire();
    expect(published).toBe(1);
    expect(navigation.snapshot()).toEqual({ screen: "review", worldId: "w-2" });
  });
});
