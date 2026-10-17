import type { Composition } from "../../../model/editor";
import { domainOf, serialise } from "../../../model/editor";
import type { ExamplePlant } from "../../../model/plant";
import type { Outcome, PlantFailure, PlantGateway, PlantSource } from "../../ports";
import { failure, success } from "../../ports";
import type { Made, MakeMode, PlantMaker as PlantMakerApi } from "./api";

type MakeOutcome = Outcome<Made, PlantFailure>;

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

  async create(source: PlantSource, id: string, title: string, mode: MakeMode): Promise<MakeOutcome> {
    const made = await this.write(source, id, title, mode);
    if (made.ok && mode === "copy") await this.compose(made.value);
    return made;
  }

  async tryExample(example: ExamplePlant, id: string): Promise<MakeOutcome> {
    const made = await this.write({ kind: "example", id: example.id }, id, example.title, "copy");
    if (!made.ok) return made;
    const composed = await this.compose(made.value);
    return composed.ok ? made : failure(composed.failure);
  }

  private async write(source: PlantSource, id: string, title: string, mode: MakeMode): Promise<MakeOutcome> {
    const existing = await this.gateway.readFile(domainOf({ id }));
    if (existing.ok) return failure(taken(id));
    if (existing.failure.kind !== "missing") return failure(existing.failure);
    const generated = await this.generate(source, id, title, mode);
    if (!generated.ok) return failure(worded(generated.failure));
    return this.writeComposition(id, generated.value);
  }

  private async generate(source: PlantSource, id: string, title: string, mode: MakeMode) {
    return this.gateway.generateComposition({ source, id, title, mode });
  }

  private async writeComposition(id: string, composition: Composition): Promise<MakeOutcome> {
    const made = asMade(id);
    const written = await this.gateway.writeFile(made.composition, serialise(composition), "");
    if (written.ok) return success(made);
    return failure(written.failure.kind === "conflict" ? taken(id) : written.failure);
  }

  private compose(made: Made) {
    return this.gateway.compose({ composition: made.composition, output: made.domain, etag: "" });
  }
}

export function createPlantMaker(gateway: PlantGateway): PlantMakerApi {
  return new PlantMaker(gateway);
}
