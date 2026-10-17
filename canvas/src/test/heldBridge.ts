import { success } from "../app/ports";
import type { Canned } from "./fakeFailures";
import { FakeBridge } from "./fakeBridge";
import { launchBridge } from "./launchFixture";

/** A bridge whose `clock.advance` stays unanswered until the test releases it. */
export class HeldBridge extends FakeBridge {
  private held: (() => void) | null = null;
  private open = false;

  constructor() {
    super();
    launchBridge().actions.forEach((outcome, name) => this.actions.set(name, outcome));
  }

  override action(name: string, args: Readonly<Record<string, unknown>>): Promise<Canned<unknown>> {
    if (name !== "clock.advance" || this.open) return super.action(name, args);
    this.calls.push({ name, args });
    return new Promise((resolve) => {
      this.held = () => {
        resolve(success({}));
      };
    });
  }

  release(): void {
    this.held?.();
  }

  free(): void {
    this.open = true;
    this.held?.();
  }
}
