import { captureContent } from "../../../model/devices";
import type { CapturedKind, CaptureSource, DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import { Observable, createOperationIdentity } from "../../shared";
import type { CaptureSession } from "./api";
import { canSkip, canWrite, reviewedOf, savedEtag } from "./captureSelectors";
import type { Way } from "./captureSelectors";
import { reviewed, withDecision } from "./captureState";
import type { CaptureState, Card, Decision, SavedFile } from "./captureState";
import { conflictWords } from "./deviceFailures";

function decisionOf(outcome: Outcome<string, DeviceFailure>): Decision {
  return outcome.ok ? { kind: "saved", path: outcome.value } : { kind: "failed", failure: outcome.failure };
}

class CaptureStore extends Observable<CaptureState> implements CaptureSession {
  private readonly operations = createOperationIdentity();
  private last: CaptureSource | null = null;
  private returning = false;

  constructor(private readonly gateway: DevicesGateway) {
    super({ phase: "idle" });
  }

  async read(source: CaptureSource): Promise<void> {
    if (this.snapshot().phase === "saving") return;
    this.last = source;
    const operation = this.operations.begin();
    this.set({ phase: "reading", source });
    const result = await this.gateway.captureKinds(source);
    if (!operation.live()) return;
    if (!result.ok) {
      this.set({ phase: "refused", source, failure: result.failure });
      return;
    }
    const saved = await Promise.all(result.value.kinds.map((kind) => this.savedOf(kind)));
    if (operation.live()) this.set(reviewed(source, result.value, saved));
  }

  async retry(): Promise<void> {
    if (this.last !== null) await this.read(this.last);
  }

  clear(): void {
    this.operations.invalidate();
    this.last = null;
    this.set({ phase: "idle" });
  }

  leave(): void {
    this.returning = true;
  }

  comingBack(): boolean {
    const was = this.returning;
    this.returning = false;
    return was;
  }

  save(index: number): Promise<void> {
    return this.write(index, "create");
  }

  replace(index: number): Promise<void> {
    return this.write(index, "replace");
  }

  skip(index: number): void {
    const card = this.cardAt(index);
    if (card !== undefined && canSkip(card)) this.decide(index, { kind: "skipped" });
  }

  unskip(index: number): void {
    if (this.cardAt(index)?.decision.kind === "skipped") this.decide(index, { kind: "open" });
  }

  private savedOf({ comparison }: CapturedKind): Promise<SavedFile> {
    return comparison.status === "differs" ? this.gateway.readFile(comparison.target) : Promise.resolve(null);
  }

  private cardAt(index: number): Card | undefined {
    return reviewedOf(this.snapshot())?.cards[index];
  }

  private async write(index: number, way: Way): Promise<void> {
    const card = this.cardAt(index);
    const operation = this.operations.current();
    if (card === undefined || operation === null || !canWrite(card, way)) return;
    this.decide(index, { kind: "saving" });
    const outcome = await this.put(card, way);
    if (operation.live()) this.decide(index, decisionOf(outcome));
  }

  private async put(card: Card, way: Way): Promise<Outcome<string, DeviceFailure>> {
    const { kind } = card;
    const etag = way === "create" ? "" : (savedEtag(card) ?? "");
    const written = await this.gateway.writeFile(kind.comparison.target, captureContent(kind.draft), etag);
    return written.ok ? success(written.value.path) : failure(conflictWords(written.failure, way, kind.kind));
  }

  private decide(index: number, decision: Decision): void {
    const board = reviewedOf(this.snapshot());
    if (board !== null) this.set(withDecision(board, index, decision));
  }
}

export function createCaptureSession(gateway: DevicesGateway): CaptureSession {
  return new CaptureStore(gateway);
}
