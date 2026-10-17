import { describe, expect, it } from "vitest";
import { ProposalSlot } from "./proposalSlot";

const plan = { seed: 1, durationS: 600, explore: true, faults: [], calls: [] };

describe("proposal slot", () => {
  it("offers a proposal only to the plant it was made for, until forgotten", () => {
    const slot = new ProposalSlot();
    expect(slot.proposalFor("factory")).toBeNull();
    slot.propose("factory", plan, ["a call"], "run w-1");
    expect(slot.proposalFor("factory")).toEqual({ plan, left: ["a call"], from: "run w-1" });
    expect(slot.proposalFor("other")).toBeNull();
    slot.forget();
    expect(slot.proposalFor("factory")).toBeNull();
  });

  it("keeps only the latest proposal", () => {
    const slot = new ProposalSlot();
    slot.propose("factory", plan, [], "");
    slot.propose("other", { ...plan, seed: 2 }, [], "");
    expect(slot.proposalFor("factory")).toBeNull();
    expect(slot.proposalFor("other")?.plan.seed).toBe(2);
  });
});
