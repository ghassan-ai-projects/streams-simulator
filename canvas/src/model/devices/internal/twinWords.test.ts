import { describe, expect, it } from "vitest";
import { coolerDocument, kindOf, sensorDocument } from "../../../test/twinFixture";
import { driveFollows, effectWords, followChoices, followTarget, partsWords, settableChoices } from "./twinWords";
import { choicesFor } from "./words";

const cooler = kindOf(coolerDocument);
const sensor = kindOf(sensorDocument);
const drive = { from: "duty", kind: "cooler", metric: "temperature", gain: -0.5, timeConstantS: 120 };

describe("twin wording", () => {
  it("offers the readings that can follow a setting, across the devices of the twin", () => {
    expect(followChoices([cooler, sensor])).toEqual([
      ["cooler/temperature", "cooler · temperature"],
      ["sensor/humidity", "sensor · humidity"],
    ]);
    expect(followChoices([])).toEqual([]);
  });

  it("splits and joins the reading that follows", () => {
    expect(followTarget("cooler/temperature")).toEqual({ kind: "cooler", metric: "temperature" });
    expect(followTarget("")).toEqual({ kind: "", metric: "" });
    expect(driveFollows(drive)).toBe("cooler/temperature");
    expect(driveFollows({ ...drive, kind: "", metric: "" })).toBe("");
  });

  it("offers the readings a command can set, with a way to leave an argument unused", () => {
    expect(settableChoices(cooler)).toEqual([
      ["", "Not used"],
      ["duty", "duty"],
    ]);
    expect(settableChoices(sensor)).toEqual([["", "Not used"]]);
  });

  it("describes an effect once it is complete, and counts the parts of a twin", () => {
    expect(effectWords(drive)).toBe(
      "temperature moves toward its starting value plus -0.5 × duty, closing about two thirds of the gap every 120 seconds.",
    );
    expect(effectWords({ ...drive, gain: null })).toBe("");
    expect(effectWords({ ...drive, timeConstantS: null })).toBe("");
    expect(effectWords({ ...drive, from: "" })).toBe("");
    expect(effectWords({ ...drive, metric: "" })).toBe("");
    expect(partsWords(1, 2, 0)).toBe("1 channel · 2 effectors · 0 faults");
  });

  it("does not offer a second way to leave a choice blank", () => {
    expect(choicesFor(settableChoices(cooler), "")).toEqual(settableChoices(cooler));
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
