import type { Composition } from "../model/composition";
import { domainOf, serialise } from "../model/composition";
import type { ExamplePlant } from "../model/library";
import type { PlantFailure, PlantGateway, PlantSource } from "./ports";

export type { PlantSource } from "./ports";

export interface Made {
  readonly composition: string;
  readonly domain: string;
}

const idPointer = "/id";

function taken(id: string): PlantFailure {
  return {
    kind: "conflict",
    diagnostic: { displayCode: "plant_exists", message: `A plant with the id ${id} already exists.` },
  };
}

export function idProblemOf(failure: PlantFailure): string {
  return issuesOf(failure).find((issue) => issue.pointer === idPointer)?.message ?? "";
}

function worded(failure: PlantFailure): PlantFailure {
  const issues = issuesOf(failure);
  if (issues.length === 0) return failure;
  const others = issues.filter((issue) => issue.pointer !== idPointer);
  return {
    ...failure,
    diagnostic: { ...failure.diagnostic, message: others.map((issue) => issue.message).join("; ") },
  };
}

function issuesOf(failure: PlantFailure) {
  return "issues" in failure ? (failure.issues ?? []) : [];
}

function asMade(id: string): Made {
  return { composition: `compositions/${id}.composition.json`, domain: domainOf({ id }) };
}

export class PlantMaker {
  constructor(private readonly gateway: PlantGateway) {}

  async examples(): Promise<readonly ExamplePlant[]> {
    const listed = await this.gateway.examples();
    return listed.ok ? listed.value : [];
  }

  async idProblem(source: PlantSource, id: string, title: string): Promise<string> {
    const existing = await this.gateway.readFile(domainOf({ id }));
    if (existing.ok) return taken(id).diagnostic.message;
    if (existing.failure.kind !== "missing") return existing.failure.diagnostic.message;
    const made = await this.generate(source, id, title, "copy");
    return made.ok ? "" : idProblemOf(made.failure);
  }

  async create(source: PlantSource, id: string, title: string, copy: boolean): Promise<PlantFailure | Made> {
    const result = await this.write(source, id, title, copy);
    if ("composition" in result && copy) await this.compose(result);
    return result;
  }

  async tryExample(example: ExamplePlant, id: string): Promise<PlantFailure | Made> {
    const result = await this.write({ kind: "example", id: example.id }, id, example.title, true);
    if (!("composition" in result)) return result;
    const composed = await this.compose(result);
    return composed.ok ? result : composed.failure;
  }

  private async write(source: PlantSource, id: string, title: string, copy: boolean): Promise<PlantFailure | Made> {
    const domain = domainOf({ id });
    const existing = await this.gateway.readFile(domain);
    if (existing.ok) return taken(id);
    if (existing.failure.kind !== "missing") return existing.failure;
    const generated = await this.generate(source, id, title, copy ? "copy" : "empty");
    if (!generated.ok) return worded(generated.failure);
    return this.writeComposition(id, generated.value);
  }

  private async generate(source: PlantSource, id: string, title: string, mode: "copy" | "empty") {
    return this.gateway.generateComposition({ source, id, title, mode });
  }

  private async writeComposition(id: string, composition: Composition): Promise<PlantFailure | Made> {
    const made = asMade(id);
    const written = await this.gateway.writeFile(made.composition, serialise(composition), "");
    if (written.ok) return made;
    return written.failure.kind === "conflict" ? taken(id) : written.failure;
  }

  private compose(made: Made) {
    return this.gateway.compose({ composition: made.composition, output: made.domain, etag: "" });
  }
}
