import { describe, expect, it } from "vitest";
import { report } from "../../../test/runFixture";
import { reviewState } from "../../../test/reviewFixture";
import { rerunProposalOf } from "./rerunProposal";

const reportWith = (overrides: Partial<typeof report>) => ({ ...report, ...overrides });

describe("a rerun proposal", () => {
  it("takes the run's plan and what it cannot repeat, for the plant the run used", () => {
    const proposal = rerunProposalOf(reviewState({}));
    expect(proposal).toMatchObject({ domainId: "mini", plan: { seed: 7, explore: false } });
    expect(proposal?.fresh).toBe(false);
    expect(proposal?.plan.calls).toHaveLength(4);
    expect(proposal?.plan.calls[0]).toEqual({
      entity: "CH-1",
      effector: "start",
      atS: 1200,
      args: { chiller_id: "CH-1", level: 0.5 },
    });
    expect(proposal?.left.length).toBeGreaterThan(0);
  });

  it("offers fresh settings for a run that did not finish, in the mode it opened with", () => {
    const open = reportWith({ ended: false, complete: false, mode: "explore" });
    expect(rerunProposalOf(reviewState({ report: { phase: "ready", value: open } }))).toMatchObject({
      domainId: "mini",
      fresh: true,
      left: [],
      plan: { explore: true, faults: [], calls: [] },
    });
  });

  it("is none before the report is ready", () => {
    expect(rerunProposalOf(reviewState({ report: { phase: "loading" } }))).toBeNull();
  });
});
