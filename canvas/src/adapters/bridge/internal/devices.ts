import type { BridgeAdapters, InstrumentQuery } from "../../../app/ports";
import { fileContent, written } from "./decoders";
import { deviceFailure } from "./failures";
import { boardCapture } from "./captureDecoder";
import type { Transport } from "./http";
import { action, getFile, putFile, request } from "./http";
import { kindChecked } from "./kindIssues";
import { lintAction } from "./lintAction";
import { deviceListing, domainChecked, instrumentRefs, suggestedChoices, twinMade } from "./twinDecoders";

function listArguments({ unit, min, max }: InstrumentQuery): Record<string, unknown> {
  if (unit === "") return {};
  return min !== null && max !== null ? { unit, min, max } : { unit };
}

function twinGateway(
  transport: Transport,
): Omit<BridgeAdapters["devices"], "readFile" | "writeFile" | "check" | "checkContent" | "captureKinds"> {
  return {
    listing: () => request(transport, "/api/workspace", deviceListing, deviceFailure("read")),
    instruments: (query) =>
      action(transport, "component.list", listArguments(query), instrumentRefs, deviceFailure("read")),
    suggest: (kinds) => action(transport, "kind.suggest", { kinds }, suggestedChoices, deviceFailure("check")),
    twin: ({ kinds, choices, out, etag }) =>
      action(transport, "kind.twin", { kinds, choices, out, etag }, twinMade, deviceFailure("twin")),
    validateDomain: (path) => action(transport, "domain.validate", { path }, domainChecked, deviceFailure("check")),
    lint: lintAction(transport, deviceFailure("read")),
  };
}

export function devicesGateway(transport: Transport): BridgeAdapters["devices"] {
  return {
    readFile: (path) => getFile(transport, path, fileContent, deviceFailure("read")),
    writeFile: (path, content, etag) => putFile(transport, path, content, etag, written, deviceFailure("write")),
    check: (path) => action(transport, "kind.validate", { path }, kindChecked, deviceFailure("check")),
    checkContent: (content, twin) =>
      action(transport, "kind.check", { content, twin }, kindChecked, deviceFailure("check")),
    captureKinds: (source) => action(transport, "kind.capture", { ...source }, boardCapture, deviceFailure("capture")),
    ...twinGateway(transport),
  };
}
