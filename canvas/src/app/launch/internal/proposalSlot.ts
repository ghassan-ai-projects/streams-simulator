import type { LaunchPlan } from "../../../model/launch";

export interface Proposal {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
  readonly from: string;
}

export class ProposalSlot {
  private held: (Proposal & { readonly domain: string }) | null = null;

  propose(domain: string, plan: LaunchPlan, left: readonly string[], from: string): void {
    this.held = { domain, plan, left, from };
  }

  forget(): void {
    this.held = null;
  }

  proposalFor(domain: string): Proposal | null {
    return this.held?.domain === domain ? { plan: this.held.plan, left: this.held.left, from: this.held.from } : null;
  }
}
