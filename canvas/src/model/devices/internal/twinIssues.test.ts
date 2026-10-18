import { describe, expect, it } from "vitest";
import { cooler, coolerPath, sensor, startingChoices } from "../../../test/twinFixture";
import { changeDrive, chooseInstrument, chooseResponse } from "./twinEdits";
import { keepCurrentIssues, resolveTwinIssues, twinIssueTitle, twinIssuesOn } from "./twinIssues";

const members = [cooler, sensor];
const choices = startingChoices(members);

function place(pointer: string, extra: { keyword?: string; member?: string } = {}) {
  const [issue] = resolveTwinIssues(choices, [{ pointer, message: "raw words", ...extra }], members);
  if (issue === undefined) throw new Error("no issue");
  return issue;
}

describe("placing the binary's issues", () => {
  it.each([
    ["/id", { scope: "twin", field: "id" }],
    ["/title", { scope: "twin", field: "title" }],
    ["/kinds", { scope: "twin", field: "kinds" }],
    ["/kinds/0/kind", { scope: "kind", kind: coolerPath, field: "kind" }],
    ["/kinds/1/instance", { scope: "kind", field: "instance" }],
    ["/kinds/0/metrics", { scope: "kind", field: "metrics" }],
    ["/kinds/0/metrics/temperature", { scope: "metric", kind: coolerPath, key: "m1", field: "" }],
    ["/kinds/0/metrics/temperature/instrument", { scope: "metric", key: "m1", field: "instrument" }],
    ["/kinds/0/metrics/duty/initial", { scope: "metric", key: "m2", field: "initial" }],
    ["/kinds/0/commands/set_duty", { scope: "command", key: "c1", field: "", spot: "" }],
    ["/kinds/0/commands/set_duty/response", { scope: "command", field: "response", spot: "" }],
    ["/kinds/0/commands/set_duty/sets/duty", { scope: "command", field: "sets", spot: "duty" }],
    ["/kinds/0/commands/set_duty/drives", { scope: "command", field: "drives", spot: "" }],
    ["/kinds/0/commands/set_duty/drives/2", { scope: "command", field: "drives", spot: "2" }],
    ["/kinds/0/commands/set_duty/drives/0/from", { scope: "command", field: "from", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/0/kind", { scope: "command", field: "follows", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/0/metric", { scope: "command", field: "follows", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/1/gain", { scope: "command", field: "gain", spot: "1" }],
    ["/kinds/0/commands/set_duty/drives/0/time_constant_s", { scope: "command", field: "time_constant_s", spot: "0" }],
    ["/kinds/0/other", { scope: "kind", kind: coolerPath, field: "other" }],
    ["", { scope: "elsewhere" }],
    ["/kinds/9/kind", { scope: "elsewhere" }],
    ["/kinds/x/kind", { scope: "elsewhere" }],
  ])("%s", (pointer, expected) => {
    expect(place(pointer)).toMatchObject(expected);
  });

  it("names a reading that is not in the device with no key, and unescapes pointer parts", () => {
    expect(place("/kinds/0/metrics/a~1b~0c/unit")).toMatchObject({ scope: "metric", key: "", field: "unit" });
  });

  it("says each issue in plain words and keeps the binary's own words aside", () => {
    const issue = place("/kinds/0/commands/set_duty/drives/0/time_constant_s", { keyword: "exclusiveMinimum" });
    expect(issue).toMatchObject({ text: "Must be above zero", message: "raw words", keyword: "exclusiveMinimum" });
    expect(issue.pointer).toBe("/kinds/0/commands/set_duty/drives/0/time_constant_s");
    expect(
      place("/kinds/0/metrics/temperature/instrument", { keyword: "twin_conversion_needed", member: "unit" }).text,
    ).toBe("This instrument measures a different unit");
    expect(place("/kinds/0/kind", { keyword: "twin_member_required", member: "resolution" }).text).toBe(
      "Needs a resolution",
    );
    expect(place("/id").text).toBe("The binary does not accept this");
  });

  it("finds the messages on one field", () => {
    const issues = resolveTwinIssues(
      choices,
      [
        { pointer: "/kinds/0/commands/set_duty/drives/0/gain", message: "zero", keyword: "minimum" },
        { pointer: "/kinds/0/commands/set_duty/drives/1/gain", message: "other", keyword: "maximum" },
      ],
      members,
    );
    const where = { scope: "command", kind: coolerPath, key: "c1", field: "gain" } as const;
    expect(twinIssuesOn(issues, { ...where, spot: "0" })).toEqual(["Below the lowest allowed value"]);
    expect(twinIssuesOn(issues, { ...where, spot: "1" })).toEqual(["Above the highest allowed value"]);
    expect(twinIssuesOn(issues, where)).toEqual([]);
  });

  it("titles an issue by the device, the item and the field", () => {
    const title = (pointer: string) => twinIssueTitle(place(pointer), members);
    expect(title("/kinds/0/commands/set_duty/drives/0/time_constant_s")).toBe("cooler · set_duty · Time constant");
    expect(title("/id")).toBe("Twin name");
    expect(title("/kinds/0/metrics/duty")).toBe("cooler · duty");
    expect(title("")).toBe("Twin");
    expect(title("/kinds/0/other")).toBe("cooler");
    expect(title("/kinds/1/instance")).toBe("sensor · Name in the twin");
  });
});

describe("issues that no longer describe the choices", () => {
  const first = chooseResponse(choices, coolerPath, "c1", "first_order");
  const driven = changeDrive(first, { kind: coolerPath, command: "c1", index: 0 }, { timeConstantS: -5, gain: 1 });
  const issues = resolveTwinIssues(
    driven,
    [
      { pointer: "/kinds/0/commands/set_duty/drives/0/time_constant_s", message: "slow", keyword: "exclusiveMinimum" },
      { pointer: "/kinds/0/metrics/temperature/instrument", message: "unit", keyword: "twin_conversion_needed" },
      { pointer: "", message: "root" },
    ],
    members,
  );

  it("drops an issue when the value it points at has changed, and keeps the others", () => {
    const corrected = changeDrive(driven, { kind: coolerPath, command: "c1", index: 0 }, { timeConstantS: 30 });
    const kept = keepCurrentIssues(issues, driven, corrected, members);
    expect(kept.map((issue) => issue.field)).toEqual(["instrument", ""]);
    const otherwise = chooseInstrument(driven, coolerPath, "m1", "streamsim/thermistor@1");
    expect(keepCurrentIssues(issues, driven, otherwise, members).map((issue) => issue.field)).toEqual([
      "time_constant_s",
      "",
    ]);
  });

  it("keeps every issue while nothing it points at has changed", () => {
    expect(keepCurrentIssues(issues, driven, driven, members)).toEqual(issues);
    expect(keepCurrentIssues([], driven, first, members)).toEqual([]);
  });
});
