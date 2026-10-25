import { bridgeAdapter } from "../adapters/bridge";
import type { Outcome } from "../app/ports";
import { bridgeTransport } from "./bridgeTransport";
import { respond } from "./fakeWire";
import { assumptionsWire, diffWire, lineageWire } from "./provenanceFixture";

type Gateway = ReturnType<typeof bridgeAdapter>["provenance"];

async function decoded<T>(wire: unknown, read: (gateway: Gateway) => Promise<Outcome<T, unknown>>): Promise<T> {
  const transport = bridgeTransport(() => respond(200, { result: wire }));
  const outcome = await read(bridgeAdapter(transport).provenance);
  if (!outcome.ok) throw new Error("the provenance fixture does not decode");
  return outcome.value;
}

export const lineageModel = await decoded(lineageWire, (gateway) => gateway.lineage());
export const assumptionsModel = await decoded(assumptionsWire, (gateway) =>
  gateway.assumptions({ path: "domains/a.domain.json", twin: null }),
);
export const diffModel = await decoded(diffWire, (gateway) => gateway.diff("w-1", "domains/a.domain.json"));
