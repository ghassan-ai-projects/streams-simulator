import { describe, expect, it } from "vitest";
import { parseAlarmActivations } from "./reportAlarms";

describe("reported alarm activations", () => {
  it("keeps the binary's instants, entity and alarm identity", () => {
    const alarm = { state: "temp_alarm", priority: "high", activation_times_ns: [140e9, 530e9] };
    expect(parseAlarmActivations([{ entity_id: "CH-1", alarms: [alarm] }])).toEqual([
      { atNs: 140e9, entity: "CH-1", state: "temp_alarm", priority: "high" },
      { atNs: 530e9, entity: "CH-1", state: "temp_alarm", priority: "high" },
    ]);
  });

  it("does not invent instants from counts or invalid values", () => {
    expect(parseAlarmActivations(null)).toEqual([]);
    expect(parseAlarmActivations([null, { alarms: [null, { activations: 3 }] }])).toEqual([]);
    expect(parseAlarmActivations([{ alarms: [{ activation_times_ns: ["1", null, Infinity, 2] }] }])).toEqual([
      { atNs: 2, entity: "", state: "", priority: "" },
    ]);
  });
});
