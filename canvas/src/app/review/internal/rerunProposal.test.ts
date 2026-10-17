import { describe, expect, it } from "vitest";
import { reviewState } from "../../../test/reviewFixture";
import { rerunProposalOf } from "./rerunProposal";

describe("a rerun proposal", () => {
  it("takes the run's plan and what it cannot repeat, for the plant the run used", () => {
    const proposal = rerunProposalOf(reviewState({}));
    expect(proposal).toMatchObject({ domainId: "mini", plan: { seed: 7, explore: false } });
    expect(proposal?.left.length).toBeGreaterThan(0);
  });

  it("is none before the report is ready", () => {
    expect(rerunProposalOf(reviewState({ report: { phase: "loading" } }))).toBeNull();
  });
});
