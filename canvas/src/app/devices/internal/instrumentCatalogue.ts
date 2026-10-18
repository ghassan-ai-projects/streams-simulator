import { catalogueKey, instrumentOf, instrumentQuery, takesInstrument } from "../../../model/devices";
import type { InstrumentQuery, Member } from "../../../model/devices";
import type { DevicesGateway } from "../../ports";
import type { Catalogue } from "./twinDraft";

function queriesOf(members: readonly Member[]): ReadonlyMap<string, InstrumentQuery> {
  const metrics = members.flatMap((member) => member.kind.metrics);
  const asked = metrics.filter((metric) => metric.name !== "" && takesInstrument(metric));
  return new Map(asked.map((metric) => [catalogueKey(metric), instrumentQuery(metric)] as const));
}

async function listFor(gateway: DevicesGateway, key: string, query: InstrumentQuery) {
  const listed = await gateway.instruments(query);
  return [key, listed.ok ? listed.value.map(instrumentOf) : null] as const;
}

export async function loadCatalogue(
  gateway: DevicesGateway,
  members: readonly Member[],
  loaded: Catalogue,
): Promise<Catalogue> {
  const wanted = [...queriesOf(members)].filter(([key]) => loaded[key] === undefined);
  if (wanted.length === 0) return loaded;
  const lists = await Promise.all(wanted.map(([key, query]) => listFor(gateway, key, query)));
  return { ...loaded, ...Object.fromEntries(lists) };
}
