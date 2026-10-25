import { describe, expect, it } from "vitest";
import { lineageModel as read } from "../../../test/provenanceModels";
import { deviceBadge } from "./deviceNotes";
import { chainOf, earlierVersion, runLines } from "./lineageChain";
import { plantOf, plantRoute } from "./plantRoutes";

const devices = ["kinds/fan.kind.json", "kinds/thermal_sensor.kind.json"];

function route(file: string, paths: readonly string[] = devices) {
  return plantRoute(plantOf(read, file), paths);
}

describe("the label of a device in the list", () => {
  it.each([
    ["a device that is not valid", "invalid", "kinds/fan.kind.json", { words: "Not valid", tone: "unfinished" }],
    [
      "a device whose own twin file exists",
      "twin-ready",
      "kinds/fan.kind.json",
      { words: "Twin created", tone: "verified" },
    ],
    [
      "a device with a twin of its own that the list could not see",
      "needs-twin",
      "kinds/heater.kind.json",
      { words: "Twin created", tone: "verified" },
    ],
    [
      "a device included in another device's twin",
      "needs-twin",
      "kinds/thermal_sensor.kind.json",
      { words: "Included in twin arduino-thermal-twin", tone: "verified" },
    ],
    [
      "a device that no twin includes",
      "needs-twin",
      "kinds/lonely.kind.json",
      { words: "Needs a twin", tone: "badge" },
    ],
    [
      "a device the lineage does not know",
      "needs-twin",
      "kinds/new.kind.json",
      { words: "Needs a twin", tone: "badge" },
    ],
  ] as const)("says %s", (_name, status, path, badge) => {
    expect(deviceBadge(status, path, read)).toEqual(badge);
  });

  it("falls back to the earlier wording when there is no lineage", () => {
    expect(deviceBadge("needs-twin", "kinds/fan.kind.json", null)).toEqual({ words: "Needs a twin", tone: "badge" });
  });

  it("names the first twin and counts the others", () => {
    const many = {
      ...read,
      devices: [
        {
          ...read.devices[0],
          file: "kinds/many.kind.json",
          twinsIncluding: [
            { twin: "a", file: "t1", kinds: 2 },
            { twin: "b", file: "t2", kinds: 3 },
            { twin: "c", file: "t3", kinds: 3 },
          ],
        },
      ],
    };
    expect(deviceBadge("needs-twin", "kinds/many.kind.json", many as typeof read).words).toBe(
      "Included in twin a and 2 more",
    );
  });
});

