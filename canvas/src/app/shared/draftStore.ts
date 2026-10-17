import { Observable } from "./observable";

export abstract class DraftStore<S, D extends S & { readonly path: string }> extends Observable<S> {
  protected abstract loading(path: string): S;

  protected abstract toDraft(state: S): D | null;

  protected abstract restart(path: string): Promise<void>;

  async open(path: string): Promise<void> {
    this.set(this.loading(path));
    await this.restart(path);
  }

  async reload(): Promise<void> {
    const draft = this.draft();
    if (draft !== null) await this.restart(draft.path);
  }

  protected draft(): D | null {
    return this.toDraft(this.snapshot());
  }

  protected update(change: (draft: D) => D): void {
    const draft = this.draft();
    if (draft !== null) this.set(change(draft));
  }
}
