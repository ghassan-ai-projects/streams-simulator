import type { ComponentSpec } from "../../../model/plant";
import type { Outcome } from "../../ports";

interface ComponentGateway<F> {
  component(ref: string): Promise<Outcome<ComponentSpec, F>>;
}

const unavailable: ComponentSpec = { ports: [], category: "env" };

export async function componentSpecs<F>(
  gateway: ComponentGateway<F>,
  components: readonly (string | null)[],
): Promise<ReadonlyMap<string, ComponentSpec>> {
  const refs = [...new Set(components.filter((ref): ref is string => ref !== null))];
  const entries = await Promise.all(
    refs.map(async (ref) => {
      const shown = await gateway.component(ref);
      return [ref, shown.ok ? shown.value : unavailable] as const;
    }),
  );
  return new Map(entries);
}
