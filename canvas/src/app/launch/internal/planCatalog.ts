import type { FaultCategories, PerturbationKind } from "../../../model/launch";
import type { LaunchFailure, LaunchGateway, Outcome } from "../../ports";
import { Observable } from "../../shared";

export type Loaded<T> =
  | { readonly status: "loading" }
  | { readonly status: "ready"; readonly value: T }
  | { readonly status: "unavailable"; readonly words: string };

export interface PlanCatalogState {
  readonly perturbations: Loaded<readonly PerturbationKind[]>;
  readonly faultCategories: Loaded<FaultCategories>;
}

type CatalogGateway = Pick<LaunchGateway, "perturbationCatalog" | "faultCategories">;

const loading: PlanCatalogState = { perturbations: { status: "loading" }, faultCategories: { status: "loading" } };

function loaded<T>(outcome: Outcome<T, LaunchFailure>): Loaded<T> {
  return outcome.ok
    ? { status: "ready", value: outcome.value }
    : { status: "unavailable", words: outcome.failure.diagnostic.message };
}

export class PlanCatalog extends Observable<PlanCatalogState> {
  private domain = "";
  private asking = false;
  private readonly known = new Map<string, FaultCategories>();
  private readonly pending = new Set<string>();

  constructor(private readonly gateway: CatalogGateway) {
    super(loading);
  }

  load(domain: string): void {
    this.domain = domain;
    this.showCategories(domain);
    if (this.snapshot().perturbations.status !== "ready") void this.loadKinds();
    if (!this.known.has(domain) && !this.pending.has(domain)) void this.loadCategories(domain);
  }

  private showCategories(domain: string): void {
    const known = this.known.get(domain);
    const faultCategories: Loaded<FaultCategories> =
      known === undefined ? { status: "loading" } : { status: "ready", value: known };
    this.set({ ...this.snapshot(), faultCategories });
  }

  private async loadKinds(): Promise<void> {
    if (this.asking) return;
    this.asking = true;
    const answer = loaded(await this.gateway.perturbationCatalog());
    this.asking = false;
    this.set({ ...this.snapshot(), perturbations: answer });
  }

  private async loadCategories(domain: string): Promise<void> {
    this.pending.add(domain);
    const answer = await this.gateway.faultCategories(domain);
    this.pending.delete(domain);
    if (answer.ok) this.known.set(domain, answer.value);
    if (this.domain === domain) this.set({ ...this.snapshot(), faultCategories: loaded(answer) });
  }
}
