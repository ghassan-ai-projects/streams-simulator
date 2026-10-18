import { describe, expect, it } from "vitest";
import { addMetric, blankKind, changeMetric, localIssue } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import type { Draft } from "./deviceDraft";
import { saveStatusOf } from "./saveStatus";

const kind = blankKind("fan");
const draft: Draft = {
  phase: "ready",
  path: "kinds/fan.kind.json",
  saved: kind,
  doc: kind,
  selected: { part: "kind" },
  save: { kind: "idle" },
  check: { kind: "unchecked" },
  review: { state: "idle", schema: [], needs: [] },
};
const failure = (kindName: DeviceFailure["kind"], message: string): DeviceFailure =>
  kindName === "refused"
    ? { kind: "refused", diagnostic: { message, displayCode: "c" }, issues: [] }
    : { kind: kindName, diagnostic: { message, displayCode: "c" } };

describe("save status", () => {
  it.each([
    [{ save: { kind: "saving" } }, "Saving and checking…", "busy"],
    [{ save: { kind: "write-failed", failure: failure("unavailable", "disk full") } }, "Not saved: disk full", "warn"],
    [
      { save: { kind: "write-failed", failure: failure("conflict", "changed") } },
      "Not saved: the file changed on disk. Reload it to continue.",
      "warn",
    ],
    [
      { save: { kind: "refused", failure: failure("refused", "bad") } },
      "Not saved: the binary refuses it (bad). Fix it to save.",
      "warn",
    ],
    [{ check: { kind: "unchecked" } }, "Not checked yet.", ""],
    [{ check: { kind: "checking" } }, "Checking with the binary…", "busy"],
    [
      { check: { kind: "valid", summary: { kind: "fan", digest: "d", metrics: 2, commands: 1 } } },
      "Saved and valid.",
      "ok",
    ],
    [
      { check: { kind: "refused", failure: failure("refused", "bad") } },
      "Saved, but the binary refuses it: bad.",
      "warn",
    ],
    [
      { check: { kind: "failed", failure: failure("unavailable", "down") } },
      "Saved, but it could not be checked: down",
      "warn",
    ],
  ] as const)("describes %j", (change, text, tone) => {
    const status = saveStatusOf({ ...draft, ...change });
    expect(status).toMatchObject({ text, tone });
  });

  it("gives the facts of a valid device in the detail, and counts the issues of a refused one", () => {
    const valid = saveStatusOf({
      ...draft,
      check: { kind: "valid", summary: { kind: "fan", digest: "d", metrics: 2, commands: 1 } },
    });
    expect(valid.detail).toBe("2 readings · 1 command");
    const refused = saveStatusOf({
      ...draft,
      check: { kind: "refused", failure: failure("refused", "bad") },
      review: { state: "done", schema: [localIssue("kind", "", "kind", "x")], needs: [] },
    });
    expect(refused.text).toBe("Saved, but the binary refuses it: 1 issue.");
    const blocked = saveStatusOf({
      ...draft,
      save: { kind: "refused", failure: failure("refused", "bad") },
      review: {
        state: "done",
        schema: [localIssue("kind", "", "kind", "x"), localIssue("kind", "", "kind", "y")],
        needs: [],
      },
    });
    expect(blocked.text).toBe("Not saved: the binary refuses it (2 issues). Fix it to save.");
  });

  it("asks for names when the unsaved device cannot be saved yet", () => {
    const edited = addMetric(kind).kind;
    expect(saveStatusOf({ ...draft, doc: edited }).text).toBe("Unsaved changes. Name everything to save.");
    const named = changeMetric(edited, "m1", { name: "duty" });
    expect(saveStatusOf({ ...draft, doc: named }).text).toBe("Unsaved changes");
  });
});
