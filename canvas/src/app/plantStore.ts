import type { BridgeFailure, FileContent, Outcome, Workspace } from "../model/bridge";
import type { Plant } from "../model/plant";
import { parsePlant } from "../model/plant";
import type { ComponentSpec } from "../model/ports";
import { componentSpecs } from "./componentSpecs";
import { Observable } from "./observable";
import { createOperationIdentity } from "./shared/operationIdentity";
import type { OperationIdentity, OperationId } from "./shared/operationIdentity";
import type { Bridge } from "./ports";

export type PlantState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: BridgeFailure }
  | {
      readonly phase: "ready";
      readonly path: string;
      readonly etag: string;
      readonly plant: Plant;
      readonly specs: ReadonlyMap<string, ComponentSpec>;
      readonly composition: string | null;
      readonly selected: string | null;
    };

function parsed(content: string): unknown {
  try {
    return JSON.parse(content) as unknown;
  } catch {
    return null;
  }
}

function pairOf(id: string, workspace: Workspace): string | null {
  return workspace.compositions.find((entry) => entry.id === id && id !== "")?.path ?? null;
}

function unchanged(
  held: Extract<PlantState, { phase: "ready" }> | null,
  file: Outcome<FileContent>,
  workspace: Workspace,
): boolean {
  return (
    held !== null && file.ok && file.value.etag === held.etag && pairOf(held.plant.id, workspace) === held.composition
  );
}

export class PlantStore extends Observable<PlantState> {
  constructor(
    private readonly bridge: Bridge,
    private readonly operations: OperationIdentity = createOperationIdentity(),
  ) {
    super({ phase: "idle" });
  }

  async load(path: string, workspace: Workspace): Promise<void> {
    if (this.loading(path)) return;
    const request = this.operations.begin();
    const held = this.held(path);
    if (held === null) this.set({ phase: "loading", path });
    const file = await this.bridge.readFile(path);
    if (!this.current(request) || unchanged(held, file, workspace)) return;
    if (!file.ok) {
      this.set({ phase: "failed", path, failure: file.failure });
      return;
    }
    await this.open(request, path, file.value, workspace, held?.selected ?? null);
  }

  private current(request: OperationId): boolean {
    return this.operations.isCurrent(request);
  }

  private loading(path: string): boolean {
    const state = this.snapshot();
    return state.phase === "loading" && state.path === path;
  }

  private held(path: string): Extract<PlantState, { phase: "ready" }> | null {
    const state = this.snapshot();
    return state.phase === "ready" && state.path === path ? state : null;
  }

  private async open(
    request: OperationId,
    path: string,
    file: FileContent,
    workspace: Workspace,
    selected: string | null,
  ): Promise<void> {
    const domain = parsed(file.content);
    const id = typeof domain === "object" && domain !== null && "id" in domain ? String(domain.id) : "";
    const composition = pairOf(id, workspace);
    const plant = parsePlant(domain, composition === null ? null : await this.composition(composition));
    const kept = plant.instances.some((instance) => instance.id === selected) ? selected : null;
    const specs = await this.specs(plant);
    if (!this.current(request)) return;
    this.set({ phase: "ready", path, etag: file.etag, plant, specs, composition, selected: kept });
  }

  async show(key: string, domain: unknown): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "idle" && state.phase !== "failed" && state.path === key) return;
    const request = this.operations.begin();
    this.set({ phase: "loading", path: key });
    const plant = parsePlant(domain);
    const specs = await this.specs(plant);
    if (!this.current(request)) return;
    this.set({ phase: "ready", path: key, etag: "", plant, specs, composition: null, selected: null });
  }

  close(): void {
    this.operations.invalidate();
    this.set({ phase: "idle" });
  }

  select(id: string | null): void {
    const state = this.snapshot();
    if (state.phase === "ready") {
      this.set({ ...state, selected: id });
    }
  }

  private async composition(path: string): Promise<unknown> {
    const file = await this.bridge.readFile(path);
    return file.ok ? parsed(file.value.content) : null;
  }

  private specs(plant: Plant): Promise<ReadonlyMap<string, ComponentSpec>> {
    return componentSpecs(
      this.bridge,
      plant.instances.map((instance) => instance.component),
    );
  }
}
