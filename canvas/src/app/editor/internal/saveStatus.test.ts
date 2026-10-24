import { describe, expect, it } from "vitest";
import type { Composition } from "../../../model/editor";
import type { Draft, SaveState } from "./editorDraft";
import { canSave, checked, isDirty, isSaving } from "./editorSelectors";
import { saveStatusOf } from "./saveStatus";

const nothing: Composition = { id: "lab" };
const doc: Composition = {
  id: "lab",
  equipment: [{ use: "lib/chiller@1", names: ["CH-1"], area: "Site" }],
};
const failure = { kind: "compose-refused" as const, diagnostic: { displayCode: "x", message: "too hot" } };
const issues = {
  instances: new Map(),
  listed: [{ target: "CH-1", message: "unknown", pointer: "/equipment/0/use" }],
};

function draft(change: Partial<Draft> = {}): Draft {
  return {
    phase: "ready",
    path: "compositions/lab.composition.json",
    domain: null,
    saved: doc,
    doc,
    save: { kind: "idle" },
    issues: null,
    past: [],
    future: [],
    specs: new Map(),
    library: [],
    selected: null,
    ...change,
  };
}

function savedAs(save: SaveState, change: Partial<Draft> = {}): Draft {
  return draft({ save, ...change });
}

describe("save status", () => {
  it("says what is happening while a save is under way or has failed to write", () => {
    expect(saveStatusOf(savedAs({ kind: "saving" }))).toEqual({
      text: "Saving and composing…",
      detail: null,
      tone: "busy",
    });
    expect(saveStatusOf(savedAs({ kind: "write-failed", failure }))).toMatchObject({
      text: "Not saved: too hot",
      tone: "warn",
    });
  });

  it("puts unsaved changes before the result of an older save", () => {
    const edited = { ...doc, title: "Mine" };
    const composed: SaveState = {
      kind: "composed",
      summary: "Composed.",
      doc,
      digest: "",
      plausibility: { kind: "running" },
    };
    expect(saveStatusOf(savedAs(composed, { doc: edited }))).toMatchObject({ text: "Unsaved changes", tone: "busy" });
    expect(saveStatusOf(savedAs(composed))).toEqual({ text: "Ready to run. Saved.", detail: "Composed.", tone: "ok" });
  });

  it("lists the issues in Build, or gives the refusal when there are none", () => {
    expect(saveStatusOf(savedAs({ kind: "compose-failed", failure }, { issues }))).toMatchObject({
      text: "Needs fixing: 1 issue, listed in Build. Your changes are saved.",
      tone: "warn",
    });
    expect(saveStatusOf(savedAs({ kind: "compose-failed", failure }))).toMatchObject({
      text: "Saved, but composition failed: too hot",
      tone: "warn",
    });
  });

  it("separates a checked draft from one never checked and one with nothing to save", () => {
    expect(saveStatusOf(draft({ domain: "domains/lab.domain.json" }))).toMatchObject({ text: "No changes" });
    expect(saveStatusOf(draft())).toMatchObject({ text: "Not checked yet: Save to see its issues." });
    const empty = draft({ doc: nothing, saved: nothing });
    expect(saveStatusOf(empty)).toMatchObject({ text: "Nothing to save yet: add equipment in Build." });
  });
});

describe("save selectors", () => {
  it("counts a draft as checked only when its domain is known or its last compose was of this document", () => {
    expect(checked(draft())).toBe(false);
    expect(checked(draft({ domain: "domains/lab.domain.json" }))).toBe(true);
    expect(
      checked(savedAs({ kind: "composed", summary: "", doc, digest: "", plausibility: { kind: "running" } })),
    ).toBe(true);
    expect(
      checked(
        savedAs({ kind: "composed", summary: "", doc: { ...doc }, digest: "", plausibility: { kind: "running" } }),
      ),
    ).toBe(false);
    expect(checked(savedAs({ kind: "saving" }, { domain: "domains/lab.domain.json" }))).toBe(false);
    expect(checked(savedAs({ kind: "compose-failed", failure }, { domain: "domains/lab.domain.json" }))).toBe(false);
  });

  it("allows Save for a failed save, an unchecked plant or a change, and not for a checked or empty one", () => {
    expect(canSave({ phase: "idle" })).toBe(false);
    expect(canSave(draft())).toBe(true);
    expect(canSave(draft({ domain: "domains/lab.domain.json" }))).toBe(false);
    expect(canSave(savedAs({ kind: "write-failed", failure }, { domain: "domains/lab.domain.json" }))).toBe(true);
    expect(canSave(savedAs({ kind: "compose-failed", failure }, { domain: "domains/lab.domain.json" }))).toBe(true);
    expect(canSave(draft({ doc: nothing, saved: nothing }))).toBe(false);
  });

  it("tells a saving, a dirty and a clean draft apart", () => {
    expect(isSaving(savedAs({ kind: "saving" }))).toBe(true);
    expect([isSaving(draft()), isSaving({ phase: "idle" })]).toEqual([false, false]);
    expect([isDirty(draft({ doc: nothing })), isDirty(draft()), isDirty({ phase: "idle" })]).toEqual([
      true,
      false,
      false,
    ]);
  });
});
