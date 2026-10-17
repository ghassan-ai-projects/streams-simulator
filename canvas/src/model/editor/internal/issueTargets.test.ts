import { describe, expect, it } from "vitest";
import type { Composition } from "./composition";
import { groupIssues, issueTargets, profileWeightOf, removeProfileWeight } from "./issueTargets";

const doc: Composition = {
  equipment: [
    { use: "streamsim/chiller@1", names: ["CH-1", "CH-2"], area: "Cooling" },
    { use: "streamsim/cooling-tower@1", names: ["CT-1"], area: "Cooling" },
  ],
  connections: [
    { from: "CH-1.heat_out", to: "CT-1.heat_in" },
    { from: "CT-1.basin_out", to: "CH-2.cond_in" },
  ],
};

describe("issue targets", () => {
  it("puts each issue of the binary on the instances its pointer names", () => {
    const targets = issueTargets(doc, [
      { pointer: "/equipment/0/area", message: "bad area" },
      { pointer: "/connections/1/from", message: "no port" },
      { pointer: "/equipment", message: "too few" },
      { pointer: "/equipment/0/names/1", message: "bad name" },
      { pointer: "/equipment/0/names/7", message: "gone" },
    ]);
    expect(targets.instances.get("CH-1")).toEqual(["bad area"]);
    expect(targets.instances.get("CH-2")).toEqual(["bad area", "no port", "bad name"]);
    expect(targets.instances.get("CT-1")).toEqual(["no port"]);
    expect(targets.listed.map((entry) => entry.target)).toEqual(["CH-1", "CT-1", null, "CH-2", null]);
  });

  it("counts repeated issues on one target once, keeping the latest pointer", () => {
    const groups = groupIssues([
      { target: "CH-1", message: "bad", pointer: "/equipment/0/a" },
      { target: null, message: "bad", pointer: "/x" },
      { target: "CH-1", message: "bad", pointer: "/equipment/0/b" },
      { target: null, message: "bad", pointer: "/y" },
    ]);
    expect(groups.map((group) => [group.target, group.pointer, group.count])).toEqual([
      ["CH-1", "/equipment/0/b", 2],
      [null, "/x", 1],
      [null, "/y", 1],
    ]);
  });

  it("reads the profile and fault of an undeclared weight", () => {
    const issue = { target: null, message: 'profile "p" weights undeclared fault "f"', pointer: "" };
    expect(profileWeightOf(issue)).toEqual({ profile: "p", fault: "f" });
    expect(profileWeightOf({ ...issue, message: "other" })).toBeNull();
  });
});

describe("profile weights a plant no longer declares", () => {
  it("removes the weight the binary's pointer names and keeps the others", () => {
    const profiled: Composition = {
      ...doc,
      profiles: [{ name: "p", fault_weights: { "a/b__c": 1, keep: 2 } }],
    };
    expect(removeProfileWeight(profiled, "/profiles/0/fault_weights/a~1b__c").profiles).toEqual([
      { name: "p", fault_weights: { keep: 2 } },
    ]);
    expect(removeProfileWeight(profiled, "/profiles/3/fault_weights/keep")).toBe(profiled);
  });
});
