import { describe, expect, it } from "vitest";
import { coolerDocument, instruments, kindOf, sensorDocument } from "../../../test/twinFixture";
import { startingChoices } from "./twinDefaults";
import { interfaceLabel, interfaceNeeds, resolveTwinIssues, twinIssueTitle, twinIssuesOn } from "./twinIssues";

const kinds = [kindOf(coolerDocument), kindOf(sensorDocument)];
const choices = startingChoices(kinds, instruments);

function place(pointer: string, message = "refused") {
  const [issue] = resolveTwinIssues(choices, [{ pointer, message }], kinds);
  if (issue === undefined) throw new Error("no issue");
  return issue;
}

describe("placing the binary's issues", () => {
  it.each([
    ["/id", { scope: "twin", field: "id" }],
    ["/title", { scope: "twin", field: "title" }],
    ["/kinds", { scope: "twin", field: "kinds" }],
    ["/kinds/0/kind", { scope: "kind", kind: "cooler", field: "kind" }],
    ["/kinds/1/instance", { scope: "kind", kind: "sensor", field: "instance" }],
    ["/kinds/0/metrics", { scope: "kind", kind: "cooler", field: "metrics" }],
    ["/kinds/0/metrics/temperature", { scope: "metric", kind: "cooler", key: "temperature", field: "" }],
    ["/kinds/0/metrics/temperature/instrument", { scope: "metric", key: "temperature", field: "instrument" }],
    ["/kinds/0/metrics/duty/initial", { scope: "metric", key: "duty", field: "initial" }],
    ["/kinds/0/commands/set_duty", { scope: "command", key: "set_duty", field: "", spot: "" }],
    ["/kinds/0/commands/set_duty/response", { scope: "command", field: "response", spot: "" }],
    ["/kinds/0/commands/set_duty/sets/duty", { scope: "command", field: "sets", spot: "duty" }],
    ["/kinds/0/commands/set_duty/drives", { scope: "command", field: "drives", spot: "" }],
    ["/kinds/0/commands/set_duty/drives/2", { scope: "command", field: "drives", spot: "2" }],
    ["/kinds/0/commands/set_duty/drives/0/from", { scope: "command", field: "from", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/0/kind", { scope: "command", field: "follows", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/0/metric", { scope: "command", field: "follows", spot: "0" }],
    ["/kinds/0/commands/set_duty/drives/1/gain", { scope: "command", field: "gain", spot: "1" }],
    ["/kinds/0/commands/set_duty/drives/0/time_constant_s", { scope: "command", field: "time_constant_s", spot: "0" }],
    ["/kinds/0/other", { scope: "kind", kind: "cooler", field: "other" }],
    ["", { scope: "elsewhere" }],
    ["/kinds/9/kind", { scope: "elsewhere" }],
    ["/kinds/x/kind", { scope: "elsewhere" }],
  ])("%s", (pointer, expected) => {
    expect(place(pointer)).toMatchObject(expected);
  });

  it("unescapes pointer parts", () => {
    expect(place("/kinds/0/metrics/a~1b~0c/unit")).toMatchObject({ scope: "metric", key: "a/b~c", field: "unit" });
  });

  it("finds the messages on one field", () => {
    const issues = resolveTwinIssues(
      choices,
      [
        { pointer: "/kinds/0/commands/set_duty/drives/0/gain", message: "zero" },
        { pointer: "/kinds/0/commands/set_duty/drives/1/gain", message: "other" },
      ],
      kinds,
    );
    const where = { scope: "command", kind: "cooler", key: "set_duty", field: "gain" } as const;
    expect(twinIssuesOn(issues, { ...where, spot: "0" })).toEqual(["zero"]);
    expect(twinIssuesOn(issues, { ...where, spot: "1" })).toEqual(["other"]);
    expect(twinIssuesOn(issues, where)).toEqual([]);
  });

  it("titles an issue by the device, the item and the field", () => {
    expect(twinIssueTitle(place("/kinds/0/commands/set_duty/drives/0/time_constant_s"))).toBe(
      "cooler · set_duty · Time constant",
    );
    expect(twinIssueTitle(place("/id"))).toBe("Twin name");
    expect(twinIssueTitle(place("/kinds/0/metrics/duty"))).toBe("cooler · duty");
    expect(twinIssueTitle(place(""))).toBe("Twin");
    expect(twinIssueTitle(place("/kinds/0/other"))).toBe("cooler");
  });
});

const refusal = (index: number, field = "resolution", section = "metrics") =>
  `kind file "cooler": component: the kind file holds what a twin cannot reproduce: /${section}/${String(index)} needs "${field}": the interface a twin reproduces always states it`;

describe("a kind file that lacks what a twin needs", () => {
  it("is placed on the reading it names, in plain words", () => {
    const issue = place("/kinds/0/kind", refusal(0));
    expect(issue).toEqual({
      scope: "metric",
      kind: "cooler",
      key: "temperature",
      field: "resolution",
      spot: "",
      message: "Needs a resolution.",
    });
    expect(twinIssueTitle(issue)).toBe("cooler · temperature · Resolution");
    expect(place("/kinds/0/kind", refusal(1, "scan_class"))).toMatchObject({
      key: "duty",
      message: "Needs a scan class.",
    });
    expect(place("/kinds/0/kind", refusal(0, "writable"))).toMatchObject({
      field: "writable",
      message: "Needs a writable.",
    });
  });

  it("is placed on the command it names", () => {
    const issue = place("/kinds/0/kind", refusal(0, "safe_stop", "commands"));
    expect(issue).toMatchObject({ scope: "command", kind: "cooler", key: "set_duty", field: "safe_stop" });
    expect(issue.message).toBe("Needs a safe stop.");
    expect(interfaceNeeds([issue], "command", "cooler", "set_duty")).toEqual([issue]);
    expect(interfaceNeeds([issue], "metric", "cooler", "set_duty")).toEqual([]);
  });

  it("stays on the device when the reading cannot be found or the message says something else", () => {
    expect(place("/kinds/0/kind", refusal(9))).toMatchObject({ scope: "kind", field: "kind" });
    expect(place("/kinds/0/kind", "no kind file defines it")).toMatchObject({ scope: "kind", field: "kind" });
    expect(place("/kinds/0/instance", refusal(0))).toMatchObject({ scope: "kind", field: "instance" });
  });

  it("lists what a reading needs from its interface, and words the field for the interface form", () => {
    const issues = resolveTwinIssues(choices, [{ pointer: "/kinds/0/kind", message: refusal(0) }], kinds);
    expect(interfaceNeeds(issues, "metric", "cooler", "temperature")).toEqual(issues);
    expect(interfaceNeeds(issues, "metric", "cooler", "duty")).toEqual([]);
    const choiceIssue = place("/kinds/0/metrics/temperature/instrument");
    expect(interfaceNeeds([choiceIssue], "metric", "cooler", "temperature")).toEqual([]);
    expect(interfaceLabel("resolution")).toBe("Resolution");
    expect(interfaceLabel("scan_class")).toBe("Scan class");
    expect(interfaceLabel("writable")).toBe("Setpoint");
    expect(interfaceLabel("safe_stop")).toBe("Safe stop");
  });
});
