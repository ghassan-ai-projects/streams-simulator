import { Observable } from "../../shared";
import type { ConnectGateway, Outcome, PolicyDefaultFailure } from "../../ports";

export interface PolicyDefaultState {
  readonly file: "unread" | "absent" | "present" | "unreadable";
  readonly policies: readonly string[];
  readonly save: "idle" | "saving" | "saved" | "conflict" | "failed";
  readonly problem: string;
}

const unread: PolicyDefaultState = { file: "unread", policies: [], save: "idle", problem: "" };

const unreadableWords = "The workspace default could not be read, so it was not changed.";

type Saved = Outcome<{ readonly etag: string }, PolicyDefaultFailure>;

export class PolicyDefault extends Observable<PolicyDefaultState> {
  private etag: string | null = null;

  constructor(private readonly gateway: Pick<ConnectGateway, "readPolicyDefault" | "savePolicyDefault">) {
    super(unread);
  }

  async load(): Promise<readonly string[] | null> {
    this.set({ ...this.snapshot(), save: "idle", problem: "" });
    return this.read();
  }

  forgetSave(): void {
    const { save } = this.snapshot();
    if (save !== "idle" && save !== "saving") this.set({ ...this.snapshot(), save: "idle", problem: "" });
  }

  async save(policies: readonly string[]): Promise<void> {
    if (this.snapshot().save === "saving") return;
    this.set({ ...this.snapshot(), save: "saving", problem: "" });
    if (this.etag === null) await this.read();
    const etag = this.etag;
    if (etag === null) this.set({ ...this.snapshot(), save: "failed", problem: unreadableWords });
    else await this.finish(policies, await this.gateway.savePolicyDefault(policies, etag));
  }

  private async read(): Promise<readonly string[] | null> {
    const outcome = await this.gateway.readPolicyDefault();
    if (!outcome.ok) {
      this.etag = null;
      this.set({ ...this.snapshot(), file: "unreadable", policies: [] });
      return null;
    }
    const file = outcome.value;
    const policies = file.state === "present" ? file.policies : [];
    this.etag = file.state === "absent" ? "" : file.etag;
    this.set({ ...this.snapshot(), file: file.state, policies });
    return file.state === "present" ? policies : null;
  }

  private async finish(policies: readonly string[], outcome: Saved): Promise<void> {
    if (outcome.ok) {
      this.etag = outcome.value.etag;
      this.set({ file: "present", policies, save: "saved", problem: "" });
      return;
    }
    const { kind, diagnostic } = outcome.failure;
    if (kind === "conflict") await this.read();
    this.set({ ...this.snapshot(), save: kind === "conflict" ? "conflict" : "failed", problem: diagnostic.message });
  }
}
