import type { BridgeAdapters } from "../../../app/ports";
import { assumptions } from "./assumptionsDecoder";
import { composeSummary, fileContent } from "./decoders";
import { modelDiff } from "./diffDecoder";
import { documentFailure } from "./failures";
import type { Transport } from "./http";
import { action, getFile } from "./http";
import { lineage } from "./lineageDecoder";

type Gateway = BridgeAdapters["provenance"];

async function outputEtag(transport: Transport, output: string): Promise<string> {
  const file = await getFile(transport, output, fileContent, documentFailure);
  return file.ok ? file.value.etag : "";
}

async function compose(transport: Transport, input: Parameters<Gateway["compose"]>[0]) {
  const etag = await outputEtag(transport, input.output);
  const args = { composition: input.composition, out: input.output, etag };
  return action(transport, "domain.compose", args, composeSummary, documentFailure);
}

export function provenanceGateway(transport: Transport): Gateway {
  return {
    lineage: () => action(transport, "domain.lineage", {}, lineage, documentFailure),
    assumptions: ({ path, twin }) =>
      action(transport, "domain.assumptions", twin === null ? { path } : { path, twin }, assumptions, documentFailure),
    diff: (a, b) => action(transport, "domain.diff", { a, b }, modelDiff, documentFailure),
    compose: (input) => compose(transport, input),
  };
}
