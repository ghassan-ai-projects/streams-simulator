import { instrumentOf, takesInstrument } from "../../../model/devices";
import type { Member } from "../../../model/devices";
import type { DevicesGateway } from "../../ports";
import type { Catalogue } from "./twinDraft";

function unitsOf(members: readonly Member[]): readonly string[] {
  const metrics = members.flatMap((member) => member.kind.metrics);
  return [...new Set(metrics.filter((metric) => metric.name !== "" && takesInstrument(metric)).map((m) => m.unit))];
}

async function listFor(gateway: DevicesGateway, unit: string) {
  const listed = await gateway.instruments(unit);
  return [unit, listed.ok ? listed.value.map(instrumentOf) : null] as const;
}

export async function loadCatalogue(
  gateway: DevicesGateway,
  members: readonly Member[],
  loaded: Catalogue,
): Promise<Catalogue> {
  const wanted = unitsOf(members).filter((unit) => loaded[unit] === undefined);
  if (wanted.length === 0) return loaded;
  const lists = await Promise.all(wanted.map((unit) => listFor(gateway, unit)));
  return { ...loaded, ...Object.fromEntries(lists) };
}
