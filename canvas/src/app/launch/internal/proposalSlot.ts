import type { LaunchPlan } from "../../../model/launch";

export interface Proposal {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
  readonly from: string;
  readonly fresh: boolean;
}

export class ProposalSlot {
  private held: (Proposal & { readonly domain: string }) | null = null;

  propose(domain: string, proposal: Proposal): void {
    this.held = { domain, ...proposal };
  }

  forget(): void {
    this.held = null;
  }

  proposalFor(domain: string): Proposal | null {
    if (this.held?.domain !== domain) return null;
    const { plan, left, from, fresh } = this.held;
    return { plan, left, from, fresh };
  }
}
