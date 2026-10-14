import type { BridgeFailure, Outcome } from "../model/bridge";
import { failed, success } from "../model/bridge";
import type { Composition } from "../model/composition";
import { domainOf, parseComposition, serialise } from "../model/composition";
import { copyComposition, newComposition } from "../model/compositionEdits";
import type { ExamplePlant } from "../model/library";
import { parseExamples } from "../model/library";
import type { Bridge } from "./ports";

export type PlantSource = { readonly file: string } | { readonly example: string };

export interface Made {
  readonly composition: string;
  readonly domain: string;
}

const unreadable: BridgeFailure = {
  status: 0,
  code: "composition_unreadable",
  message: "The file is not a JSON composition.",
  issues: [],
};

function taken(id: string): BridgeFailure {
  return { status: 409, code: "plant_exists", message: `A plant with the id ${id} already exists.`, issues: [] };
}

export class PlantMaker {
  constructor(private readonly bridge: Bridge) {}

  async examples(): Promise<readonly ExamplePlant[]> {
    const listed = await this.bridge.action("component.examples", {});
    return listed.ok ? parseExamples(listed.value) : [];
  }

  async create(source: PlantSource, id: string, title: string, copy: boolean): Promise<BridgeFailure | Made> {
    const base = await this.read(source);
    if (!base.ok) return base.failure;
    const domain = domainOf({ id });
    const existing = await this.bridge.readFile(domain);
    if (existing.ok) return taken(id);
    const composition = `compositions/${id}.composition.json`;
    const doc = copy ? copyComposition(base.value, id, title) : newComposition(base.value, id, title);
    const written = await this.bridge.writeFile(composition, serialise(doc), "");
    return written.ok ? { composition, domain } : written.failure;
  }

  async tryExample(example: ExamplePlant, id: string): Promise<BridgeFailure | Made> {
    const made = await this.create({ example: example.id }, id, example.title, true);
    if (!("composition" in made)) return made;
    const composed = await this.bridge.action("domain.compose", {
      composition: made.composition,
      out: made.domain,
      etag: "",
    });
    return composed.ok ? made : composed.failure;
  }

  private async read(source: PlantSource): Promise<Outcome<Composition>> {
    if ("example" in source) return this.example(source.example);
    const file = await this.bridge.readFile(source.file);
    if (!file.ok) return file;
    const doc = parseComposition(file.value.content);
    return doc === null ? failed(unreadable) : success(doc);
  }

  private async example(id: string): Promise<Outcome<Composition>> {
    const shown = await this.bridge.action("component.example", { id });
    if (!shown.ok) return shown;
    const body =
      typeof shown.value === "object" && shown.value !== null ? (shown.value as Record<string, unknown>) : {};
    const doc = parseComposition(JSON.stringify(body.example ?? null));
    return doc === null ? failed(unreadable) : success(doc);
  }
}
