import type { BridgeFailure, Workspace } from "../model/bridge";
import type { Plant } from "../model/plant";
import { parsePlant } from "../model/plant";
import type { PortSpec } from "../model/ports";
import { parsePorts } from "../model/ports";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export type PlantState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: BridgeFailure }
  | {
      readonly phase: "ready";
      readonly path: string;
      readonly plant: Plant;
      readonly specs: ReadonlyMap<string, readonly PortSpec[]>;
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

function document(value: unknown): unknown {
  return typeof value === "object" && value !== null && "document" in value ? value.document : null;
}

export class PlantStore extends Observable<PlantState> {
  constructor(private readonly bridge: Bridge) {
    super({ phase: "idle" });
  }

  async load(path: string, workspace: Workspace): Promise<void> {
    this.set({ phase: "loading", path });
    const file = await this.bridge.readFile(path);
    if (!file.ok) {
      this.set({ phase: "failed", path, failure: file.failure });
      return;
    }
    const domain = parsed(file.value.content);
    const id = typeof domain === "object" && domain !== null && "id" in domain ? String(domain.id) : "";
    const composition = workspace.compositions.find((entry) => entry.id === id && id !== "")?.path ?? null;
    const plant = parsePlant(domain, composition === null ? null : await this.composition(composition));
    this.set({ phase: "ready", path, plant, specs: await this.specs(plant), composition, selected: null });
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

  private async specs(plant: Plant): Promise<ReadonlyMap<string, readonly PortSpec[]>> {
    const components = [
      ...new Set(plant.instances.map((instance) => instance.component).filter((ref): ref is string => ref !== null)),
    ];
    const entries = await Promise.all(
      components.map(async (ref) => {
        const shown = await this.bridge.action("component.show", { ref });
        return [ref, shown.ok ? parsePorts(document(shown.value)) : []] as const;
      }),
    );
    return new Map(entries);
  }
}
