import { success } from "../app/ports";
import type { PerturbationKind } from "../model/launch";
import type { Canned } from "./fakeFailures";

interface Teachable {
  readonly actions: Map<string, Canned<unknown>>;
}

export const deliveryKinds: readonly PerturbationKind[] = [
  {
    kind: "drop",
    category: "delivery",
    description: "Loses records in transport; the consumer never receives them.",
    requiresEnd: false,
    params: [{ name: "rate", type: "number", minimum: 0, maximum: 1 }],
  },
  {
    kind: "producer_flap",
    category: "delivery",
    description: "Silences the producer during its window, then delivers what it held.",
    requiresEnd: true,
    params: [
      { name: "mode", type: "enum", values: ["hold", "lose"] },
      { name: "burst", type: "integer", minimum: 1, maximum: 9 },
    ],
  },
  {
    kind: "reorder",
    category: "delivery",
    description: "Delivers records out of order, displaced by a few positions.",
    requiresEnd: false,
    params: [],
  },
];

const wireOf = (kind: PerturbationKind) => ({
  kind: kind.kind,
  category: kind.category,
  description: kind.description,
  requires_end: kind.requiresEnd,
  params: kind.params,
});

export const miniCategories = {
  chiller: {
    chiller__fouling: "world",
    chiller__sensor_drift: "measurement",
    chiller__stuck_valve: "world_and_measurement",
  },
} as const;

export const perturbationsWire = { perturbations: deliveryKinds.map(wireOf) };

export const describedWire = { spec: {}, digest: "sha256:x", fault_categories: miniCategories };

export function teachCatalog(
  fake: Teachable,
  categories: object = miniCategories,
  kinds: readonly PerturbationKind[] = deliveryKinds,
): void {
  fake.actions.set("catalog.perturbations", success({ perturbations: kinds.map(wireOf) }));
  fake.actions.set("catalog.describe", success({ ...describedWire, fault_categories: categories }));
}
