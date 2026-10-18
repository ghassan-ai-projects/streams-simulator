import { captureContent } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { CaptureSource, DeviceFailure, DevicesGateway, Outcome } from "../../ports";
import { failure, success } from "../../ports";
import { Observable, createOperationIdentity } from "../../shared";
import type { CaptureSession } from "./api";
import { canWrite, reviewedOf } from "./captureSelectors";
import type { Way } from "./captureSelectors";
import { reviewed, withDecision } from "./captureState";
import type { CaptureState, Card, Decision } from "./captureState";
import { conflictWords } from "./deviceFailures";

function decisionOf(outcome: Outcome<string, DeviceFailure>): Decision {
  return outcome.ok ? { kind: "saved", path: outcome.value } : { kind: "failed", failure: outcome.failure };
}

class CaptureStore extends Observable<CaptureState> implements CaptureSession {
  private readonly operations = createOperationIdentity();
  private kinds: readonly KindEntry[] = [];
  private last: CaptureSource | null = null;

  constructor(private readonly gateway: DevicesGateway) {
    super({ phase: "idle" });
  }

  async read(source: CaptureSource, kinds: readonly KindEntry[]): Promise<void> {
    this.kinds = kinds;
    this.last = source;
    const operation = this.operations.begin();
    this.set({ phase: "reading", source });
    const result = await this.gateway.captureKinds(source);
    if (!operation.live()) return;
    this.set(result.ok ? reviewed(source, result.value, kinds) : { phase: "refused", source, failure: result.failure });
  }

  async retry(): Promise<void> {
    if (this.last !== null) await this.read(this.last, this.kinds);
  }

  close(): void {
    this.operations.invalidate();
    this.last = null;
    this.set({ phase: "idle" });
  }

  save(name: string): Promise<void> {
    return this.write(name, "create");
  }

  replace(name: string): Promise<void> {
    return this.write(name, "replace");
  }

  skip(name: string): void {
    const kind = reviewedOf(this.snapshot())?.cards.find((card) => card.kind.kind === name)?.decision.kind;
    if (kind === "open" || kind === "failed") this.decide(name, { kind: "skipped" });
  }

  private async write(name: string, way: Way): Promise<void> {
    const card = reviewedOf(this.snapshot())?.cards.find((entry) => entry.kind.kind === name);
    const operation = this.operations.current();
    if (card === undefined || operation === null || !canWrite(card, way)) return;
    this.decide(name, { kind: "saving" });
    const outcome = await this.put(card, way);
    if (operation.live()) this.decide(name, decisionOf(outcome));
  }

  private async put({ kind, path }: Card, way: Way): Promise<Outcome<string, DeviceFailure>> {
    const etag = way === "create" ? success("") : await this.etagOf(path);
    if (!etag.ok) return failure(etag.failure);
    const written = await this.gateway.writeFile(path, captureContent(kind.draft), etag.value);
    return written.ok ? success(written.value.path) : failure(conflictWords(written.failure, way, kind.kind));
  }

  private async etagOf(path: string): Promise<Outcome<string, DeviceFailure>> {
    const file = await this.gateway.readFile(path);
    if (file.ok) return success(file.value.etag);
    return file.failure.kind === "missing" ? success("") : failure(file.failure);
  }

  private decide(name: string, decision: Decision): void {
    const board = reviewedOf(this.snapshot());
    if (board !== null) this.set(withDecision(board, name, decision));
  }
}

export function createCaptureSession(gateway: DevicesGateway): CaptureSession {
  return new CaptureStore(gateway);
}