describe("why a plant cannot be edited, and the way on", () => {
  it("has nothing to say about a plant that its composition produces", () => {
    expect(route("domains/factory-utilities.domain.json")).toBeNull();
    expect(plantRoute(null, devices)).toBeNull();
  });

  it("sends a twin plant to the Behaviour step of the device the twin choices are named for", () => {
    expect(route("domains/arduino-thermal-twin.domain.json")).toMatchObject({
      headline: "Read-only: this plant is generated from a twin.",
      origin: "This plant was generated from a twin: its links are named driver_drives_driven and carry no ports.",
      action: { kind: "behaviour", device: "kinds/fan.kind.json" },
      label: "Open the twin's Behaviour step",
    });
    expect(route("domains/arduino-thermal-twin.domain.json")?.sentence).toContain("twins/fan.twin.json");
  });

  it("gives the binary's sentence and no button when no device carries the twin choices", () => {
    expect(route("domains/arduino-thermal-twin.domain.json", ["kinds/other.kind.json"])).toMatchObject({
      action: { kind: "none" },
      label: "",
    });
    expect(route("domains/arduino-thermal-twin.domain.json", ["kinds/other.kind.json"])?.sentence).toContain(
      "twins/fan.twin.json",
    );
  });

  it("sends a twin plant that no longer matches its twin choices to the same step", () => {
    const plant = plantOf(read, "domains/arduino-thermal-twin.domain.json") ?? fail();
    const twin = plant.twin ?? fail();
    const drifted = {
      ...plant,
      twin: { ...twin, status: "out_of_sync" as const },
      editable: {
        ...plant.editable,
        reason: "twin_out_of_sync" as const,
        route: { code: "rescaffold_twin" as const, sentence: "s" },
      },
    };
    expect(plantRoute(drifted, devices)).toMatchObject({
      headline: "Read-only: this plant no longer matches the twin it was generated from.",
      action: { kind: "behaviour", device: "kinds/fan.kind.json" },
    });
  });

  it("offers a copy of a shipped example", () => {
    expect(route("domains/chiller-pilot.domain.json")).toMatchObject({
      headline: "Read-only: a shipped example.",
      origin: "",
      action: { kind: "example", example: "chiller-pilot" },
      label: "Create an editable copy from this example",
    });
  });

  it("explains a hand-written plant with the binary's sentence and never converts it", () => {
    expect(route("domains/alarm-playback.domain.json")).toMatchObject({
      headline: "Read-only: written by hand as a domain file.",
      action: { kind: "none" },
      label: "",
    });
    expect(route("domains/alarm-playback.domain.json")?.sentence).toContain("copy the file as a new domain id");
  });

  it("offers to compose again a plant whose composition changed after it was composed", () => {
    expect(route("domains/campus-3.domain.json")).toMatchObject({
      headline: "This plant no longer matches its composition.",
      action: {
        kind: "recompose",
        composition: "compositions/campus-3.composition.json",
        domain: "domains/campus-3.domain.json",
      },
      label: "Compose again",
    });
  });
});

describe("where a plant comes from", () => {
  it("lists device files, twin choices and the plant, with their state", () => {
    const steps = chainOf(plantOf(read, "domains/arduino-thermal-twin.domain.json") ?? fail());
    expect(steps.map((step) => step.label)).toEqual(["Device files", "Twin choices", "This plant"]);
    expect(steps[0]).toMatchObject({ value: "fan, thermal_sensor", tone: "ok" });
    expect(steps[1]).toMatchObject({
      value: "arduino-thermal-twin 0.1.0",
      detail: "twins/fan.twin.json · in sync with this plant",
      tone: "ok",
    });
  });

  it("flags a composition that no longer produces the plant, with the binary's reason", () => {
    const steps = chainOf(plantOf(read, "domains/campus-3.domain.json") ?? fail());
    expect(steps[0]).toMatchObject({ label: "Composition", tone: "warn" });
    expect(steps[0]?.detail).toBe(
      "compositions/campus-3.composition.json · out of sync: it no longer produces this plant: version differs",
    );
  });

  it("flags device files that are missing", () => {
    const plant = plantOf(read, "domains/arduino-thermal-twin.domain.json") ?? fail();
    const lost = { ...plant, kinds: [...plant.kinds, { kind: "gone", found: false, file: "" }] };
    expect(chainOf(lost)[0]).toMatchObject({ detail: "missing: gone", tone: "warn" });
  });

  it("states which runs used this version of the plant and which an earlier one", () => {
    const lines = runLines(plantOf(read, "domains/factory-utilities.domain.json") ?? fail());
    expect(lines).toEqual([
      { id: "w-1", facts: "evaluate · version 0.1.0", verdict: "same as current", current: true },
      { id: "w-2", facts: "explore · version 0.1.0", verdict: earlierVersion, current: false },
    ]);
    expect(earlierVersion).toBe("ran on an earlier version of this plant");
  });

  it("says when a run has no record to compare", () => {
    const plant = plantOf(read, "domains/factory-utilities.domain.json") ?? fail();
    const bare = { ...plant, runs: [{ ...(plant.runs[0] ?? fail()), recorded: false, mode: "", domainVersion: "" }] };
    expect(runLines(bare)[0]).toMatchObject({ facts: "", verdict: "no run record to compare" });
  });
});

function fail(): never {
  throw new Error("the fixture plant is missing");
}
