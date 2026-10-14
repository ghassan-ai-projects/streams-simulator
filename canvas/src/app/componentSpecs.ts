import type { ComponentSpec } from "../model/ports";
import { parseComponent } from "../model/ports";
import type { Bridge } from "./ports";

export async function componentSpecs(
  bridge: Bridge,
  components: readonly (string | null)[],
): Promise<ReadonlyMap<string, ComponentSpec>> {
  const refs = [...new Set(components.filter((ref): ref is string => ref !== null))];
  const entries = await Promise.all(
    refs.map(async (ref) => {
      const shown = await bridge.action("component.show", { ref });
      return [ref, parseComponent(shown.ok ? shown.value : null)] as const;
    }),
  );
  return new Map(entries);
}
