import { describe, expect, it } from "vitest";
import {
  filledFromDefault,
  policyHint,
  policyListWords,
  policyNotes,
  policyNoteWords,
  prefilledWords,
  recordedPolicyWords,
  saveDefaultHint,
  saveDefaultWords,
  unreadableDefaultWords,
  withPolicyListed,
  withPolicyText,
} from "./policies";
import { firstSessionDraft, policyLines, sessionPlanText } from "./session";

const a = `sha256:${"a".repeat(64)}`;
const b = `sha256:${"b".repeat(64)}`;

describe("the allowed policy digests of a session plan", () => {
  it("writes the non-empty trimmed lines in the operator's order, and no member when there are none", () => {
    const plan = (policies: string): Record<string, unknown> =>
      JSON.parse(sessionPlanText("d", { ...firstSessionDraft, policies })) as Record<string, unknown>;
    expect(plan(` ${b} \r\n\n  ${a}\n`).policies).toEqual([b, a]);
    expect(plan("").policies).toBeUndefined();
    expect(plan(" \n \n").policies).toBeUndefined();
    expect(plan(`${a}\n${a}`).policies).toEqual([a, a]);
    expect(policyLines("x\n\n y ")).toEqual(["x", "y"]);
  });

  it("puts an issue on the digest it names by its position among the listed lines, and one on the list itself", () => {
    const notes = policyNotes([
      { pointer: "/policies/1", message: "m", keyword: "policy_listed_twice" },
      { pointer: "/policies", message: "m", keyword: "maxItems" },
      { pointer: "/policies/x", message: "m" },
      { pointer: "/host_socket", message: "m" },
      { pointer: "/policiesX", message: "m" },
    ]);
    expect(notes.map((note) => note.digest)).toEqual([2, null]);
  });

  it("words each rule of the binary plainly and keeps the binary's text for a rule it does not know", () => {
    const words = (keyword: string | undefined, pointer = "/policies/0"): string => {
      const issue = { pointer, message: "binary text", ...(keyword === undefined ? {} : { keyword }) };
      const [note] = policyNotes([issue]);
      return note === undefined ? "" : policyNoteWords(note);
    };
    expect(words("pattern")).toBe("Digest 1: A policy digest is sha256: and 64 lower-case hex digits.");
    expect(words("policy_malformed")).toBe("Digest 1: A policy digest is sha256: and 64 lower-case hex digits.");
    expect(words("maxItems", "/policies")).toBe("List at most 8 policy digests.");
    expect(words("too_many_policies", "/policies")).toBe("List at most 8 policy digests.");
    expect(words("policy_listed_twice")).toBe("Digest 1: This digest is already listed.");
    expect(words("policy_reserved")).toBe("Digest 1: This digest is reserved for the node's own stops.");
    expect(words("minItems", "/policies")).toBe("binary text");
    expect(words(undefined)).toBe("Digest 1: binary text");
    expect(policyHint).toContain("safe stops always run");
  });

  it("says what is in effect: none and what that means, a count, or not recorded", () => {
    expect(policyListWords([])).toBe(
      "Allowed policies: none. Commands to simulated devices are refused, except safe stops.",
    );
    expect(policyListWords([a])).toBe("Allowed policies: 1");
    expect(policyListWords([a, b])).toBe("Allowed policies: 2");
    expect(recordedPolicyWords(null)).toBe("Allowed policies: not recorded.");
    expect(recordedPolicyWords([a, b])).toBe("Allowed policies: 2");
  });

  it("marks a typed list as the operator's and a filled one as the workspace default's, and lists a digest once", () => {
    const filled = filledFromDefault(firstSessionDraft, [a, b]);
    expect(filled).toMatchObject({ policies: `${a}\n${b}`, policiesFrom: "default" });
    expect(withPolicyText(filled, a)).toMatchObject({ policies: a, policiesFrom: "operator" });
    expect(withPolicyListed(firstSessionDraft, a)).toMatchObject({ policies: a, policiesFrom: "operator" });
    expect(withPolicyListed(withPolicyListed(firstSessionDraft, a), b).policies).toBe(`${a}\n${b}`);
    const kept = withPolicyListed(filled, a);
    expect(kept).toBe(filled);
    expect(kept.policiesFrom).toBe("default");
  });

  it("words the outcome of saving the workspace default, and says what it is for", () => {
    expect(saveDefaultWords("saving", "", [])).toBe("Saving the workspace default…");
    expect(saveDefaultWords("saved", "", [a])).toBe("Saved 1 policy digest as the workspace default.");
    expect(saveDefaultWords("saved", "", [a, b])).toBe("Saved 2 policy digests as the workspace default.");
    expect(saveDefaultWords("saved", "", [])).toBe("The workspace default is now empty.");
    expect(saveDefaultWords("conflict", "", [])).toContain("changed elsewhere, so nothing was saved");
    expect(saveDefaultWords("failed", "disk full", [])).toBe("The workspace default was not saved: disk full");
    expect(prefilledWords).toBe("Pre-filled from the workspace default.");
    expect(unreadableDefaultWords).toBe("The workspace default could not be read.");
    expect(saveDefaultHint).toContain("A running session never changes.");
  });
});
