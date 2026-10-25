import type { Lineage } from "../../../model/workspace";
import type { ProvenanceGateway } from "../../ports";
import { Observable } from "../../shared";

export class LineageStore extends Observable<Lineage | null> {
  private running = false;
  private again = false;

  constructor(private readonly gateway: ProvenanceGateway) {
    super(null);
  }

  async refresh(): Promise<void> {
    if (this.running) {
      this.again = true;
      return;
    }
    this.running = true;
    try {
      await this.pass();
    } finally {
      this.running = false;
    }
  }

  private async pass(): Promise<void> {
    this.again = false;
    const outcome = await this.gateway.lineage();
    this.set(outcome.ok ? outcome.value : null);
    if (this.wanted()) await this.pass();
  }

  private wanted(): boolean {
    return this.again;
  }
}
