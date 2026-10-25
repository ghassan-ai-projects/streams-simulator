import { describe, expect, it } from "vitest";
import { ProposalSlot } from "./proposalSlot";

const plan = { seed: 1, durationS: 600, explore: true, faults: [], perturbations: [], calls: [] };

describe("proposal slot", () => {
  it("offers a proposal only to the plant it was made for, until forgotten", () => {
    const slot = new ProposalSlot();
    expect(slot.proposalFor("factory")).toBeNull();
    slot.propose("factory", { plan, left: ["a call"], from: "run w-1", fresh: false });
    expect(slot.proposalFor("factory")).toEqual({ plan, left: ["a call"], from: "run w-1", fresh: false });
    expect(slot.proposalFor("other")).toBeNull();
    slot.forget();
    expect(slot.proposalFor("factory")).toBeNull();
  });

  it("keeps only the latest proposal", () => {
    const slot = new ProposalSlot();
    slot.propose("factory", { plan, left: [], from: "", fresh: false });
    slot.propose("other", { plan: { ...plan, seed: 2 }, left: [], from: "", fresh: true });
    expect(slot.proposalFor("factory")).toBeNull();
    expect(slot.proposalFor("other")).toMatchObject({ plan: { seed: 2 }, fresh: true });
  });
});
