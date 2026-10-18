import { describe, expect, it } from "vitest";
import { cooler, coolerPath, sensor, sensorPath } from "../../../test/twinFixture";
import {
  argumentLabel,
  driveFollows,
  effectWords,
  followChoices,
  followTarget,
  noFollowerHint,
  noSetpointHint,
  partsWords,
  settableChoices,
  slowTimeWords,
} from "./twinWords";
import type { Drive } from "./twinTypes";
import { choicesFor } from "./words";

const drive: Drive = { from: "m2", kind: coolerPath, metric: "m1", gain: -0.5, timeConstantS: 120 };

describe("twin wording", () => {
  it("offers the readings that can follow a setting by their name alone, across the devices of the twin", () => {
    const options = followChoices([cooler, sensor]);
    expect(options).toEqual([
      [`m1@${coolerPath}`, "temperature"],
      [`m1@${sensorPath}`, "humidity"],
    ]);
    expect(followChoices([])).toEqual([]);
    expect(followChoices([{ ...cooler, kind: { ...cooler.kind, metrics: [] } }])).toEqual([]);
    expect(noFollowerHint).toContain("Add a second reading in the Interface step");
  });

  it("adds the device to a reading name that two devices share", () => {
    const twin = {
      ...sensor,
      kind: { ...sensor.kind, metrics: sensor.kind.metrics.map((m) => ({ ...m, name: "temperature" })) },
    };
    expect(followChoices([cooler, twin]).map(([, label]) => label)).toEqual([
      "temperature (cooler)",
      "temperature (sensor)",
    ]);
  });

  it("splits and joins the reading that follows by device and key", () => {
    expect(followTarget(`m1@${coolerPath}`)).toEqual({ kind: coolerPath, metric: "m1" });
    expect(followTarget("")).toEqual({ kind: "", metric: "" });
    expect(driveFollows(drive)).toBe(`m1@${coolerPath}`);
    expect(driveFollows({ ...drive, kind: "", metric: "" })).toBe("");
  });

  it("offers the readings a command can set by key, with a way to leave an argument unused", () => {
    expect(settableChoices(cooler.kind)).toEqual([
      ["", "Not used"],
      ["m2", "duty"],
    ]);
    expect(settableChoices(sensor.kind)).toEqual([["", "Not used"]]);
  });

  it("describes an effect once it is complete, and counts the parts of a twin", () => {
    const words = (change: Partial<typeof drive>) => effectWords({ ...drive, ...change }, coolerPath, [cooler, sensor]);
    expect(words({})).toBe(
      "temperature moves toward its starting value plus -0.5 × duty, closing about two thirds of the gap every 120 seconds.",
    );
    expect(words({ gain: null })).toBe("");
    expect(words({ timeConstantS: null })).toBe("");
    expect(words({ from: "" })).toBe("");
    expect(words({ metric: "" })).toBe("");
    expect(words({ kind: sensorPath, metric: "m1" })).toContain("humidity moves toward");
    expect(words({ timeConstantS: 0 })).toBe(slowTimeWords);
    expect(words({ timeConstantS: -5, gain: null })).toBe("Choose a time constant above zero.");
    expect(partsWords(1, 2, 0)).toBe("1 channel · 2 effectors · 0 faults");
  });

  it("names an argument by what it is, and says what a setpoint needs when there is none", () => {
    expect(argumentLabel("target")).toBe("Reading set by the target value");
    expect(argumentLabel("lease_ms")).toBe("Reading set by the lease ms value");
    expect(noSetpointHint).toBe("Mark a reading as a setpoint and give it a range in the Interface step");
  });

  it("does not offer a second way to leave a choice blank", () => {
    expect(choicesFor(settableChoices(cooler.kind), "")).toEqual(settableChoices(cooler.kind));
    expect(choicesFor([["a", "A"]], "")).toEqual([
      ["", "Choose…"],
      ["a", "A"],
    ]);
    expect(choicesFor([["a", "A"]], "b")).toEqual([
      ["a", "A"],
      ["b", "b (not accepted)"],
    ]);
  });
});
