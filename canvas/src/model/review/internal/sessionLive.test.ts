import { describe, expect, it } from "vitest";
import type { SessionCommand, SessionDevice, SessionLive, SessionReading } from "./sessionLive";
import { deviceOf, deviceReadings, originOf, readingWords } from "./sessionLive";
import {
  acceptanceDetail,
  acceptanceWords,
  commandSide,
  deviceLine,
  effectWords,
  refusalWords,
  updateAgeWords,
} from "./sessionWords";

const reading = (state: SessionReading["state"], quality: string, value: SessionReading["value"] = 28) => ({
  name: "temperature_c",
  unit: "degC",
  value,
  quality,
  state,
});

const board: SessionDevice = {
  device: "sensor",
  board: "uno-07",
  instance: "zone-01",
  environment: "hardware",
  bootId: "61",
  health: "online",
  reboots: 0,
  readings: [reading("good", "good")],
};

const command: SessionCommand = {
  sentAtNs: 1,
  updatedAtNs: 2,
  commandId: "c-1",
  device: "fan",
  environment: "simulated",
  command: "set_fan",
  bootId: "none",
  receipt: "refused",
  reason: "local_control",
  report: "pending",
  effect: "unknown",
  effectEvidence: "",
  policyDigest: "",
};

describe("session readings", () => {
  it.each([
    [reading("good", "good"), { valueText: "28 degC", stateWords: "good" }],
    [reading("simulated", "good"), { valueText: "28 degC", stateWords: "simulated · good" }],
    [reading("stale", "stale"), { valueText: "—", stateWords: "stale" }],
    [reading("untrusted", "warming_up"), { valueText: "—", stateWords: "untrusted · warming up" }],
    [reading("good", "", null), { valueText: "—", stateWords: "good" }],
  ])("words a reading without showing a value that is not current", (shown, words) => {
    expect(readingWords(shown)).toEqual(words);
  });

  it("puts the short state in place of the value on the node readout", () => {
    const device = { ...board, readings: [reading("good", "good"), reading("untrusted", "warming_up")] };
    expect(deviceReadings(device)).toEqual([
      { channel: "temperature_c", value: 28, unit: "degC" },
      { channel: "temperature_c", value: "untrusted", unit: "" },
    ]);
  });
});

describe("session devices and origins", () => {
  const live = { devices: [board] } as unknown as SessionLive;

  it("finds a device by instance and maps its environment to one origin", () => {
    expect(deviceOf(live, "zone-01")).toBe(board);
    expect(deviceOf(live, "fan-01")).toBeNull();
    expect(deviceOf(null, "zone-01")).toBeNull();
    expect(originOf(board)).toEqual({
      instance: "zone-01",
      kind: "real",
      board: "uno-07",
      device: "sensor",
      bootId: "61",
    });
    expect(originOf({ ...board, environment: "unknown" })).toBeNull();
    expect(originOf({ ...board, instance: "" })).toBeNull();
  });

  it("words health, reboots and the age of the last update", () => {
    expect(deviceLine({ ...board, reboots: 1 })).toBe("Online · board uno-07 · boot 61 · 1 reboot");
    expect(deviceLine({ ...board, health: "simulated", reboots: 2 })).toBe(
      "Simulated · no link · board uno-07 · 2 reboots",
    );
    expect(deviceLine(board)).toBe("Online · board uno-07 · boot 61");
    expect(updateAgeWords(12_400)).toBe("Last live snapshot 12 s ago");
    expect(updateAgeWords(45_000)).toBe("No live snapshot for 45 s");
  });
});

describe("session commands", () => {
  it("names the side, the refusal and the effect without reading the binary's text", () => {
    expect(commandSide(command)).toBe("Simulated");
    expect(commandSide({ ...command, environment: "hardware" })).toBe("Real board");
    expect(commandSide({ ...command, environment: "unknown" })).toBe("Origin unknown");
    expect(refusalWords("local_control")).toBe("Refused at the node: local control");
    expect(refusalWords("board_busy")).toBe("board busy");
    expect(effectWords(command)).toBe("Effect unknown");
    expect(effectWords({ ...command, effect: "observed" })).toBe("Host observed");
  });

  it("tells a refusal at the node from a refusal of the device or the host", () => {
    expect(acceptanceWords(command)).toBe("Refused at the node: local control");
    expect(acceptanceDetail(command)).toBe("");
    const device = { ...command, reason: "not_ready" };
    expect(acceptanceWords(device)).toBe("Refused");
    expect(acceptanceDetail(device)).toBe("Not ready");
    const waiting = { ...command, receipt: "pending", reason: "" } as const;
    expect(acceptanceWords(waiting)).toBe("Unconfirmed");
    expect(acceptanceDetail(waiting)).toBe("");
    expect(refusalWords("expired")).toBe("Expired before it was approved or accepted");
    expect(refusalWords("node_busy")).toBe("The node was busy and may still apply this later");
  });
});
