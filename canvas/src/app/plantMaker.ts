import type { BridgeFailure } from "../model/bridge";
import type { Composition } from "../model/composition";
import { domainOf, parseComposition, serialise } from "../model/composition";
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

function baseOf(source: PlantSource): Readonly<Record<string, string>> {
  return "example" in source ? { example: source.example } : { from: source.file };
}

function worded(failure: BridgeFailure): BridgeFailure {
  if (failure.issues.length === 0) return failure;
  return { ...failure, message: failure.issues.map((issue) => issue.message).join("; ") };
}

function compositionOf(value: unknown): Composition | null {
  const body = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  return parseComposition(JSON.stringify(body.composition ?? null));
}

export class PlantMaker {
  constructor(private readonly bridge: Bridge) {}

  async examples(): Promise<readonly ExamplePlant[]> {
    const listed = await this.bridge.action("component.examples", {});
    return listed.ok ? parseExamples(listed.value) : [];
  }

  async create(source: PlantSource, id: string, title: string, copy: boolean): Promise<BridgeFailure | Made> {
    const domain = domainOf({ id });
    if ((await this.bridge.readFile(domain)).ok) return taken(id);
    const made = await this.bridge.action("component.new", {
      id,
      title,
      ...baseOf(source),
      ...(copy ? {} : { empty: true }),
    });
    if (!made.ok) return worded(made.failure);
    const doc = compositionOf(made.value);
    if (doc === null) return unreadable;
    const composition = `compositions/${id}.composition.json`;
    const written = await this.bridge.writeFile(composition, serialise(doc), "");
    if (written.ok) return { composition, domain };
    return written.failure.code === "conflict" ? taken(id) : written.failure;
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
}
