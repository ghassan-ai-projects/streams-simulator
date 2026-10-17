import type { BridgeAdapters } from "../../../app/ports";
import { fileContent, workspace, written } from "./decoders";
import { deviceFailure } from "./failures";
import type { Transport } from "./http";
import { action, getFile, putFile, request } from "./http";
import { kindChecked } from "./kindIssues";
import { domainChecked, instrumentRefs, instrumentSpec, lintReport, twinMade } from "./twinDecoders";

function kindEntries(value: unknown) {
  return workspace(value)?.kinds ?? null;
}

function twinGateway(transport: Transport): Omit<BridgeAdapters["devices"], "readFile" | "writeFile" | "check"> {
  return {
    kinds: () => request(transport, "/api/workspace", kindEntries, deviceFailure("read")),
    instruments: () => action(transport, "component.list", {}, instrumentRefs, deviceFailure("read")),
    instrument: (ref) => action(transport, "component.show", { ref }, instrumentSpec, deviceFailure("read")),
    twin: ({ choices, out, etag }) =>
      action(transport, "kind.twin", { choices, out, etag }, twinMade, deviceFailure("twin")),
    validateDomain: (path) => action(transport, "domain.validate", { path }, domainChecked, deviceFailure("check")),
    lint: (path) =>
      action(transport, "domain.lint", { path }, lintReport, deviceFailure("read"), (problem) =>
        problem.code === "domain_lint" ? lintReport(problem.result) : null,
      ),
  };
}

export function devicesGateway(transport: Transport): BridgeAdapters["devices"] {
  return {
    readFile: (path) => getFile(transport, path, fileContent, deviceFailure("read")),
    writeFile: (path, content, etag) => putFile(transport, path, content, etag, written, deviceFailure("write")),
    check: (path) => action(transport, "kind.validate", { path }, kindChecked, deviceFailure("check")),
    ...twinGateway(transport),
  };
}
