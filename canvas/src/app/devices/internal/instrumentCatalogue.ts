import { instrumentOf, sortInstruments } from "../../../model/devices";
import type { Instrument } from "../../../model/devices";
import type { DevicesGateway } from "../../ports";

export async function loadInstruments(gateway: DevicesGateway): Promise<readonly Instrument[] | null> {
  const listed = await gateway.instruments();
  if (!listed.ok) return null;
  const all = await Promise.all(
    listed.value.map(async (entry) => {
      const spec = await gateway.instrument(entry.ref);
      return instrumentOf(entry, spec.ok ? spec.value : null);
    }),
  );
  return sortInstruments(all);
}
