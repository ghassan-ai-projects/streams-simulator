import { deferred } from "./deferred";
import type { Canned } from "./fakeFailures";
import { FakeBridge } from "./fakeBridge";
import { advanced, launchBridge } from "./launchFixture";

export class HeldBridge extends FakeBridge {
  private held: (() => void) | null = null;
  private freed = false;

  constructor() {
    super();
    launchBridge().actions.forEach((outcome, name) => this.actions.set(name, outcome));
    this.handlers.set("clock.advance", () => this.advance());
  }

  release(): void {
    this.held?.();
  }

  free(): void {
    this.freed = true;
    this.held?.();
  }

  private advance(): Promise<Canned<unknown>> | Canned<unknown> {
    if (this.freed) return advanced;
    const gate = deferred<Canned<unknown>>();
    this.held = () => {
      gate.resolve(advanced);
    };
    return gate.promise;
  }
}
